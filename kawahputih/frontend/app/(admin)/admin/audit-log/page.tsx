"use client";

import { useState } from "react";
import { RefreshCw } from "lucide-react";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { useAuditFeed } from "@/hooks/use-audit";
import { auditActionLabel } from "@/services/audit.api.service";

// Audit log — riwayat aktivitas sistem, termasuk setiap hasil validasi tiket
// oleh petugas (valid maupun ditolak). Sumber: GET /admin/audit.

const ACTIONS = [
  { value: "", label: "Semua aksi" },
  { value: "ticket.validate", label: "Validasi tiket" },
  { value: "ticket.issue", label: "Penerbitan tiket" },
  { value: "booking.create", label: "Booking baru" },
  { value: "booking.cancel", label: "Booking dibatalkan" },
  { value: "booking.status_change", label: "Status booking" },
  { value: "payment.process", label: "Pembayaran" },
  { value: "payment.status_change", label: "Status pembayaran" },
  { value: "user.login", label: "Login" },
  { value: "user.register", label: "Registrasi" },
];

const STATUSES = [
  { value: "", label: "Semua status" },
  { value: "SUCCESS", label: "Berhasil" },
  { value: "FAILED", label: "Gagal" },
  { value: "ERROR", label: "Bermasalah" },
];

function statusLabel(status: string): string {
  return STATUSES.find((s) => s.value === status)?.label ?? status;
}

export default function AdminAuditLogPage() {
  const [page, setPage] = useState(1);
  const [action, setAction] = useState("");
  const [status, setStatus] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");

  const { data, isLoading, isError, refetch, isFetching } = useAuditFeed({
    page,
    limit: 20,
    action,
    status,
    from,
    to,
  });

  const logs = data?.items ?? [];
  const meta = data?.meta;
  const totalPages = meta?.total_pages ?? 1;

  return (
    <div>
      <PageHeader
        title="Audit Log"
        subtitle="Riwayat aktivitas seluruh modul, termasuk hasil validasi tiket"
        action={
          <button
            onClick={() => refetch()}
            disabled={isFetching}
            className="inline-flex h-9 items-center gap-1.5 rounded-md border border-[#d7e3ed] bg-white px-3 text-xs font-semibold text-[#29445e] hover:bg-[#f2f7fb] disabled:opacity-50"
          >
            <RefreshCw className={isFetching ? "h-4 w-4 animate-spin" : "h-4 w-4"} /> Muat ulang
          </button>
        }
      />

      <SectionCard>
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <select
            value={action}
            onChange={(e) => {
              setAction(e.target.value);
              setPage(1);
            }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter aksi"
          >
            {ACTIONS.map((a) => (
              <option key={a.value} value={a.value}>
                {a.label}
              </option>
            ))}
          </select>
          <select
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter status"
          >
            {STATUSES.map((s) => (
              <option key={s.value} value={s.value}>
                {s.label}
              </option>
            ))}
          </select>
          <input
            type="date"
            value={from}
            onChange={(e) => {
              setFrom(e.target.value);
              setPage(1);
            }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Dari tanggal"
          />
          <input
            type="date"
            value={to}
            onChange={(e) => {
              setTo(e.target.value);
              setPage(1);
            }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Sampai tanggal"
          />
          <span className="ml-auto text-xs text-[#698097]">
            {(meta?.total ?? logs.length).toLocaleString("id-ID")} entri
          </span>
        </div>

        {isLoading ? (
          <div className="animate-pulse space-y-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-12 rounded border border-[#e5edf3] bg-gray-100" />
            ))}
          </div>
        ) : isError ? (
          <p className="py-10 text-center text-sm text-[#c53030]">Gagal memuat audit log.</p>
        ) : logs.length === 0 ? (
          <p className="py-10 text-center text-sm text-[#94A3B8]">Tidak ada aktivitas sesuai filter.</p>
        ) : (
          <div className="panel-scroll rounded-md border border-[#e5edf3]">
            <table className="panel-table">
              <thead>
                <tr>
                  <th>Waktu</th>
                  <th>Aktivitas</th>
                  <th>Aktor</th>
                  <th>Target</th>
                  <th>Status</th>
                  <th>Keterangan</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((log) => (
                  <tr key={log.id}>
                    <td className="whitespace-nowrap tabular-nums text-[#698097]">
                      {new Date(log.created_at).toLocaleString("id-ID")}
                    </td>
                    <td className="font-medium text-[#17324d]">{auditActionLabel(log.action)}</td>
                    <td>{log.actor_role || "SYSTEM"}</td>
                    <td className="text-[11px] text-[#698097]">
                      {log.target_type}
                      {log.target_id ? ` · ${log.target_id.slice(0, 8)}` : ""}
                    </td>
                    <td>
                      <StatusPill status={statusLabel(log.status)} />
                    </td>
                    <td className="max-w-[220px] truncate text-[11px] text-[#698097]">
                      {typeof log.metadata?.scan_status === "string"
                        ? `Scan: ${log.metadata.scan_status}${log.metadata.message ? ` — ${log.metadata.message}` : ""}`
                        : typeof log.metadata?.booking_code === "string"
                          ? `Booking ${log.metadata.booking_code}`
                          : "-"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {totalPages > 1 && (
          <div className="mt-3 flex items-center justify-end gap-1 text-xs text-[#698097]">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40"
            >
              ‹
            </button>
            <span className="px-2">
              Hal {page} / {totalPages}
            </span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40"
            >
              ›
            </button>
          </div>
        )}
      </SectionCard>
    </div>
  );
}
