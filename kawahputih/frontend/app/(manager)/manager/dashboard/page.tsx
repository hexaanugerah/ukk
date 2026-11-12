"use client";

import { useEffect, useState } from "react";
import { LineChart, PageHeader, ProgressList, SectionCard, StatCard, StatGrid, StatusPill } from "@/components/panel/panel-ui";
import { analyticsService } from "@/services/analytics.api.service";
import { rupiah } from "@/lib/business";
import type { ManagerAnalytics, TrendPoint, DailyRow } from "@/types/api";

// Dashboard manager — hanya baca. Semua metrik dihitung dari backend API.

function ChangeBadge({ pct }: { pct: number }) {
  const up = pct >= 0;
  return (
    <p className={`mt-0.5 text-[10px] ${up ? "text-[#16876a]" : "text-red-600"}`}>
      {up ? "▲" : "▼"} {Math.abs(pct)}% dari periode sebelumnya
    </p>
  );
}

export default function ManagerDashboardPage() {
  const [analytics, setAnalytics] = useState<ManagerAnalytics | null>(null);
  const [monthAnalytics, setMonthAnalytics] = useState<ManagerAnalytics | null>(null);  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadAnalytics();
  }, []);

  async function loadAnalytics() {
    setLoading(true);
    try {
      const [a, month] = await Promise.all([
        analyticsService.getManagerAnalytics({ range: "7d" }),
        analyticsService.getManagerAnalytics({ range: "month" }),
      ]);
      setAnalytics(a);
      setMonthAnalytics(month);
      
    } catch (error) {
      console.error("Failed to load analytics:", error);
    } finally {
      setLoading(false);
    }
  }

  if (loading) {
    return (
      <div>
        <PageHeader title="Dashboard" subtitle="Memuat data..." />
        <StatGrid cols={4}>
          {[1, 2, 3, 4].map((i) => (
            <StatCard key={i} label="Loading..." value="—" />
          ))}
        </StatGrid>
      </div>
    );
  }

  const a = analytics;
  const month = monthAnalytics;

  if (!a || !month) {
    return (
      <div>
        <PageHeader title="Dashboard" subtitle="Data tidak tersedia" />
      </div>
    );
  }

  const trend = a.revenue_trend.map((p: TrendPoint) => ({ label: p.label, value: Math.round(p.value) }));
  const bestSelling = a.ticket_type_performance.slice(0, 5).map((t: TrendPoint) => ({ name: t.label, value: t.value }));

  return (
    <div>
      <PageHeader title="Dashboard" subtitle="Ringkasan Aktivitas & Kinerja" />

      <StatGrid cols={4}>
        <StatCard
          label="Total Pengunjung"
          value={month.summary.total_visitors.toLocaleString("id-ID")}
          delta="dari periode sebelumnya"
          trend="up"
        />
        <StatCard
          label="Tiket Terjual"
          value={a.summary.total_tickets.toLocaleString("id-ID")}
          delta="dari periode sebelumnya"
          trend="up"
        />
        <StatCard
          label="Total Booking"
          value={month.summary.total_bookings.toLocaleString("id-ID")}
          delta="dari periode sebelumnya"
          trend="up"
        />
        <StatCard
          label="Total Pendapatan"
          value={rupiah(month.summary.total_revenue)}
          delta="dari periode sebelumnya"
          trend="up"
        />
      </StatGrid>

      <div className="mt-2.5 grid gap-2.5 lg:grid-cols-3">
        <SectionCard title="Tren Pendapatan (7 Hari Terakhir)" className="lg:col-span-2">
          <LineChart data={trend} />
        </SectionCard>

        <div className="space-y-2.5">
          <div className="rounded-md border border-[#1768ad] bg-[#1768ad] p-3 text-white">
            <p className="text-[10px] text-blue-100">Pendapatan Hari Ini</p>
            <p className="mt-1 text-xl font-bold tabular-nums">{rupiah(a.summary.revenue_today)}</p>
            <ChangeBadge pct={a.changes.revenue} />
          </div>
          <SectionCard title="Tiket Terlaris (7 hari)">
            <ProgressList data={bestSelling} />
          </SectionCard>
        </div>
      </div>


    </div>
  );
}