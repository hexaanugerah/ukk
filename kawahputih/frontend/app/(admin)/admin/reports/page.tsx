"use client";

import { useEffect, useMemo, useState } from "react";
import { Download, RefreshCw } from "lucide-react";
import { PageHeader, SectionCard, StatCard, StatGrid, StatusPill } from "@/components/panel/panel-ui";
import { SearchInput } from "@/components/panel/panel-toolbar";
import { analyticsService } from "@/services/analytics.api.service";
import { bookingService } from "@/services/booking.api.service";
import { paymentService } from "@/services/payment.api.service";
import { visitorService } from "@/services/visitor.api.service";
import { rupiah } from "@/lib/business";
import type {
  AdminDashboardStats,
  Booking,
  BookingStatus,
  ManagerAnalytics,
  Payment,
  PaymentStatus,
  Visitor,
} from "@/types/api";

// Laporan admin — menggabungkan data booking, pembayaran, dan pengunjung
// dari backend. Semua angka dihitung ulang dari respons API, bukan mock.

const BOOKING_ID: Record<BookingStatus, string> = {
  PENDING: "Menunggu",
  PAID: "Dibayar",
  CONFIRMED: "Terkonfirmasi",
  USED: "Digunakan",
  CANCELLED: "Dibatalkan",
  EXPIRED: "Kadaluarsa",
};

const PAYMENT_ID: Record<PaymentStatus, string> = {
  PENDING: "Menunggu",
  PROCESSING: "Diproses",
  PAID: "Berhasil",
  FAILED: "Gagal",
  EXPIRED: "Kadaluarsa",
  REFUNDED: "Refund",
};

interface ReportRow {
  date: string;
  bookings: number;
  tickets: number;
  revenue: number;
  visitors: number;
}

