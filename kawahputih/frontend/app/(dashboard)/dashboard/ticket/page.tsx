"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import QRCode from "react-qr-code";
import { QrCode, Download, TicketCheck } from "lucide-react";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { eTicketService } from "@/services/eticket.api.service";
import { rupiah } from "@/lib/business";
import type { ETicket } from "@/types/api";

// Tiket Saya — e-tiket yang diterbitkan setelah pembayaran berhasil.
// Sumber data tunggal: GET /e-tickets/mine, sehingga status tiket selalu sama
// dengan yang dibaca petugas saat scan.

const STATUS_ID: Record<string, string> = {
  VALID: "Valid",
  USED: "Sudah Digunakan",
  EXPIRED: "Kadaluarsa",
  CANCELLED: "Dibatalkan",
};

type Tab = "active" | "used" | "inactive";

export default function MyTicketsPage() {
  const [tickets, setTickets] = useState<ETicket[] | null>(null);
  const [tab, setTab] = useState<Tab>("active");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let alive = true;
    eTicketService
      .listMine()
      .then((rows) => {
        if (alive) setTickets(rows);
      })
      .catch((e) => {
        if (alive) setError(e?.message || "Gagal memuat e-tiket");
      });
    return () => {
      alive = false;
    };
  }, []);

  const grouped = useMemo(() => {
    const rows = tickets ?? [];
    return {
      active: rows.filter((t) => t.status === "VALID"),
      used: rows.filter((t) => t.status === "USED"),
      inactive: rows.filter((t) => t.status === "EXPIRED" || t.status === "CANCELLED"),
    };
  }, [tickets]);

  if (tickets === null) {
    return (
      <div>
        <PageHeader title="Tiket Saya" subtitle="Memuat e-tiket..." />
        <div className="grid gap-4 sm:grid-cols-2">
          {[1, 2].map((i) => (
            <div key={i} className="h-64 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />
          ))}
        </div>
      </div>
    );
  }

  const tabs: { key: Tab; label: string; count: number }[] = [
    { key: "active", label: "Aktif", count: grouped.active.length },
    { key: "used", label: "Digunakan", count: grouped.used.length },
    { key: "inactive", label: "Tidak Aktif", count: grouped.inactive.length },
  ];
  const visible = grouped[tab];

  return (
    <div>
      <PageHeader title="Tiket Saya" subtitle="E-tiket aktif dan riwayat penggunaan tiket" />

      {error && (
        <p className="mb-3 rounded-md bg-[#fdeaea] px-3 py-2 text-sm text-[#c53030]">{error}</p>
      )}

      <div className="mb-3 flex flex-wrap gap-2">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`rounded-full px-4 py-1.5 text-xs font-semibold ${
              tab === t.key ? "bg-[#1768ad] text-white" : "bg-white text-[#29445e] hover:bg-[#f2f7fb]"
            }`}
          >
            {t.label} ({t.count})
          </button>
        ))}
      </div>

      {visible.length === 0 ? (
        <SectionCard>
          <div className="py-10 text-center">
            <TicketCheck className="mx-auto h-10 w-10 text-[#94A3B8]" />
            <p className="mt-2 text-sm text-[#698097]">
              {tab === "active" ? "Belum ada e-tiket aktif." : "Tidak ada tiket pada kategori ini."}
            </p>
            {tab === "active" && (
              <Link
                href="/packages"
                className="mt-4 inline-flex h-9 items-center rounded-md bg-[#1768ad] px-4 text-sm font-semibold text-white hover:bg-[#14548f]"
              >
                Pesan Tiket
            </Link>
            )}
          </div>
        </SectionCard>
      ) : (
        <div className="grid gap-4 lg:grid-cols-2">
          {visible.map((ticket) => (
            <SectionCard key={ticket.id}>
              <div className="flex items-start justify-between gap-3 border-b border-[#e5edf3] pb-3">
                <div>
                  <p className="text-xs text-[#698097]">Kode Booking</p>
                  <p className="text-base font-bold tracking-wide text-[#17324d]">{ticket.booking_code}</p>
                  <p className="mt-1 text-xs text-[#698097]">{ticket.ticket_name} × {ticket.quantity}</p>
                </div>
                <StatusPill status={STATUS_ID[ticket.status] ?? ticket.status} />
              </div>

              <div className="mt-3 flex flex-wrap items-center gap-4">
                <div className="rounded-md border border-[#e5edf3] bg-white p-2">
                  <QRCode value={ticket.qr_data || ticket.id} size={104} aria-label={`QR ${ticket.booking_code}`} />
                </div>
                <div className="min-w-0 flex-1 text-sm">
                  <p className="text-[#698097]">Nama</p>
                  <p className="font-semibold text-[#17324d]">{ticket.visitor_name}</p>
                  <p className="mt-2 text-[#698097]">Tanggal Kunjungan</p>
                  <p className="font-semibold text-[#17324d]">{ticket.visit_date}</p>
                  <p className="mt-2 text-[#698097]">Total</p>
                  <p className="font-semibold tabular-nums text-[#17324d]">{rupiah(ticket.total)}</p>
                </div>
              </div>

              <div className="mt-3 flex flex-wrap gap-2 border-t border-[#e5edf3] pt-3">
                <Link
                  href={`/e-ticket?code=${ticket.booking_id}`}
                  className="inline-flex h-8 items-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f]"
                >
                  <QrCode className="h-3.5 w-3.5" /> Lihat E-Tiket
                </Link>
                <Link
                  href={`/e-ticket?code=${ticket.booking_id}`}
                  className="inline-flex h-8 items-center gap-1.5 rounded-md border border-[#d7e3ed] px-3 text-xs font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
                >
                  <Download className="h-3.5 w-3.5" /> Unduh PDF
                </Link>
              </div>
            </SectionCard>
          ))}
        </div>
      )}
    </div>
  );
}
