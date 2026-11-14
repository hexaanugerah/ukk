"use client";

import { useRouteGuard } from "@/hooks/use-route-guard";
import { useAuditFeed } from "@/hooks/use-audit";
import { PanelLayout, type PanelLink } from "@/components/layout/panel-layout";
import {
  Bell,
  ChartNoAxesColumnIncreasing,
  History,
  LayoutDashboard,
  ReceiptText,
  Settings,
  Ticket,
  Users,
  Wallet,
  Layers,
} from "lucide-react";

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const guard = useRouteGuard(["ADMIN"]);
  // Badge = jumlah entri audit nyata (feed aktivitas dari backend).
  const { data: auditFeed } = useAuditFeed({ limit: 1 });
  const activityCount = auditFeed?.meta?.total ?? 0;

  if (guard !== "allowed") return null;

  const LINKS: PanelLink[] = [
    { href: "/admin/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { href: "/admin/tickets", label: "Kelola Tiket", icon: Ticket },
    { href: "/admin/bookings", label: "Kelola Booking", icon: ReceiptText },
    { href: "/admin/payments", label: "Kelola Pembayaran", icon: Wallet },
    { href: "/admin/payment-destinations", label: "Tujuan Pembayaran", icon: Wallet },
    { href: "/admin/visitors", label: "Kelola Pengunjung", icon: Users },
    { href: "/admin/users", label: "Kelola Petugas", icon: Users },
    { href: "/admin/content", label: "Kelola Konten", icon: Layers },
    { href: "/admin/reports", label: "Laporan", icon: ChartNoAxesColumnIncreasing },
    { href: "/admin/notifications", label: "Aktivitas", icon: Bell, badge: activityCount },
    { href: "/admin/audit-log", label: "Audit Log", icon: History },
    { href: "/admin/settings", label: "Settings", icon: Settings },
  ];

  return (
    <PanelLayout role="Admin" links={LINKS}>
      {children}
    </PanelLayout>
  );
}
