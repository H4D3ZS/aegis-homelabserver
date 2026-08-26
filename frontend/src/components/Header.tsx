"use client";

import React from "react";

interface HeaderProps {
  onRunSpeedtest: () => void;
  onOpenUnbreak: () => void;
  onOpenReboot: () => void;
  isSpeedtesting: boolean;
}

export function Header({ onRunSpeedtest, onOpenUnbreak, onOpenReboot, isSpeedtesting }: HeaderProps) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-6 pb-6 border-b border-[#27272A]">
      <div>
        <div className="flex items-center gap-3">
          <div className="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse" />
          <h1 className="text-2xl font-bold tracking-tight text-white">Aegis Sentinel</h1>
          <span className="text-xs text-[#71717A] font-mono">x86_64 // N4100 Bare-Metal</span>
        </div>
        <p className="text-xs text-[#71717A] mt-1">
          Converge FiberX 500M • Primary Router: <span className="font-mono text-zinc-300">192.168.100.1</span> (Huawei EG8041X6-10)
        </p>
      </div>

      <div className="flex items-center gap-2">
        <button
          onClick={onRunSpeedtest}
          disabled={isSpeedtesting}
          className="px-3.5 py-1.5 rounded bg-[#18181B] hover:bg-[#27272A] text-white text-xs font-medium transition cursor-pointer disabled:opacity-50"
        >
          {isSpeedtesting ? "Testing..." : "Run Speedtest"}
        </button>
        <button
          onClick={onOpenUnbreak}
          className="px-3.5 py-1.5 rounded bg-[#18181B] hover:bg-[#27272A] text-white text-xs font-medium transition cursor-pointer"
        >
          Fix Video Stream
        </button>
        <button
          onClick={onOpenReboot}
          className="px-3.5 py-1.5 rounded bg-[#18181B] hover:bg-[#27272A] text-zinc-400 hover:text-rose-400 text-xs font-medium transition cursor-pointer"
        >
          Reboot Fiber Router
        </button>
      </div>
    </header>
  );
}
