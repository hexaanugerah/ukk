"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Eye, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { SearchInput } from "@/components/panel/panel-toolbar";
import { bookingService, BookingError } from "@/services/booking.api.service";
import { rupiah, paginate, searchIn, sortData } from "@/lib/business";
import type { Booking, BookingStatus } from "@/types/api";

const STATUS_OPTIONS: BookingStatus[] = ["PENDING", "PAID", "CONFIRMED", "USED", "CANCELLED", "EXPIRED"];

const STATUS_ID: Record<BookingStatus, string> = {
  PENDING: "Menunggu",
  PAID: "Berhasil (Dibayar)",
  CONFIRMED: "Terkonfirmasi",
  USED: "Selesai (Digunakan)",
  CANCELLED: "Dibatalkan",
  EXPIRED: "Kadaluarsa",
};

type SortKey = "created_at" | "total" | "visit_date";

export default function AdminBookingsPage() {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);

  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");
  const [date, setDate] = useState("");
  const [sortKey, setSortKey] = useState<SortKey>("created_at");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(8);
  const [detail, setDetail] = useState<Booking | null>(null);
  const [delTarget, setDelTarget] = useState<Booking | null>(null);
  const [deleting, setDeleting] = useState(false);

  const loadBookings = useCallback(async () => {
    setLoading(true);
    try {
      const data = await bookingService.list({ page, limit: pageSize, status: status || undefined });
      setBookings(data);
    } catch (error: any) {
      toast.error("Gagal memuat booking", { description: error.message });
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, status]);

  // Muat booking saat mount dan ketika page/size/status berubah. Filter tanggal diterapkan di sisi klien (lihat `rows`), jadi tidak boleh memicu refetch.
  useEffect(() => {
    loadBookings();
  }, [loadBookings]);

  async function changeStatus(b: Booking, next: BookingStatus) {
    try {
      await bookingService.updateStatus(b.id, next);
      toast.success(`Status ${b.booking_code} → ${STATUS_ID[next]}`);
      loadBookings(); // Reload to get updated data
    } catch (error: any) {
      toast.error("Gagal mengubah status", { description: error.message });
    }
  }

  async function handleDelete() {
    if (!delTarget) return;
    setDeleting(true);
    try {
      // Backend: DELETE /bookings/:id (admin diizinkan) → status jadi DIBATALKAN.
      await bookingService.cancel(delTarget.id);
      toast.success(`${delTarget.id} dibatalkan`);
      setDelTarget(null);
      loadBookings();
    } catch (error: any) {
      toast.error("Gagal menghapus booking", { description: error.message });
    } finally {
      setDeleting(false);
    }
  }

  function toggleSort(key: SortKey) {
    if (sortKey === key) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortKey(key);
      setSortDir("desc");
    }
  }

  // Perhitungan yang di-memoize (harus sebelum early return)
  const rows = useMemo(() => {
    let list = searchIn(bookings, q, ["booking_code", "visitor_name", "email"]);
    if (status) list = list.filter((b) => b.status === status);
    if (date) list = list.filter((b) => b.visit_date === date);
    list = sortData(list, sortKey, sortDir);
    return list;
  }, [bookings, q, status, date, sortKey, sortDir]);

  const paged = paginate(rows, page, pageSize);

  const summary = useMemo(() => {
    const count = (s: BookingStatus) => bookings.filter((b) => b.status === s).length;
    return [
      { label: "Total Booking", value: bookings.length },
      { label: "Berhasil", value: count("PAID") + count("CONFIRMED") },
      { label: "Menunggu Pembayaran", value: count("PENDING") },
      { label: "Dibatalkan", value: count("CANCELLED") },
      { label: "Selesai (Digunakan)", value: count("USED") },
    ];
  }, [bookings]);

  if (loading) {
    return (
      <div>
        <PageHeader title="Kelola Booking" subtitle="Pantau dan kelola booking" />
        <SectionCard title="Semua Booking">
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
      <PageHeader title="Kelola Booking" subtitle="Pantau dan kelola booking" />

      <div className="mb-3 grid grid-cols-2 gap-3 lg:grid-cols-5">
        {summary.map((s) => (
          <div key={s.label} className="rounded-md border border-[#d7e3ed] bg-white px-4 py-3">
            <p className="text-xs font-semibold text-[#698097]">{s.label}</p>
            <p className="mt-1 text-xl font-bold tabular-nums text-[#17324d]">{s.value.toLocaleString("id-ID")}</p>
          </div>
        ))}
      </div>

      <SectionCard title="Semua Booking" action={<span className="text-sm font-semibold text-[#1768ad]">{rows.length} data</span>}>
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <SearchInput value={q} onChange={(v) => { setQ(v); setPage(1); }} placeholder="Cari kode booking / nama…" className="w-full sm:max-w-xs" />
          <select
            value={status}
            onChange={(e) => { setStatus(e.target.value); setPage(1); }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter status"
          >
            <option value="">Semua status</option>
            {STATUS_OPTIONS.map((s) => (
              <option key={s} value={s}>{STATUS_ID[s]}</option>
            ))}
          </select>
          <input
            type="date"
            value={date}
            onChange={(e) => { setDate(e.target.value); setPage(1); }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter tanggal kunjungan"
          />
        </div>

        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>NO</th>
                <th>Kode Booking</th>
                <th>Nama Pemesan</th>
                <th>
                  <button onClick={() => toggleSort("visit_date")} className="font-bold uppercase">
                    Tanggal {sortKey === "visit_date" && (sortDir === "asc" ? "↑" : "↓")}
                  </button>
                </th>
                <th>Status</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {paged.rows.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-10 text-center text-sm text-[#94A3B8]">
                    Belum ada data booking.
                  </td>
                </tr>
              ) : (
                paged.rows.map((b, index) => (
                  <tr key={b.id}>
                    <td className="tabular-nums text-[#698097]">{(paged.page - 1) * pageSize + index + 1}.</td>
                    <td className="font-medium text-[#17324d]">{b.booking_code}</td>
                    <td>{b.visitor_name}</td>
                    <td className="tabular-nums">{b.visit_date}</td>
                    <td><StatusPill status={STATUS_ID[b.status]} /></td>
                    <td>
                      <div className="flex items-center gap-1.5">
                        <button
                          onClick={() => setDetail(b)}
                          className="grid h-7 w-7 place-items-center rounded-md border border-[#d7e3ed] text-[#29445e] hover:bg-[#f2f7fb]"
                          aria-label={`Detail ${b.id}`}
                        >
                          <Eye className="h-3.5 w-3.5" />
                        </button>
                        <select
                          value={b.status}
                          onChange={(e) => changeStatus(b, e.target.value as BookingStatus)}
                          className="h-7 rounded-md border border-[#d7e3ed] bg-white px-1.5 text-xs outline-none focus:border-[#1768ad]"
                          aria-label={`Ubah status ${b.id}`}
                        >
                          {STATUS_OPTIONS.map((s) => (
                            <option key={s} value={s}>{STATUS_ID[s]}</option>
                          ))}
                        </select>
                        <button
                          onClick={() => setDelTarget(b)}
                          className="grid h-7 w-7 place-items-center rounded-md border border-[#fdeaea] text-[#c53030] hover:bg-[#fdeaea]"
                          aria-label={`Hapus ${b.id}`}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        <Pagination page={paged.page} totalPages={paged.totalPages} total={paged.total} onPage={setPage} pageSize={pageSize} onPageSize={setPageSize} />
      </SectionCard>

      {/* Konfirmasi hapus */}
      {delTarget && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" onClick={() => setDelTarget(null)} role="dialog" aria-modal="true">
          <div className="w-full max-w-sm rounded-md bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-base font-bold text-[#17324d]">Hapus booking ini?</h3>
            <p className="mt-2 text-sm text-[#698097]">
              Booking <span className="font-mono font-semibold text-[#17324d]">{delTarget.booking_code}</span> ({delTarget.visitor_name}) akan dibatalkan dan hilang dari daftar aktif.
            </p>
            <div className="mt-5 flex justify-end gap-2">
              <button onClick={() => setDelTarget(null)} className="h-9 rounded-md border border-[#d7e3ed] px-4 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]">Batal</button>
              <button onClick={handleDelete} disabled={deleting} className="h-9 rounded-md bg-[#c53030] px-4 text-sm font-semibold text-white hover:brightness-95 disabled:opacity-60">
                {deleting ? "Menghapus..." : "Hapus"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Detail dialog */}
      {detail && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" onClick={() => setDetail(null)} role="dialog" aria-modal="true">
          <div className="max-h-[85vh] w-full max-w-lg overflow-y-auto rounded-md bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between border-b border-[#e5edf3] pb-3">
              <h3 className="text-base font-bold text-[#17324d]">Detail Booking</h3>
              <button onClick={() => setDetail(null)} className="text-sm text-[#698097] hover:text-[#17324d]" aria-label="Tutup">✕</button>
            </div>
            <dl className="mt-3 space-y-2 text-sm">
              <Row label="Kode Booking" value={detail.booking_code} />
              <Row label="Nama" value={detail.visitor_name} />
              <Row label="Email" value={detail.email} />
              <Row label="Telepon" value={detail.phone} />
              <Row label="Tanggal Kunjungan" value={detail.visit_date} />
              <Row label="Status" value={STATUS_ID[detail.status]} />
              <Row label="Dibuat" value={new Date(detail.created_at).toLocaleString("id-ID")} />
              {detail.note && <Row label="Catatan" value={detail.note} />}
            </dl>
            <div className="mt-3 rounded-md bg-[#f7fbff] p-3">
              <p className="text-xs font-bold text-[#29445e]">Rincian Tiket</p>
              <ul className="mt-2 space-y-1.5 text-sm">
                {detail.items.map((i) => (
                  <li key={i.ticket_id} className="flex justify-between">
                    <span className="text-[#698097]">{i.ticket_name} × {i.quantity}</span>
                    <span className="tabular-nums font-medium">{rupiah(i.price * i.quantity)}</span>
                  </li>
                ))}
              </ul>
              <div className="mt-2 flex justify-between border-t border-[#e5edf3] pt-2 text-sm font-bold">
                <span>Total</span>
                <span className="tabular-nums">{rupiah(detail.total)}</span>
              </div>
            </div>
            {detail.ticket_id && (
              <p className="mt-3 text-xs text-[#698097]">Kode E-Tiket: <span className="font-mono font-bold text-[#17324d]">{detail.ticket_id}</span></p>
            )}
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

function Pagination({
  page, totalPages, total, onPage, pageSize, onPageSize,
}: {
  page: number; totalPages: number; total: number; onPage: (p: number) => void; pageSize: number; onPageSize: (n: number) => void;
}) {
  return (
    <div className="mt-3 flex flex-wrap items-center justify-between gap-2 text-xs text-[#698097]">
      <div className="flex items-center gap-2">
        <span>Baris/hal</span>
        <select
          value={pageSize}
          onChange={(e) => onPageSize(Number(e.target.value))}
          className="h-7 rounded-md border border-[#d7e3ed] bg-white px-1.5 outline-none"
          aria-label="Ukuran halaman"
        >
          {[8, 15, 30].map((n) => <option key={n} value={n}>{n}</option>)}
        </select>
        <span>{total} data</span>
      </div>
      <div className="flex items-center gap-1">
        <button onClick={() => onPage(page - 1)} disabled={page <= 1} className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40">‹</button>
        <span className="px-2">Hal {page} / {totalPages}</span>
        <button onClick={() => onPage(page + 1)} disabled={page >= totalPages} className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40">›</button>
      </div>
    </div>
  );
}