"use client";

import React from "react";
import { QueryEvent } from "../lib/useSSE";

interface ThreatStreamProps {
  queries: QueryEvent[];
}

export function ThreatStream({ queries }: ThreatStreamProps) {
  const handleAction = async (action: "whitelist" | "block", domain: string) => {
    await fetch(`/api/v1/${action}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ domain }),
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold tracking-tight text-zinc-200">Threat &amp; Activity Feed</h2>
        <span className="text-xs text-[#71717A] font-mono">Live Tail</span>
      </div>

      <div className="overflow-y-auto max-h-[480px] text-xs border border-[#18181B] rounded-lg">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-[#121215] border-b border-[#27272A] text-zinc-400 uppercase text-[10px] tracking-wider">
              <th className="py-2.5 px-3">Time</th>
              <th className="py-2.5 px-3">Domain</th>
              <th className="py-2.5 px-3 text-right">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[#18181B]">
            {queries.length === 0 ? (
              <tr>
                <td colSpan={3} className="py-8 text-center text-[#71717A]">
                  Listening for live network queries...
                </td>
              </tr>
            ) : (
              queries.map((q, idx) => {
                const isThreat = q.threat && (q.threat.is_threat || q.threat.threat_score >= 0.75);
                const timeStr = new Date(q.timestamp).toLocaleTimeString([], {
                  hour: "2-digit",
                  minute: "2-digit",
                  second: "2-digit",
                });

                return (
                  <tr
                    key={idx}
                    className={
                      isThreat ? "bg-rose-950/20 text-rose-300 transition" : "hover:bg-[#121215] text-zinc-400 transition"
                    }
                  >
                    <td className="py-2.5 px-3 text-[#71717A] font-mono text-[11px] whitespace-nowrap">{timeStr}</td>
                    <td className="py-2.5 px-3 pr-2 truncate max-w-[240px]">
                      {isThreat ? (
                        <>
                          <span className="font-mono text-rose-300 font-semibold">{q.domain}</span>
                          <span className="px-1.5 py-0.5 rounded bg-rose-900/60 text-rose-300 text-[10px] font-bold ml-1">
                            BLOCKED THREAT
                          </span>
                        </>
                      ) : (
                        <span className="font-mono text-zinc-300 font-normal">{q.domain}</span>
                      )}
                    </td>
                    <td className="py-2.5 px-3 text-right whitespace-nowrap">
                      <button
                        onClick={() => handleAction("whitelist", q.domain)}
                        className="px-2 py-0.5 text-zinc-400 hover:text-white mr-1 text-[11px]"
                      >
                        Allow
                      </button>
                      <button
                        onClick={() => handleAction("block", q.domain)}
                        className="px-2 py-0.5 text-zinc-400 hover:text-rose-400 text-[11px]"
                      >
                        Block
                      </button>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
