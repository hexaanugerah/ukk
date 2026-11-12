"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { History } from "lucide-react";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { bookingService } from "@/services/booking.api.service";
import { rupiah } from "@/lib/business";
import type { Booking, BookingStatus } from "@/types/api";

// Riwayat kunjungan — semua booking pengunjung, terbaru lebih dulu.
// Status diambil langsung dari backend sehingga konsisten dengan tiket.

const STATUS_ID: Record<BookingStatus, string> = {
  PENDING: "Menunggu Pembayaran",
  PAID: "Dibayar",
  CONFIRMED: "Terkonfirmasi",
  USED: "Digunakan",
  CANCELLED: "Dibatalkan",
  EXPIRED: "Kadaluarsa",
};

export default function MyHistoryPage() {
  const [bookings, setBookings] = useState<Booking[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let alive = true;
    bookingService
      .listMine({ limit: 100 })
      .then((rows) => {
        if (alive) setBookings(rows);
      })
      .catch((e) => {
        if (alive) setError(e?.message || "Gagal memuat riwayat");
      });
    return () => {
      alive = false;
    };
  }, []);

  const rows = useMemo(
    () =>
      [...(bookings ?? [])].sort((a, b) => (b.created_at ?? "").localeCompare(a.created_at ?? "")),
    [bookings]
  );

  if (bookings === null) {
    return (
      <div>
        <PageHeader title="Riwayat" subtitle="Memuat riwayat..." />
        <div className="h-48 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />
      </div>
    );
  }

  return (
    <div>
      <PageHeader title="Riwayat" subtitle="Riwayat booking dan kunjungan kamu" />

      {error && (
        <p className="mb-3 rounded-md bg-[#fdeaea] px-3 py-2 text-sm text-[#c53030]">{error}</p>
      )}

      <SectionCard>
        {rows.length === 0 ? (
          <div className="py-10 text-center">
            <History className="mx-auto h-10 w-10 text-[#94A3B8]" />
            <p className="mt-2 text-sm text-[#698097]">Belum ada riwayat booking.</p>
            <Link
              href="/packages"
              className="mt-4 inline-flex h-9 items-center rounded-md bg-[#1768ad] px-4 text-sm font-semibold text-white hover:bg-[#14548f]"
            >
              Pesan Tiket
            </Link>
          </div>
        ) : (
          <div className="panel-scroll rounded-md border border-[#e5edf3]">
            <table className="panel-table">
              <thead>
                <tr>
                  <th>Kode Booking</th>
                  <th>Tanggal Kunjungan</th>
                  <th>Tiket</th>
                  <th>Total</th>
                  <th>Status</th>
                  <th>Aksi</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((b) => (
                  <tr key={b.id}>
                    <td className="font-medium text-[#17324d]">{b.booking_code}</td>
                    <td className="tabular-nums">{b.visit_date}</td>
                    <td className="tabular-nums">{b.total_tickets}</td>
                    <td className="tabular-nums">{rupiah(b.total)}</td>
                    <td>
                      <StatusPill status={STATUS_ID[b.status] ?? b.status} />
                    </td>
                    <td>
                      {["PAID", "CONFIRMED", "USED"].includes(b.status) ? (
                        <Link
                          href={`/e-ticket?code=${b.id}`}
                          className="font-semibold text-[#1768ad] hover:underline"
                        >
                          E-Tiket
                        </Link>
                      ) : b.status === "PENDING" ? (
                        <Link
                          href={`/payment?code=${b.id}`}
                          className="font-semibold text-[#1768ad] hover:underline"
                        >
                          Bayar
                        </Link>
                      ) : (
                        <span className="text-[#94A3B8]">-</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </SectionCard>
    </div>
  );
}
