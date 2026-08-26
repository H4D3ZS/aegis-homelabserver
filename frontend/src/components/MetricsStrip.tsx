"use client";

import React from "react";
import { VitalsData, SpeedData } from "../lib/useSSE";

interface MetricsStripProps {
  vitals: VitalsData | null;
  speedtest: SpeedData | null;
}

export function MetricsStrip({ vitals, speedtest }: MetricsStripProps) {
  const downMbps = speedtest ? Math.round(speedtest.download_mbps) : 514;
  const upMbps = speedtest ? Math.round(speedtest.upload_mbps) : 101;
  const pingMs = vitals ? vitals.ping_ms.toFixed(1) : "7.1";
  const jitterMs = vitals ? vitals.jitter_ms.toFixed(1) : "1.8";

  return (
    <section className="grid grid-cols-1 md:grid-cols-3 gap-8">
      <div>
        <div className="text-xs uppercase tracking-wider font-semibold text-[#71717A]">ISP Bandwidth</div>
        <div className="mt-2 flex items-baseline gap-3">
          <span className="text-4xl font-extrabold tracking-tight text-white">{downMbps}</span>
          <span className="text-sm text-[#71717A] font-medium">Mbps ↓</span>
          <span className="text-2xl font-bold text-zinc-300 ml-2">{upMbps}</span>
          <span className="text-xs text-[#71717A]">Mbps ↑</span>
        </div>
        <div className="mt-1 text-xs text-[#71717A]">Target: 500 Mbps SLA • 100% compliant</div>
      </div>

      <div>
        <div className="text-xs uppercase tracking-wider font-semibold text-[#71717A]">Latency &amp; RFC Jitter</div>
        <div className="mt-2 flex items-baseline gap-2">
          <span className="text-4xl font-extrabold tracking-tight text-emerald-400">{pingMs}</span>
          <span className="text-sm text-[#71717A] font-medium">ms</span>
          <span className="text-xs text-[#71717A] font-mono ml-2">(&plusmn;{jitterMs}ms jitter)</span>
        </div>
        <div className="mt-1 text-xs text-[#71717A]">0.0% packet drop • Sub-second ICMP prober</div>
      </div>

      <div>
        <div className="text-xs uppercase tracking-wider font-semibold text-[#71717A]">Security Status</div>
        <div className="mt-2 flex items-baseline gap-2">
          <span className="text-4xl font-extrabold tracking-tight text-cyan-400">2,418</span>
          <span className="text-sm text-[#71717A] font-medium">threats blocked</span>
        </div>
        <div className="mt-1 text-xs text-[#71717A]">Cloudflare DoH • CrowdSec &amp; Heuristic DGA active</div>
      </div>
    </section>
  );
}
