"use client";

import { useQuery } from "@tanstack/react-query";
import { BarChart, PageHeader, SectionCard, StatCard, StatGrid, StatusPill } from "@/components/panel/panel-ui";
import { staffService } from "@/services/staff.api.service";
import { scannerService } from "@/services/scanner.api.service";
import type { ManagerMonitoring, Staff } from "@/types/api";

// Monitoring petugas — hanya baca. Manager menerima angka agregat dari
// /manager/monitoring, bukan riwayat scan mentah yang merupakan area petugas.

export default function ManagerMonitoringPage() {
  const staffQuery = useQuery<Staff[]>({
    queryKey: ["admin", "staff", "monitoring"],
    queryFn: () => staffService.list(),
  });

  const monitorQuery = useQuery<ManagerMonitoring>({
    queryKey: ["manager", "monitoring"],
    queryFn: () => scannerService.getMonitoring(),
  });

  const staff = staffQuery.data ?? [];
  const monitor = monitorQuery.data;
  const perPetugas = monitor?.per_petugas ?? [];

  const staffStats = staff.map((s) => {
    const own = perPetugas.find((p) => p.name === s.name);
    return {
      id: s.id,
      name: s.name,
      role: s.role,
      gate: s.gate || own?.gate,
      validated: own?.valid ?? 0,
      lastActivity: own?.last_activity ?? null,
      status: s.status,
    };
  });

  const activeStaff = staff.filter((s) => s.status === "AKTIF").length;
  const totalValidated = monitor?.total_valid ?? 0;
  const validatedToday = monitor?.valid_today ?? 0;
  const perHour = (monitor?.hourly ?? []).map((h) => ({
    label: `${String(h.hour).padStart(2, "0")}:00`,
    value: h.count,
  }));

  const loading = staffQuery.isLoading || monitorQuery.isLoading;
  const errored = staffQuery.isError || monitorQuery.isError;

  return (
    <div>
      <PageHeader title="Monitoring Petugas" subtitle="Aktivitas & Kinerja Petugas" />

      {errored && (
        <p className="mb-3 rounded-md bg-[#fdeaea] px-3 py-2 text-sm text-[#c53030]">
          Gagal memuat data monitoring.
        </p>
      )}

      <StatGrid cols={3}>
        <StatCard label="Petugas Aktif" value={loading ? "…" : `${activeStaff} / ${staff.length}`} />
        <StatCard label="Total Tiket Validasi" value={loading ? "…" : totalValidated.toLocaleString("id-ID")} />
        <StatCard label="Validasi Hari Ini" value={loading ? "…" : validatedToday.toLocaleString("id-ID")} />
      </StatGrid>

      <div className="mt-3 grid gap-3 lg:grid-cols-3">
        <SectionCard title="Aktivitas Petugas" className="lg:col-span-2">
          <div className="panel-scroll">
            <table className="panel-table">
              <thead>
                <tr>
                  <th>Nama Petugas</th>
                  <th>Role</th>
                  <th>Gate</th>
                  <th>Tiket Validasi</th>
                  <th>Aktivitas Terakhir</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr><td colSpan={6} className="py-10 text-center text-sm text-[#94A3B8]">Memuat…</td></tr>
                ) : staffStats.length === 0 ? (
                  <tr><td colSpan={6} className="py-10 text-center text-sm text-[#94A3B8]">Tidak ada data petugas.</td></tr>
                ) : (
                  staffStats.map((o) => (
                    <tr key={o.id}>
                      <td className="font-medium text-[#17324d]">{o.name}</td>
                      <td>{o.role}</td>
                      <td>{o.gate || "-"}</td>
                      <td className="tabular-nums">{o.validated.toLocaleString("id-ID")}</td>
                      <td className="text-xs text-[#698097]">
                        {o.lastActivity ? new Date(o.lastActivity).toLocaleString("id-ID") : "-"}
                      </td>
                      <td><StatusPill status={o.status === "AKTIF" ? "Aktif" : "Nonaktif"} /></td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </SectionCard>

        <SectionCard title="Validasi Per Jam (hari ini)">
          <BarChart data={perHour} height={160} />
        </SectionCard>
      </div>

      <SectionCard title="Rekap Kunjungan" className="mt-3">
        <div className="flex items-center justify-between">
          <span className="text-sm text-[#29445e]">Total kunjungan tercatat (semua waktu)</span>
          <span className="text-2xl font-bold tabular-nums text-[#1768ad]">
            {totalValidated.toLocaleString("id-ID")}
          </span>
        </div>
      </SectionCard>
    </div>
  );
}
