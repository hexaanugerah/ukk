"use client";

import { useEffect, useMemo, useState } from "react";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { SearchInput } from "@/components/panel/panel-toolbar";
import { scannerService, type ScanRecord } from "@/services/scanner.api.service";
import { paginate, searchIn, sortData } from "@/lib/business";

const STATUS_OPTIONS = ["VALID", "USED", "INVALID", "EXPIRED", "CANCELLED"];
const STATUS_ID: Record<string, string> = {
  VALID: "Valid",
  USED: "Sudah Digunakan",
  INVALID: "Invalid",
  EXPIRED: "Kadaluarsa",
  CANCELLED: "Dibatalkan",
};

export default function PetugasHistoryPage() {
  const [scans, setScans] = useState<ScanRecord[]>([]);
  const [loading, setLoading] = useState(true);

  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");
  const [page, setPage] = useState(1);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    (async () => {
      try {
        const data = await scannerService.getHistory({ page, limit: 10, status: status || undefined });
        if (alive) setScans(data);
      } catch (error) {
        console.error("Failed to load scans:", error);
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => {
      alive = false;
    };
  }, [page, status]);

  // Perhitungan yang di-memoize (harus sebelum early return)
  const rows = useMemo(() => {
    let list = searchIn(scans, q, ["ticket_id", "booking_id", "visitor_name", "petugas"]);
    if (status) list = list.filter((s) => s.status === status);
    return sortData(list, "timestamp", sortDir);
  }, [scans, q, status, sortDir]);

  const paged = useMemo(() => paginate(rows, page, 10), [rows, page]);

  if (loading) {
    return (
      <div>
        <PageHeader title="Riwayat Validasi" subtitle="Memuat data..." />
        <SectionCard>
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
      <PageHeader title="Riwayat Validasi" subtitle="Riwayat Tiket" />

      <SectionCard action={<span className="text-sm font-semibold text-[#1768ad]">{rows.length} scan</span>}>
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <SearchInput value={q} onChange={(v) => { setQ(v); setPage(1); }} placeholder="Cari kode / nama / petugas…" className="w-full sm:max-w-xs" />
          <select
            value={status}
            onChange={(e) => { setStatus(e.target.value); setPage(1); }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter status scan"
          >
            <option value="">Semua status</option>
            {STATUS_OPTIONS.map((s) => (
              <option key={s} value={s}>{STATUS_ID[s]}</option>
            ))}
          </select>
          <button
            onClick={() => setSortDir((d) => (d === "asc" ? "desc" : "asc"))}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] hover:bg-[#f2f7fb]"
          >
            Waktu {sortDir === "asc" ? "↑" : "↓"}
          </button>
        </div>

        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>Waktu Scan</th>
                <th>Kode Tiket</th>
                <th>Nama Pengunjung</th>
                <th>Petugas</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {paged.rows.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-10 text-center text-sm text-[#94A3B8]">
                    Belum ada riwayat scan.
                  </td>
                </tr>
              ) : (
                paged.rows.map((s) => (
                  <tr key={s.id}>
                    <td className="tabular-nums text-xs">{new Date(s.timestamp).toLocaleString("id-ID")}</td>
                    <td className="font-medium text-[#17324d]">{s.ticket_id}</td>
                    <td>{s.visitor_name}</td>
                    <td>{s.petugas}</td>
                    <td><StatusPill status={STATUS_ID[s.status] ?? s.status} /></td>
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
    </div>
  );
}