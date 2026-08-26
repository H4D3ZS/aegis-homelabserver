package router

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// RouterConfig contains gateway connection credentials and parameters.
type RouterConfig struct {
	GatewayURL     string
	Username       string
	Password       string
	RouterType     string
	AutoHeal       bool
	HealTimeoutSec int
}

// RouterStatus contains operational status of the fiber gateway.
type RouterStatus struct {
	IsConnected      bool      `json:"is_connected"`
	WANIP            string    `json:"wan_ip"`
	GatewayIP        string    `json:"gateway_ip"`
	ConnectionUptime string    `json:"connection_uptime"`
	SignalQuality    string    `json:"signal_quality"`
	DHCPClientsCount int       `json:"dhcp_clients_count"`
	LastRebootTime   time.Time `json:"last_reboot_time"`
	AutoHealActive   bool      `json:"auto_heal_active"`
	RouterModel      string    `json:"router_model"`
}

// HuaweiDriver implements the 3-step token challenge handshake for Huawei OptiXstar ONT.
type HuaweiDriver struct {
	client     *http.Client
	gatewayURL string
	username   string
	password   string
}

// NewHuaweiDriver creates a Huawei ONT automation driver with cookie jar.
func NewHuaweiDriver(gatewayURL, username, password string) *HuaweiDriver {
	jar, _ := cookiejar.New(nil)
	return &HuaweiDriver{
		client: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
		},
		gatewayURL: strings.TrimRight(gatewayURL, "/"),
		username:   username,
		password:   password,
	}
}

// Authenticate performs the ASP/CGI nonce challenge login flow.
func (d *HuaweiDriver) Authenticate(ctx context.Context) error {
	// Step 1: Hit /asp/GetRandCount.asp to obtain challenge token
	tokenURL := fmt.Sprintf("%s/asp/GetRandCount.asp", d.gatewayURL)
	req1, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenURL, nil)
	if err != nil {
		return err
	}
	resp1, err := d.client.Do(req1)
	if err != nil {
		return fmt.Errorf("step 1 GetRandCount failed: %w", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	_ = resp1.Body.Close()

	reToken := regexp.MustCompile(`x\.X_HW_Token\s*=\s*"([^"]+)"`)
	matches := reToken.FindStringSubmatch(string(body1))
	token := ""
	if len(matches) > 1 {
		token = matches[1]
	}

	// Step 2: POST credentials with Cookie: Cookie=body:Language:english:id=-1
	loginURL := fmt.Sprintf("%s/login.cgi", d.gatewayURL)
	form := url.Values{}
	form.Set("UserName", d.username)
	form.Set("PassWord", base64.StdEncoding.EncodeToString([]byte(d.password)))
	if token != "" {
		form.Set("x.X_HW_Token", token)
	}

	req2, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Cookie", "Cookie=body:Language:english:id=-1")

	resp2, err := d.client.Do(req2)
	if err != nil {
		return fmt.Errorf("step 2 login.cgi failed: %w", err)
	}
	_ = resp2.Body.Close()

	log.Printf("[ROUTER-HUAWEI] Authenticated with ONT at %s", d.gatewayURL)
	return nil
}

// Reboot extracts onttoken from reset.asp and triggers hardware restart.
func (d *HuaweiDriver) Reboot(ctx context.Context) error {
	if err := d.Authenticate(ctx); err != nil {
		return fmt.Errorf("reboot auth failed: %w", err)
	}

	// Step 3: GET /html/ssmp/reset/reset.asp to scrape onttoken
	resetPageURL := fmt.Sprintf("%s/html/ssmp/reset/reset.asp", d.gatewayURL)
	req3, err := http.NewRequestWithContext(ctx, http.MethodGet, resetPageURL, nil)
	if err != nil {
		return err
	}
	resp3, err := d.client.Do(req3)
	if err != nil {
		return fmt.Errorf("step 3 reset.asp failed: %w", err)
	}
	body3, _ := io.ReadAll(resp3.Body)
	_ = resp3.Body.Close()

	reOntToken := regexp.MustCompile(`onttoken\s*=\s*"([^"]+)"`)
	m := reOntToken.FindStringSubmatch(string(body3))
	ontToken := ""
	if len(m) > 1 {
		ontToken = m[1]
	}

	// Step 4: POST /html/ssmp/reset/set.cgi?x=InternetGatewayDevice.X_HW_DEBUG.SMP.DM.ResetBoard
	actionURL := fmt.Sprintf("%s/html/ssmp/reset/set.cgi?x=InternetGatewayDevice.X_HW_DEBUG.SMP.DM.ResetBoard&RequestFile=html/ssmp/reset/reset.asp", d.gatewayURL)
	form := url.Values{}
	if ontToken != "" {
		form.Set("onttoken", ontToken)
	}

	req4, err := http.NewRequestWithContext(ctx, http.MethodPost, actionURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req4.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req4.Header.Set("Referer", resetPageURL)

	resp4, err := d.client.Do(req4)
	if err != nil {
		return fmt.Errorf("step 4 ResetBoard failed: %w", err)
	}
	_ = resp4.Body.Close()

	log.Printf("[ROUTER-HUAWEI] Programmatic ONT reboot command dispatched successfully!")
	return nil
}
