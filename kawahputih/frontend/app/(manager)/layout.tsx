"use client";

import { useRouteGuard } from "@/hooks/use-route-guard";
import { PanelLayout, type PanelLink } from "@/components/layout/panel-layout";
import { ChartNoAxesColumnIncreasing, LayoutDashboard, MonitorCheck, ReceiptText, Settings, Users, Wallet } from "lucide-react";

// Dashboard | Statistik(Penjualan) | Laporan(Pendapatan, Booking,
// Statistik Pengunjung) | Monitoring(Petugas)
const LINKS: PanelLink[] = [
  { href: "/manager/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/manager/sales", label: "Penjualan", icon: ChartNoAxesColumnIncreasing, section: "Statistik" },
  { href: "/manager/revenue", label: "Pendapatan", icon: Wallet, section: "Laporan" },
  { href: "/manager/bookings", label: "Booking", icon: ReceiptText, section: "Laporan" },
  { href: "/manager/visitors", label: "Statistik Pengunjung", icon: Users, section: "Laporan" },
  { href: "/manager/monitoring", label: "Petugas", icon: MonitorCheck, section: "Monitoring" },
  { href: "/manager/settings", label: "Keamanan Akun", icon: Settings },
];

export default function ManagerLayout({ children }: { children: React.ReactNode }) {
  const guard = useRouteGuard(["MANAGER"]);

  if (guard !== "allowed") return null;

  return (
    <PanelLayout role="Manager" links={LINKS}>
      {children}
    </PanelLayout>
  );
}
