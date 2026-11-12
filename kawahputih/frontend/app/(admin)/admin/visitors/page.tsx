"use client";

import { useEffect, useMemo, useState } from "react";
import { Eye, History } from "lucide-react";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { SearchInput } from "@/components/panel/panel-toolbar";
import { visitorService } from "@/services/visitor.api.service";
import { bookingService } from "@/services/booking.api.service";
import { rupiah, paginate, searchIn, sortData } from "@/lib/business";
import type { Visitor, Booking, BookingStatus } from "@/types/api";

type VisitorRow = Visitor;

export default function AdminVisitorsPage() {
  const [visitors, setVisitors] = useState<VisitorRow[]>([]);
  const [loading, setLoading] = useState(true);

  const [q, setQ] = useState("");
  const [page, setPage] = useState(1);
  const [detail, setDetail] = useState<VisitorRow | null>(null);
  const [visitorBookings, setVisitorBookings] = useState<Booking[]>([]);
  const [bookingLoading, setBookingLoading] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  // Riwayat booking diambil dari backend saat detail pengunjung dibuka,
  // lengkap dengan snapshot pembayaran terakhir.
  useEffect(() => {
    if (!detail) {
      setVisitorBookings([]);
      return;
    }
    let alive = true;
    setBookingLoading(true);
    bookingService
      .listForVisitor(detail.id, { limit: 100 })
      .then((rows) => {
        if (alive) setVisitorBookings(rows);
      })
      .catch(() => {
        if (alive) setVisitorBookings([]);
      })
      .finally(() => {
        if (alive) setBookingLoading(false);
      });
    return () => {
      alive = false;
    };
  }, [detail]);

  async function loadData() {
    setLoading(true);
    try {
      const visitorsData = await visitorService.list();
      setVisitors(visitorsData);
    } catch (error: any) {
      console.error("Failed to load visitors:", error);
    } finally {
      setLoading(false);
    }
  }

  const rows = useMemo(() => {
    let list = searchIn(visitors, q, ["name", "email", "phone"]);
    list = sortData(list, "joined_at", "desc");
    return list;
  }, [visitors, q]);

  const paged = paginate(rows, page, 8);

  if (loading) {
    return (
      <div>
        <PageHeader title="Kelola Pengunjung" subtitle="Data pengunjung beserta riwayat kunjungan" />
        <SectionCard title="Daftar Pengunjung">
          <div className="animate-pulse space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-14 bg-gray-200 rounded border border-[#e5edf3]" />
            ))}
          </div>
        </SectionCard>
      </div>
    );
  }

  return (
    <div>
      <PageHeader title="Kelola Pengunjung" subtitle="Data pengunjung beserta riwayat kunjungan" />

      <SectionCard title="Daftar Pengunjung" action={<span className="text-sm font-semibold text-[#1768ad]">{rows.length} data</span>}>
        <div className="mb-3">
          <SearchInput value={q} onChange={(v) => { setQ(v); setPage(1); }} placeholder="Cari nama / email…" className="w-full sm:max-w-xs" />
        </div>

        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>NO</th>
                <th>Nama</th>
                <th>Email</th>
                <th>Tanggal Daftar</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {paged.rows.length === 0 ? (
                <tr><td colSpan={5} className="py-10 text-center text-sm text-[#94A3B8]">Belum ada data pengunjung.</td></tr>
              ) : (
                paged.rows.map((v, index) => (
                  <tr key={v.id}>
                    <td className="tabular-nums text-[#698097]">{(paged.page - 1) * 8 + index + 1}.</td>
                    <td className="font-medium text-[#17324d]">{v.name}</td>
                    <td>{v.email}</td>
                    <td className="tabular-nums">{v.joined_at}</td>
                    <td>
                      <button
                        onClick={() => setDetail(v)}
                        className="inline-flex h-7 items-center gap-1 rounded-md border border-[#d7e3ed] px-2 text-xs font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
                      >
                        <History className="h-3.5 w-3.5" /> Riwayat
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {paged.totalPages > 1 && (
          <div className="mt-3 flex items-center justify-end gap-1 text-xs text-[#698097]">
            <button onClick={() => setPage(page - 1)} disabled={page <= 1} className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40">‹</button>
            <span className="px-2">Hal {paged.page} / {paged.totalPages}</span>
            <button onClick={() => setPage(page + 1)} disabled={page >= paged.totalPages} className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40">›</button>
          </div>
        )}
      </SectionCard>

      {detail && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" onClick={() => setDetail(null)} role="dialog" aria-modal="true">
          <div className="max-h-[85vh] w-full max-w-2xl overflow-y-auto rounded-md bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between border-b border-[#e5edf3] pb-3">
              <div>
                <h3 className="text-base font-bold text-[#17324d]">{detail.name}</h3>
                <p className="text-xs text-[#698097]">{detail.email} • {detail.phone}</p>
              </div>
              <button onClick={() => setDetail(null)} className="text-[#698097] hover:text-[#17324d]" aria-label="Tutup">✕</button>
            </div>

            <div className="mt-3 grid grid-cols-3 gap-2 text-center">
              <div className="rounded-md bg-[#f7fbff] p-3">
                <p className="text-xs text-[#698097]">Total Booking</p>
                <p className="text-lg font-bold tabular-nums">{detail.total_bookings ?? 0}</p>
              </div>
              <div className="rounded-md bg-[#f7fbff] p-3">
                <p className="text-xs text-[#698097]">Total Belanja</p>
                <p className="text-lg font-bold tabular-nums">{rupiah(detail.total_spent ?? 0)}</p>
              </div>
              <div className="rounded-md bg-[#f7fbff] p-3">
                <p className="text-xs text-[#698097]">Terakhir Berkunjung</p>
                <p className="text-sm font-bold">{detail.last_visit ?? "-"}</p>
              </div>
            </div>

            <p className="mt-4 text-xs font-bold text-[#29445e]">RIWAYAT BOOKING</p>
            <div className="mt-2 space-y-2">
              {bookingLoading ? (
                <p className="py-4 text-center text-sm text-[#94A3B8]">Memuat riwayat booking...</p>
              ) : visitorBookings.length === 0 ? (
                <p className="py-4 text-center text-sm text-[#94A3B8]">Belum ada booking.</p>
              ) : (
                visitorBookings.map((b: Booking) => {
                  const statusLabel: Record<BookingStatus, string> = {
                    CONFIRMED: "Terkonfirmasi",
                    USED: "Selesai",
                    PENDING: "Menunggu",
                    CANCELLED: "Dibatalkan",
                    EXPIRED: "Kadaluarsa",
                    PAID: "Berhasil",
                  };
                  return (
                    <div key={b.id} className="flex items-center justify-between rounded-md border border-[#e5edf3] px-3 py-2 text-sm">
                      <div>
                        <p className="font-medium text-[#17324d]">{b.booking_code}</p>
                        <p className="text-xs text-[#698097]">Kunjungan {b.visit_date} • {b.total_tickets} tiket</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="tabular-nums text-sm">{rupiah(b.total)}</span>
                        <StatusPill status={statusLabel[b.status] ?? "Berhasil"} />
                        <StatusPill status={b.payment ? b.payment.status : "Belum Bayar"} />
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="shrink-0 text-[#698097]">{label}</dt>
      <dd className="text-right font-medium text-[#17324d]">{value}</dd>
    </div>
  );
}