"use client";

import { useEffect, useState } from "react";
import { Download } from "lucide-react";
import { toast } from "sonner";
import { BarChart, PageHeader, PieChart, ProgressList, SectionCard } from "@/components/panel/panel-ui";
import { DateRangeFilter } from "@/components/panel/date-range-filter";
import { analyticsService } from "@/services/analytics.api.service";
import type { ManagerAnalytics } from "@/types/api";
import type { DateRangePreset } from "@/types/domain";
import { exportReportPDF, exportReportExcel, type ReportData } from "@/lib/report-export";
import { rupiah } from "@/lib/business";

// Statistik penjualan — hanya baca, dihitung dari backend API sesuai rentang tanggal yang dipilih. Bisa diekspor ke PDF & Excel.

export default function ManagerSalesPage() {
  const [analytics, setAnalytics] = useState<ManagerAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [range, setRange] = useState<DateRangePreset>("7d");
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

  const ticketSales = analytics?.ticket_type_performance.map((t) => ({ name: t.label, value: t.value })) ?? [];

  function buildReport(): ReportData {
    if (!analytics) return {} as ReportData;
    const a = analytics;
    return {
      title: "Laporan Statistik Penjualan",
      dateRangeLabel: a.range_label,
      summary: [
        { label: "Total Booking", value: a.summary.total_bookings.toLocaleString("id-ID") },
        { label: "Tiket Terjual", value: a.summary.total_tickets.toLocaleString("id-ID") },
        { label: "Total Pendapatan", value: rupiah(a.summary.total_revenue) },
        { label: "Rata-rata Nilai Booking", value: rupiah(a.summary.avg_booking_value) },
        { label: "Conversion Rate", value: `${a.summary.conversion_rate}%` },
      ],
      revenueSeries: a.revenue_trend,
      bookingSeries: a.booking_trend,
      ticketPerformance: a.ticket_type_performance,
      paymentStatus: a.payment_status,
      dailyRows: a.daily_rows.map((r) => ({
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
          title="Statistik Penjualan"
          subtitle="Memuat data..."
        />
        <div className="grid gap-4 lg:grid-cols-3">
          <SectionCard title="Loading..." className="lg:col-span-2">
            <div className="animate-pulse h-64" />
          </SectionCard>
          <SectionCard title="Loading...">
            <div className="animate-pulse h-32" />
          </SectionCard>
          <SectionCard title="Loading...">
            <div className="animate-pulse h-32" />
          </SectionCard>
        </div>
      </div>
    );
  }

  if (!analytics) {
    return (
      <div>
        <PageHeader title="Statistik Penjualan" subtitle="Data tidak tersedia" />
      </div>
    );
  }

  const a = analytics;

  return (
    <div>
      <PageHeader
        title="Statistik Penjualan"
        subtitle="Ringkasan Penjualan Tiket"
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
      <div className="grid gap-4 lg:grid-cols-3">
        <SectionCard title="Penjualan Tiket" subtitle={`Periode: ${a.range_label}`} className="lg:col-span-2">
          <PieChart data={a.ticket_type_performance.length ? a.ticket_type_performance : a.booking_trend} />
        </SectionCard>

        <SectionCard title="Total Penjualan">
          <div className="grid h-full place-items-center text-center">
            <p className="text-[40px] font-extrabold text-[#1768ad]">{a.summary.total_tickets.toLocaleString("id-ID")}</p>
            <p className="text-sm font-semibold text-[#698097]">tiket</p>
          </div>
        </SectionCard>

        <SectionCard title="Penjualan per Jenis Tiket">
          <ProgressList data={ticketSales} />
        </SectionCard>

        <SectionCard title="Weekday vs Weekend">
          <BarChart data={a.weekday_vs_weekend} color="#1768ad" height={170} />
        </SectionCard>

        <SectionCard title="Detail Penjualan">
          <div className="mt-3 flex items-center justify-between rounded-md bg-[#eef5fb] px-4 py-3">
            <span className="text-sm font-semibold text-[#29445e]">Total Pendapatan</span>
            <span className="text-lg font-bold tabular-nums text-[#1768ad]">{rupiah(a.summary.total_revenue)}</span>
          </div>
        </SectionCard>
      </div>
    </div>
  );
}