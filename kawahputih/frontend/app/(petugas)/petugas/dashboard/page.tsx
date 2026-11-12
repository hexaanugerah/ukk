"use client";

import { useEffect, useState } from "react";
import { PageHeader, SectionCard, StatGrid, StatusPill } from "@/components/panel/panel-ui";
import { useAuthStore } from "@/store/auth.store";
import { scannerService, type ScanRecord } from "@/services/scanner.api.service";
import { todayISO } from "@/lib/business";

// Dashboard petugas — ringkasan aktivitas gerbang hari ini, dari backend API.

export default function PetugasDashboardPage() {
  const user = useAuthStore((s) => s.user);
  const today = todayISO();

  const [scans, setScans] = useState<ScanRecord[]>([]);
  const [eTickets, setETickets] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDashboardData();
  }, []);

  async function loadDashboardData() {
    setLoading(true);
    try {
      const [history, stats] = await Promise.all([
        scannerService.getHistory({ page: 1, limit: 8 }),
        scannerService.getStats(),
      ]);
      setScans(history);
      // stats berisi total, valid, used, invalid, expired, cancelled. Nilainya dipakai untuk KPI.
    } catch (error) {
      console.error("Failed to load dashboard:", error);
    } finally {
      setLoading(false);
    }
  }

  if (loading) {
    return (
      <div>
        <PageHeader
          title="Dashboard"
          subtitle="Memuat data..."
        />
        <StatGrid cols={4}>
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="rounded-md border border-[#d7e3ed] bg-white px-3 py-2.5 animate-pulse">
              <p className="text-[10px] font-semibold text-[#698097]">Loading...</p>
              <p className="text-[10px] text-[#698097]">sub</p>
              <p className="mt-1.5 text-lg font-bold tabular-nums text-[#17324d]">—</p>
            </div>
          ))}
        </StatGrid>
      </div>
    );
  }

  const todayScans = scans.filter((s) => s.timestamp.slice(0, 10) === today);
  const validToday = todayScans.filter((s) => s.status === "VALID").length;
  const usedToday = todayScans.filter((s) => s.status === "USED").length;
  const problemToday = todayScans.filter((s) => ["INVALID", "EXPIRED", "CANCELLED"].includes(s.status)).length;

  const kpis = [
    { label: "Scan Valid", sub: "Hari Ini", value: validToday },
    { label: "Tiket Sudah", sub: "Digunakan", value: usedToday },
    { label: "Tiket", sub: "Bermasalah", value: problemToday },
    { label: "Total Tiket", sub: "Digunakan Sistem", value: 0 }, // Would need separate eTicket API
  ];

  return (
    <div>
      <PageHeader
        title="Dashboard"
        subtitle={`Ringkasan Aktivitas ${user?.gate ?? "Gerbang"} — ${user?.name ?? "Petugas"}`}
      />

      <StatGrid cols={4}>
        {kpis.map((k) => (
          <div key={k.label + k.sub} className="rounded-md border border-[#d7e3ed] bg-white px-3 py-2.5">
            <p className="text-[10px] font-semibold text-[#698097]">{k.label}</p>
            <p className="text-[10px] text-[#698097]">{k.sub}</p>
            <p className="mt-1.5 text-lg font-bold tabular-nums text-[#17324d]">{k.value.toLocaleString("id-ID")}</p>
          </div>
        ))}
      </StatGrid>

      <SectionCard title="Riwayat Scan Terbaru" className="mt-2.5">
        <div className="panel-scroll">
          <table className="panel-table">
            <thead>
              <tr>
                <th>NO</th>
                <th>Waktu Scan</th>
                <th>Kode Tiket</th>
                <th>Nama Pengunjung</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {scans.slice(0, 8).length === 0 ? (
                <tr><td colSpan={5} className="py-6 text-center text-[11px] text-[#94A3B8]">Belum ada scan hari ini.</td></tr>
              ) : (
                scans.slice(0, 8).map((s, i) => (
                  <tr key={s.id}>
                    <td className="tabular-nums text-[#698097]">{i + 1}.</td>
                    <td className="tabular-nums">{new Date(s.timestamp).toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" })}</td>
                    <td className="font-medium text-[#17324d]">{s.ticket_id}</td>
                    <td>{s.visitor_name}</td>
                    <td>
                      <StatusPill
                        status={
                          s.status === "VALID" ? "Valid" :
                          s.status === "USED" ? "Sudah Digunakan" :
                          s.status === "EXPIRED" ? "Kadaluarsa" :
                          s.status === "CANCELLED" ? "Dibatalkan" : "Bermasalah"
                        }
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