"use client";

import { useEffect, useState } from "react";
import { Download } from "lucide-react";
import { toast } from "sonner";
import { LineChart, ListPanel, PageHeader, SectionCard, StatCard, StatGrid } from "@/components/panel/panel-ui";
import { DateRangeFilter } from "@/components/panel/date-range-filter";
import { analyticsService } from "@/services/analytics.api.service";
import type { ManagerAnalytics } from "@/types/api";
import type { DateRangePreset } from "@/types/domain";
import { exportReportPDF, exportReportExcel, type ReportData } from "@/lib/report-export";
import { rupiah } from "@/lib/business";

// Laporan pendapatan — hanya baca, dihitung dari backend API + rentang tanggal.

export default function ManagerRevenuePage() {
  const [analytics, setAnalytics] = useState<ManagerAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [range, setRange] = useState<DateRangePreset>("30d");
  const [custom, setCustom] = useState<{ from: string; to: string } | undefined>();

  useEffect(() => {
    let alive = true;
    setLoading(true);
    (async () => {
      try {
        const a = await analyticsService.getManagerAnalytics({ range, custom });
        if (alive) setAnalytics(a);
      } catch (error: any) {
        toast.error("Gagal memuat analytics", { description: error.message });
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => {
      alive = false;
    };
  }, [range, custom]);

  function buildReport(): ReportData {
    if (!analytics) return {} as ReportData;
    return {
      title: "Laporan Pendapatan",
      dateRangeLabel: analytics.range_label,
      summary: [
        { label: "Total Pendapatan", value: rupiah(analytics.summary.total_revenue) },
        { label: "Pendapatan Hari Ini", value: rupiah(analytics.summary.revenue_today) },
        { label: "Total Booking", value: analytics.summary.total_bookings.toLocaleString("id-ID") },
        { label: "Rata-rata Nilai Booking", value: rupiah(analytics.summary.avg_booking_value) },
        { label: "Perubahan vs Periode Lalu", value: `${analytics.changes.revenue >= 0 ? "+" : ""}${analytics.changes.revenue}%` },
      ],
      revenueSeries: analytics.revenue_trend,
      bookingSeries: analytics.booking_trend,
      ticketPerformance: analytics.ticket_type_performance,
      paymentStatus: analytics.payment_status,
      dailyRows: analytics.daily_rows.map((r) => ({
        date: r.date,
        bookings: r.bookings,
        tickets: r.tickets,
        revenue: r.revenue,
        visitors: r.visitors,
        paymentStatus: r.payment_status,
      })),
    };
  }

  if (loading) {
    return (
      <div>
        <PageHeader
          title="Laporan Pendapatan"
          subtitle="Memuat data..."
        />
        <StatGrid cols={3}>
          <StatCard label="Loading..." value="—" />
          <StatCard label="Loading..." value="—" />
          <StatCard label="Loading..." value="—" />
        </StatGrid>
      </div>
    );
  }

  if (!analytics) {
    return (
      <div>
        <PageHeader title="Laporan Pendapatan" subtitle="Data tidak tersedia" />
      </div>
    );
  }

  const a = analytics;

  return (
    <div>
      <PageHeader
        title="Laporan Pendapatan"
        subtitle="Ringkasan Pendapatan"
        action={
          <div className="flex flex-wrap items-center gap-2">
            <DateRangeFilter
              value={range}
              onChange={(p, c) => {
                setRange(p);
                setCustom(c);
              }}
            />
            <button
              onClick={() => {
                exportReportPDF(buildReport());
                toast.success("Laporan PDF diunduh");
              }}
              className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f]"
            >
              <Download className="h-4 w-4" /> PDF
            </button>
            <button
              onClick={() => {
                exportReportExcel(buildReport());
                toast.success("Laporan Excel diunduh");
              }}
              className="inline-flex h-9 items-center gap-1.5 rounded-md border border-[#d7e3ed] bg-white px-3 text-xs font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
            >
              <Download className="h-4 w-4" /> Excel
            </button>
          </div>
        }
      />
      <StatGrid cols={3}>
        <StatCard label="Total Pendapatan" value={rupiah(a.summary.total_revenue)} delta={`${a.changes.revenue >= 0 ? "+" : ""}${a.changes.revenue}% vs periode lalu`} trend={a.changes.revenue >= 0 ? "up" : "down"} />
        <StatCard label="Pendapatan Hari Ini" value={rupiah(a.summary.revenue_today)} />
        <StatCard label="Rata-rata / Booking" value={rupiah(a.summary.avg_booking_value)} />
      </StatGrid>

      <div className="mt-3 grid gap-3 lg:grid-cols-3">
        <SectionCard title="Tren Pendapatan" subtitle={`Periode: ${a.range_label}`} className="lg:col-span-2">
          <LineChart data={a.revenue_trend.map((p) => ({ label: p.label, value: Math.round(p.value) }))} />
        </SectionCard>

        <SectionCard title="Pendapatan per Jenis Tiket">
          <ListPanel
            items={a.ticket_type_performance.map((t) => ({ name: t.label, value: `${t.value} tiket` }))}
          />
        </SectionCard>
      </div>
    </div>
  );
}