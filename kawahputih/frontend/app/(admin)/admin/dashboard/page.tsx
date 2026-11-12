"use client";

import { useEffect, useState } from "react";
import { Ticket, ReceiptText, Wallet, Users } from "lucide-react";
import {
  LineChart,
  PageHeader,
  SectionCard,
  StatCard,
  StatGrid,
} from "@/components/panel/panel-ui";
import { analyticsService } from "@/services/analytics.api.service";
import { rupiah } from "@/lib/business";
import type {
  AdminDashboardStats,
  ManagerAnalytics,
  TrendPoint,
} from "@/types/api";

// Dashboard admin — data diambil dari backend API.

export default function AdminDashboardPage() {
  const [stats, setStats] = useState<AdminDashboardStats | null>(null);
  const [sales, setSales] = useState<TrendPoint[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDashboard();
  }, []);

  async function loadDashboard() {
    setLoading(true);
    try {
      const [dashboardStats, managerAnalytics] = await Promise.all([
        analyticsService.getAdminDashboard(),
        analyticsService.getManagerAnalytics({ range: "7d" }),
      ]);
      setStats(dashboardStats);
      // Ubah tren pendapatan manager analytics untuk chart
      setSales(
        managerAnalytics.revenue_trend.map((d) => ({
          label: d.label,
          value: d.value,
        })),
      );
    } catch (error: any) {
      console.error("Failed to load dashboard:", error);
    } finally {
      setLoading(false);
    }
  }

  if (loading) {
    return (
      <div>
        <PageHeader title="Dashboard" subtitle="Memuat data..." />
        <StatGrid cols={3}>
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <StatCard
              key={i}
              label="Loading..."
              value="—"
              icon={<Users className="h-4 w-4" />}
            />
          ))}
        </StatGrid>
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title="Dashboard"
        subtitle="Selamat datang, Admin. Berikut adalah ringkasan data hari ini."
      />

      <StatGrid cols={3}>
        <StatCard
          label="Total Booking"
          value={stats?.total_bookings?.toLocaleString("id-ID") ?? "—"}
          icon={<ReceiptText className="h-4 w-4" />}
        />
        <StatCard
          label="Total Tiket Terjual"
          value={stats?.total_tickets_sold?.toLocaleString("id-ID") ?? "—"}
          icon={<Ticket className="h-4 w-4" />}
        />
        <StatCard
          label="Total Pengunjung"
          value={stats?.total_visitors?.toLocaleString("id-ID") ?? "—"}
          icon={<Users className="h-4 w-4" />}
        />
        <StatCard
          label="Pendapatan"
          value={stats?.revenue ? rupiah(stats.revenue) : "—"}
          icon={<Wallet className="h-4 w-4" />}
        />
        <StatCard
          label="Booking Hari Ini"
          value={stats?.bookings_today?.toLocaleString("id-ID") ?? "—"}
          icon={<ReceiptText className="h-4 w-4" />}
        />
        <StatCard
          label="Pengunjung Hari Ini"
          value={stats?.visitors_today?.toLocaleString("id-ID") ?? "—"}
          delta="tiket untuk kunjungan hari ini"
          trend="up"
          icon={<Users className="h-4 w-4" />}
        />
      </StatGrid>

      <SectionCard
        title="Grafik Penjualan Tiket (7 Hari Terakhir)"
        className="mt-4"
        bodyClassName="px-4 pb-4 pt-3"
      >
        <LineChart data={sales} />
      </SectionCard>
    </div>
  );
}
