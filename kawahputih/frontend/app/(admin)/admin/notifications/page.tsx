"use client";

import { RefreshCw, History } from "lucide-react";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { useAuditFeed } from "@/hooks/use-audit";
import { auditActionLabel } from "@/services/audit.api.service";

// Aktivitas admin — feed aktivitas nyata dari audit log backend (GET /admin/audit). Tidak ada mock DB di sini.

function statusLabel(status: string): string {
  switch (status) {
    case "SUCCESS":
      return "Berhasil";
    case "FAILED":
      return "Gagal";
    case "ERROR":
      return "Bermasalah";
    default:
      return status;
  }
}

export default function AdminActivityPage() {
  const { data, isLoading, isError, refetch, isFetching } = useAuditFeed({ limit: 50 });
  const logs = data?.items ?? [];
  const total = data?.meta?.total ?? logs.length;

  return (
    <div>
      <PageHeader
        title="Aktivitas"
        subtitle={`Feed audit dari backend — ${total.toLocaleString("id-ID")} entri tercatat`}
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
        {isLoading ? (
          <div className="animate-pulse space-y-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-12 rounded border border-[#e5edf3] bg-gray-100" />
            ))}
          </div>
        ) : isError ? (
          <p className="py-10 text-center text-sm text-[#c53030]">Gagal memuat feed aktivitas.</p>
        ) : logs.length === 0 ? (
          <p className="py-10 text-center text-sm text-[#94A3B8]">Belum ada aktivitas tercatat.</p>
        ) : (
          <ul className="divide-y divide-[#eef2f6]">
            {logs.map((log) => (
              <li key={log.id} className="flex items-start gap-3 py-3">
                <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-[#e9f3fb] text-[#1768ad]">
                  <History className="h-4 w-4" />
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="text-sm font-semibold text-[#17324d]">{auditActionLabel(log.action)}</p>
                    <StatusPill status={statusLabel(log.status)} />
                  </div>
                  <p className="mt-0.5 text-sm text-[#698097]">
                    {log.actor_role || "SYSTEM"} · {log.target_type || "-"}
                    {log.target_id ? ` · ${log.target_id.slice(0, 8)}` : ""}
                  </p>
                  <p className="mt-1 text-[11px] text-[#94A3B8]">
                    {new Date(log.created_at).toLocaleString("id-ID")}
                  </p>
                </div>
              </li>
            ))}
          </ul>
        )}
      </SectionCard>
    </div>
  );
}
