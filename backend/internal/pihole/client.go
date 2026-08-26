package pihole

import (
	"context"
	"log"
	"os/exec"
)

// Client wraps Pi-hole CLI commands.
type Client struct {
	ExecutablePath string
}

// NewClient initializes a new Pi-hole client.
func NewClient(executablePath string) *Client {
	if executablePath == "" {
		executablePath = "/usr/local/bin/pihole"
	}
	return &Client{ExecutablePath: executablePath}
}

// WhitelistDomain adds a domain to Pi-hole's exact whitelist.
func (c *Client) WhitelistDomain(ctx context.Context, domain string) error {
	log.Printf("[PIHOLE] Whitelisting domain: %s", domain)
	if path, err := exec.LookPath(c.ExecutablePath); err == nil {
		cmd := exec.CommandContext(ctx, path, "-w", domain)
		return cmd.Run()
	}
	return nil
}

// RemoveWhitelist removes a domain from Pi-hole's exact whitelist.
func (c *Client) RemoveWhitelist(ctx context.Context, domain string) error {
	log.Printf("[PIHOLE] Removing domain from whitelist: %s", domain)
	if path, err := exec.LookPath(c.ExecutablePath); err == nil {
		cmd := exec.CommandContext(ctx, path, "-w", "-d", domain)
		return cmd.Run()
	}
	return nil
}

// BlacklistDomain adds a domain to Pi-hole's exact blacklist.
func (c *Client) BlacklistDomain(ctx context.Context, domain string) error {
	log.Printf("[PIHOLE] Blacklisting domain: %s", domain)
	if path, err := exec.LookPath(c.ExecutablePath); err == nil {
		cmd := exec.CommandContext(ctx, path, "-b", domain)
		return cmd.Run()
	}
	return nil
}

// ReloadLists triggers Pi-hole FTL reload without full restart.
func (c *Client) ReloadLists(ctx context.Context) error {
	if path, err := exec.LookPath(c.ExecutablePath); err == nil {
		cmd := exec.CommandContext(ctx, path, "restartdns", "reload-lists")
		return cmd.Run()
	}
	return nil
}
