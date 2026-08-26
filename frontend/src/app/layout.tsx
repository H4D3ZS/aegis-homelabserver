import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Aegis Sentinel | Sovereign Homelab",
  description: "Native Bare-Metal ISP & Threat Sentinel",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="bg-[#09090B] text-[#FAFAFA] min-h-screen antialiased">
        {children}
      </body>
    </html>
  );
}