export default function AdminReportsPage() {
  const [stats, setStats] = useState<AdminDashboardStats | null>(null);
  const [analytics, setAnalytics] = useState<ManagerAnalytics | null>(null);
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [payments, setPayments] = useState<Payment[]>([]);
  const [visitors, setVisitors] = useState<Visitor[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");

  async function load() {
    setLoading(true);
    setError(null);
    try {
      const [dashboard, manager, bookingRows, paymentRows, visitorRows] = await Promise.all([
        analyticsService.getAdminDashboard(),
        analyticsService.getManagerAnalytics({ range: "30d" }),
        bookingService.list({ page: 1, limit: 100 }),
        paymentService.list({ page: 1, limit: 100 }),
        visitorService.list(),
      ]);
      setStats(dashboard);
      setAnalytics(manager);
      setBookings(bookingRows);
      setPayments(paymentRows);
      setVisitors(visitorRows);
    } catch (e: any) {
      setError(e?.message || "Gagal memuat laporan");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const filtered = useMemo(() => {
    const needle = q.trim().toLowerCase();
    return bookings.filter((b) => {
      if (status && b.status !== status) return false;
      if (!needle) return true;
      return [b.booking_code, b.visitor_name, b.email].some((v) =>
        (v ?? "").toLowerCase().includes(needle)
      );
    });
  }, [bookings, q, status]);

  const daily = useMemo(() => {
    const map = new Map<string, ReportRow>();
    for (const b of bookings) {
      if (!["PAID", "CONFIRMED", "USED"].includes(b.status)) continue;
      const row = map.get(b.visit_date) ?? {
        date: b.visit_date,
        bookings: 0,
        tickets: 0,
        revenue: 0,
        visitors: 0,
      };
      row.bookings += 1;
      row.tickets += b.total_tickets;
      row.revenue += b.total;
      row.visitors += b.total_tickets;
      map.set(b.visit_date, row);
    }
    return [...map.values()].sort((a, b) => b.date.localeCompare(a.date));
  }, [bookings]);

  const paymentSummary = useMemo(() => {
    const out: Record<string, number> = {};
    for (const p of payments) out[p.status] = (out[p.status] ?? 0) + 1;
    return out;
  }, [payments]);

  function exportCSV() {
    const header = ["Kode Booking", "Nama", "Email", "Tanggal Kunjungan", "Tiket", "Total", "Status Booking", "Status Pembayaran"];
    const lines = filtered.map((b) => [
      b.booking_code,
      b.visitor_name,
      b.email,
      b.visit_date,
      String(b.total_tickets),
      String(b.total),
      b.status,
      b.payment?.status ?? "-",
    ]);
    const csv = [header, ...lines]
      .map((row) => row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(","))
      .join("\n");
    const blob = new Blob(["\uFEFF" + csv], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `laporan-booking-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  if (loading) {
    return (
      <div>
        <PageHeader title="Laporan" subtitle="Memuat data laporan..." />
        <StatGrid cols={4}>
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-20 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />
          ))}
        </StatGrid>
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title="Laporan"
        subtitle="Penjualan, pembayaran, booking, dan pengunjung — 30 hari terakhir"
        action={
          <div className="flex items-center gap-2">
            <button
              onClick={load}
              className="inline-flex h-9 items-center gap-1.5 rounded-md border border-[#d7e3ed] bg-white px-3 text-xs font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
            >
              <RefreshCw className="h-4 w-4" /> Muat ulang
            </button>
            <button
              onClick={exportCSV}
              className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f]"
            >
              <Download className="h-4 w-4" /> Ekspor CSV
            </button>
          </div>
        }
      />

      {error && (
        <p className="mb-3 rounded-md bg-[#fdeaea] px-3 py-2 text-sm text-[#c53030]">{error}</p>
      )}

      <StatGrid cols={4}>
        <StatCard
          label="Pendapatan (30 hari)"
          value={rupiah(analytics?.summary.total_revenue ?? 0)}
        />
        <StatCard
          label="Booking (30 hari)"
          value={(analytics?.summary.total_bookings ?? 0).toLocaleString("id-ID")}
        />
        <StatCard
          label="Tiket Terjual (30 hari)"
          value={(analytics?.summary.total_tickets ?? 0).toLocaleString("id-ID")}
        />
        <StatCard
          label="Pengunjung Terdaftar"
          value={(stats?.total_visitors ?? visitors.length).toLocaleString("id-ID")}
        />
      </StatGrid>

      <div className="mt-3 grid gap-3 lg:grid-cols-2">
        <SectionCard title="Ringkasan Pembayaran" subtitle={`${payments.length} transaksi terbaca`}>
          {payments.length === 0 ? (
            <p className="py-4 text-center text-sm text-[#94A3B8]">Belum ada transaksi pembayaran.</p>
          ) : (
            <ul className="space-y-2 text-sm">
              {Object.entries(PAYMENT_ID).map(([key, label]) => (
                <li key={key} className="flex items-center justify-between rounded-md border border-[#e5edf3] px-3 py-2">
                  <StatusPill status={label} />
                  <span className="font-semibold tabular-nums text-[#17324d]">
                    {(paymentSummary[key] ?? 0).toLocaleString("id-ID")}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </SectionCard>

        <SectionCard title="Rekap Kunjungan Harian" subtitle="Hanya booking berbayar">
          {daily.length === 0 ? (
            <p className="py-4 text-center text-sm text-[#94A3B8]">Belum ada data kunjungan.</p>
          ) : (
            <div className="panel-scroll max-h-64 rounded-md border border-[#e5edf3]">
              <table className="panel-table">
                <thead>
                  <tr>
                    <th>Tanggal</th>
                    <th>Booking</th>
                    <th>Tiket</th>
                    <th>Pendapatan</th>
                  </tr>
                </thead>
                <tbody>
                  {daily.map((r) => (
                    <tr key={r.date}>
                      <td className="tabular-nums">{r.date}</td>
                      <td className="tabular-nums">{r.bookings}</td>
                      <td className="tabular-nums">{r.tickets}</td>
                      <td className="tabular-nums">{rupiah(r.revenue)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </SectionCard>
      </div>

      <SectionCard
        className="mt-3"
        title="Laporan Booking & Pengunjung"
        subtitle={`${filtered.length} booking ditampilkan`}
      >
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <SearchInput
            value={q}
            onChange={setQ}
            placeholder="Cari kode booking / nama / email…"
            className="w-full sm:max-w-xs"
          />
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter status booking"
          >
            <option value="">Semua status</option>
            {Object.entries(BOOKING_ID).map(([key, label]) => (
              <option key={key} value={key}>
                {label}
              </option>
            ))}
          </select>
        </div>

        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>Kode</th>
                <th>Pengunjung</th>
                <th>Kunjungan</th>
                <th>Tiket</th>
                <th>Total</th>
                <th>Booking</th>
                <th>Pembayaran</th>
              </tr>
            </thead>
            <tbody>
              {filtered.length === 0 ? (
                <tr>
                  <td colSpan={7} className="py-10 text-center text-sm text-[#94A3B8]">
                    Tidak ada data sesuai filter.
                  </td>
                </tr>
              ) : (
                filtered.map((b) => (
                  <tr key={b.id}>
                    <td className="font-medium text-[#17324d]">{b.booking_code}</td>
                    <td>
                      <p className="font-medium text-[#17324d]">{b.visitor_name}</p>
                      <p className="text-[11px] text-[#698097]">{b.email}</p>
                    </td>
                    <td className="tabular-nums">{b.visit_date}</td>
                    <td className="tabular-nums">{b.total_tickets}</td>
                    <td className="tabular-nums">{rupiah(b.total)}</td>
                    <td>
                      <StatusPill status={BOOKING_ID[b.status] ?? b.status} />
                    </td>
                    <td>
                      <StatusPill
                        status={b.payment ? PAYMENT_ID[b.payment.status] ?? b.payment.status : "Belum Bayar"}
                      />
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </SectionCard>
    </div>
  );
}
