"use client";

import { useEffect, useState } from "react";

export interface VitalsData {
  ping_ms: number;
  jitter_ms: number;
  packet_loss: number;
  primary_target: string;
  is_degraded: boolean;
  degraded_reason?: string;
}

export interface SpeedData {
  download_mbps: number;
  upload_mbps: number;
  ping_ms: number;
  jitter_ms: number;
  packet_loss: number;
  isp: string;
  is_degraded: boolean;
}

export interface ThreatScore {
  shannon_entropy: number;
  threat_score: number;
  is_threat: boolean;
}

export interface QueryEvent {
  timestamp: string;
  client_ip: string;
  device_name: string;
  domain: string;
  query_type: string;
  status: string;
  reason: string;
  threat?: ThreatScore;
  response_time_ms: number;
}

export function useSSE() {
  const [vitals, setVitals] = useState<VitalsData | null>(null);
  const [speedtest, setSpeedtest] = useState<SpeedData | null>(null);
  const [queries, setQueries] = useState<QueryEvent[]>([]);

  useEffect(() => {
    const es = new EventSource("/api/v1/stream");

    es.onmessage = (e) => {
      try {
        const payload = JSON.parse(e.data);
        if (payload.type === "vitals") {
          if (payload.data.vitals) setVitals(payload.data.vitals);
          if (payload.data.speedtest) setSpeedtest(payload.data.speedtest);
        } else if (payload.type === "query") {
          setQueries((prev) => [payload.data, ...prev.slice(0, 49)]);
        }
      } catch (err) {
        console.error("SSE parse error", err);
      }
    };

    return () => {
      es.close();
    };
  }, []);

  return { vitals, speedtest, queries };
}
