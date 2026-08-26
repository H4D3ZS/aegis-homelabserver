"use client";

import React, { useState } from "react";

interface ModalsProps {
  unbreakOpen: boolean;
  onCloseUnbreak: () => void;
  rebootOpen: boolean;
  onCloseReboot: () => void;
}

export function Modals({ unbreakOpen, onCloseUnbreak, rebootOpen, onCloseReboot }: ModalsProps) {
  const [unbreakResult, setUnbreakResult] = useState<string | null>(null);
  const [isUnbreaking, setIsUnbreaking] = useState(false);

  const handleExecuteUnbreak = async () => {
    setIsUnbreaking(true);
    try {
      const res = await fetch("/api/v1/unbreak", { method: "POST" });
      const json = await res.json();
      setUnbreakResult(json.message || "Unblocked streaming CDNs for 15 minutes.");
    } catch {
      setUnbreakResult("Failed to trigger unbreak.");
    }
    setIsUnbreaking(false);
  };

  const handleExecuteReboot = async () => {
    try {
      const res = await fetch("/api/v1/router/reboot", { method: "POST" });
      const json = await res.json();
      alert(json.message || "Reboot command dispatched to Huawei ONT.");
    } catch {
      alert("Reboot dispatched.");
    }
    onCloseReboot();
  };

  return (
    <>
      {/* Unbreak Modal */}
      {unbreakOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80">
          <div className="w-full max-w-md bg-[#121215] p-6 rounded-lg border border-[#27272A] space-y-4 text-xs">
            <div className="flex justify-between items-start">
              <div>
                <h3 className="text-base font-bold text-white">1-Click Smart Unbreak</h3>
                <p className="text-xs text-[#71717A] mt-0.5">Temporarily unblock media streaming CDNs (15-min auto-eviction)</p>
              </div>
              <button onClick={onCloseUnbreak} className="text-zinc-500 hover:text-white text-lg">
                &times;
              </button>
            </div>
            <p className="text-xs text-[#71717A]">
              Scans recent blocked stream CDNs (<code>kwik.cx</code>, <code>doodstream</code>, <code>mp4upload</code>) and permits playback without whitelisting ad trackers.
            </p>
            <button
              onClick={handleExecuteUnbreak}
              disabled={isUnbreaking}
              className="w-full py-2.5 rounded bg-white text-black font-semibold text-xs hover:bg-zinc-200 transition cursor-pointer disabled:opacity-50"
            >
              {isUnbreaking ? "Scanning & Whitelisting..." : "Unblock Streams (Last 120s)"}
            </button>
            {unbreakResult && <div className="text-xs text-emerald-400 pt-2">{unbreakResult}</div>}
          </div>
        </div>
      )}

      {/* Reboot Confirm Modal */}
      {rebootOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80">
          <div className="w-full max-w-md bg-[#121215] p-6 rounded-lg border border-[#27272A] space-y-4 text-xs">
            <div className="flex justify-between items-start">
              <h3 className="text-base font-bold text-white">Reboot Fiber Router?</h3>
              <button onClick={onCloseReboot} className="text-zinc-500 hover:text-white text-lg">
                &times;
              </button>
            </div>
            <p className="text-xs text-[#71717A]">
              This will execute a clean hardware restart of your Huawei OptiXstar ONT (<span className="font-mono text-zinc-300">192.168.100.1</span>). WAN connection will drop for ~90 seconds.
            </p>
            <div className="flex items-center gap-2 pt-2">
              <button
                onClick={onCloseReboot}
                className="flex-1 py-2 rounded bg-[#18181B] hover:bg-[#27272A] text-zinc-300 text-xs font-medium"
              >
                Cancel
              </button>
              <button
                onClick={handleExecuteReboot}
                className="flex-1 py-2 rounded bg-rose-600 hover:bg-rose-700 text-white text-xs font-semibold"
              >
                Confirm Reboot
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
