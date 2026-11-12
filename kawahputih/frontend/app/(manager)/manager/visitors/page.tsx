"use client";

import { useEffect, useState } from "react";
import { BarChart, ListPanel, PageHeader, SectionCard } from "@/components/panel/panel-ui";
import { DateRangeFilter } from "@/components/panel/date-range-filter";
import { analyticsService } from "@/services/analytics.api.service";
import type { DateRangePreset } from "@/types/domain";
import type { ManagerAnalytics } from "@/types/api";

// Statistik pengunjung — hanya baca, dari backend API analytics.

export default function ManagerVisitorsPage() {
  const [range, setRange] = useState<DateRangePreset>("7d");
  const [custom, setCustom] = useState<{ from: string; to: string } | undefined>();
  const [analytics, setAnalytics] = useState<ManagerAnalytics | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    analyticsService
      .getManagerAnalytics(range === "custom" ? { range: "custom", custom } : { range })
      .then((a) => {
        if (alive) setAnalytics(a);
      })
      .catch((e) => console.error("Failed to load analytics:", e))
      .finally(() => {
        if (alive) setLoading(false);
      });
    return () => {
      alive = false;
    };
  }, [range, custom]);

  if (loading || !analytics) {
    return (
      <div>
        <PageHeader title="Statistik Pengunjung" subtitle="Memuat data..." />
        <SectionCard>
          <div className="animate-pulse space-y-4">
            <div className="h-40 rounded-md bg-gray-200" />
            <div className="h-40 rounded-md bg-gray-200" />
          </div>
        </SectionCard>
      </div>
    );
  }

  const a = analytics;

  return (
    <div>
      <PageHeader
        title="Statistik Pengunjung"
        subtitle="Ringkasan Pengunjung"
        action={
          <DateRangeFilter
            value={range}
            onChange={(p, c) => {
              setRange(p);
              setCustom(c);
            }}
          />
        }
      />

      <SectionCard title="Tren Pengunjung" subtitle={`Periode: ${a.range_label}`}>
        <BarChart data={a.booking_trend.map((p, i) => ({ label: p.label, value: p.value }))} color="#16876a" />
      </SectionCard>

      <div className="mt-3 grid gap-3 lg:grid-cols-3">
        <SectionCard title="Weekday vs Weekend" className="lg:col-span-2">
          <BarChart data={a.weekday_vs_weekend} />
        </SectionCard>

        <div className="rounded-md border border-[#1768ad] bg-[#1768ad] p-4 text-white">
          <p className="text-xs text-blue-100">Total Pengunjung</p>
          <p className="mt-2 text-3xl font-bold tabular-nums">{a.summary.total_visitors.toLocaleString("id-ID")}</p>
          <p className={`mt-1 text-xs ${a.changes.visitors >= 0 ? "text-blue-100" : "text-red-200"}`}>
            {a.changes.visitors >= 0 ? "▲" : "▼"} {Math.abs(a.changes.visitors)}% dari periode sebelumnya
          </p>
        </div>
      </div>

      <div className="mt-3">
        <SectionCard title="Kunjungan per Jenis Tiket">
          <ListPanel items={a.ticket_type_performance.map((t) => ({ name: t.label, value: `${t.value} tiket` }))} />
        </SectionCard>
      </div>
    </div>
  );
}
