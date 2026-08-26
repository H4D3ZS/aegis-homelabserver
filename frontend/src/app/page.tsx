"use client";

import React, { useState } from "react";
import { Header } from "../components/Header";
import { MetricsStrip } from "../components/MetricsStrip";
import { BandwidthChart } from "../components/BandwidthChart";
import { ThreatStream } from "../components/ThreatStream";
import { Modals } from "../components/Modals";
import { useSSE } from "../lib/useSSE";

export default function DashboardPage() {
  const { vitals, speedtest, queries } = useSSE();
  const [unbreakOpen, setUnbreakOpen] = useState(false);
  const [rebootOpen, setRebootOpen] = useState(false);
  const [isSpeedtesting, setIsSpeedtesting] = useState(false);

  const handleRunSpeedtest = async () => {
    setIsSpeedtesting(true);
    try {
      await fetch("/api/v1/speedtest/run", { method: "POST" });
      setTimeout(() => setIsSpeedtesting(false), 4000);
    } catch {
      setIsSpeedtesting(false);
    }
  };

  return (
    <main className="p-6 lg:p-10 max-w-[1600px] mx-auto space-y-10">
      <Header
        onRunSpeedtest={handleRunSpeedtest}
        onOpenUnbreak={() => setUnbreakOpen(true)}
        onOpenReboot={() => setRebootOpen(true)}
        isSpeedtesting={isSpeedtesting}
      />

      <MetricsStrip vitals={vitals} speedtest={speedtest} />

      <section className="grid grid-cols-1 xl:grid-cols-12 gap-10 pt-4 border-t border-[#18181B]">
        <div className="xl:col-span-7">
          <BandwidthChart />
        </div>
        <div className="xl:col-span-5">
          <ThreatStream queries={queries} />
        </div>
      </section>

      <Modals
        unbreakOpen={unbreakOpen}
        onCloseUnbreak={() => setUnbreakOpen(false)}
        rebootOpen={rebootOpen}
        onCloseReboot={() => setRebootOpen(false)}
      />
    </main>
  );
}
