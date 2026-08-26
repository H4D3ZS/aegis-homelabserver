"use client";

import React, { useEffect, useState } from "react";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from "chart.js";
import { Line } from "react-chartjs-2";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler);

export function BandwidthChart() {
  const [chartData, setChartData] = useState<any>({
    labels: ["12:00", "14:00", "16:00", "18:00", "20:00", "22:00", "00:00", "02:00", "04:00", "06:00", "08:00", "10:00"],
    datasets: [
      {
        label: "Download (Mbps)",
        data: [512, 510, 514, 508, 520, 515, 510, 518, 514, 512, 508, 514],
        borderColor: "#FAFAFA",
        borderWidth: 1.8,
        tension: 0.2,
        pointRadius: 0,
      },
      {
        label: "Upload (Mbps)",
        data: [100, 98, 101, 99, 102, 100, 98, 101, 100, 99, 98, 101],
        borderColor: "#71717A",
        borderWidth: 1.2,
        tension: 0.2,
        pointRadius: 0,
      },
    ],
  });

  useEffect(() => {
    fetch("/api/v1/history/speed?hours=24")
      .then((res) => res.json())
      .then((data) => {
        if (data.records && data.records.length > 0) {
          const labels = data.records.map((r: any) =>
            new Date(r.timestamp).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
          );
          const down = data.records.map((r: any) => r.download_mbps);
          const up = data.records.map((r: any) => r.upload_mbps);
          setChartData({
            labels,
            datasets: [
              {
                label: "Download (Mbps)",
                data: down,
                borderColor: "#FAFAFA",
                borderWidth: 1.8,
                tension: 0.2,
                pointRadius: 0,
              },
              {
                label: "Upload (Mbps)",
                data: up,
                borderColor: "#71717A",
                borderWidth: 1.2,
                tension: 0.2,
                pointRadius: 0,
              },
            ],
          });
        }
      })
      .catch(() => {});
  }, []);

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false } },
    scales: {
      x: {
        grid: { color: "#18181B" },
        ticks: { color: "#71717A", font: { family: "monospace", size: 10 } },
      },
      y: {
        grid: { color: "#18181B" },
        ticks: { color: "#71717A", font: { family: "monospace", size: 10 } },
        beginAtZero: true,
      },
    },
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold tracking-tight text-zinc-200">Bandwidth History</h2>
        <span className="text-xs text-[#71717A] font-mono">Past 24 Hours</span>
      </div>
      <div className="h-64 w-full">
        <Line data={chartData} options={options} />
      </div>
    </div>
  );
}
