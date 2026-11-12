"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { CreditCard, Wallet } from "lucide-react";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { bookingService } from "@/services/booking.api.service";
import { rupiah } from "@/lib/business";
import type { Booking } from "@/types/api";

// Pembayaran Saya — menautkan booking PENDING dengan halaman pembayaran dan
// menampilkan status pembayaran terakhir dari backend (payment snapshot).

const PAYMENT_ID: Record<string, string> = {
  PENDING: "Menunggu",
  PROCESSING: "Diproses",
  PAID: "Berhasil",
  FAILED: "Gagal",
  EXPIRED: "Kadaluarsa",
  REFUNDED: "Refund",
};

export default function MyPaymentsPage() {
  const [bookings, setBookings] = useState<Booking[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let alive = true;
    bookingService
      .listMine({ limit: 50 })
      .then((rows) => {
        if (alive) setBookings(rows);
      })
      .catch((e) => {
        if (alive) setError(e?.message || "Gagal memuat pembayaran");
      });
    return () => {
      alive = false;
    };
  }, []);

  const pending = useMemo(
    () => (bookings ?? []).filter((b) => b.status === "PENDING"),
    [bookings]
  );
  const history = useMemo(
    () => (bookings ?? []).filter((b) => b.status !== "PENDING"),
    [bookings]
  );

  if (bookings === null) {
    return (
      <div>
        <PageHeader title="Pembayaran" subtitle="Memuat transaksi..." />
        <div className="h-48 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />
      </div>
    );
  }

  return (
    <div>
      <PageHeader title="Pembayaran" subtitle="Selesaikan pembayaran dan pantau status transaksi" />

      {error && (
        <p className="mb-3 rounded-md bg-[#fdeaea] px-3 py-2 text-sm text-[#c53030]">{error}</p>
      )}

      <SectionCard title="Menunggu Pembayaran" className="mb-3">
        {pending.length === 0 ? (
          <p className="py-6 text-center text-sm text-[#94A3B8]">Tidak ada pembayaran yang menunggu.</p>
        ) : (
          <ul className="divide-y divide-[#eef2f6]">
            {pending.map((b) => (
              <li key={b.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                <div>
                  <p className="text-sm font-semibold text-[#17324d]">{b.booking_code}</p>
                  <p className="text-xs text-[#698097]">
                    Kunjungan {b.visit_date} • {b.total_tickets} tiket • {rupiah(b.total)}
                  </p>
                  {b.payment && (
                    <p className="mt-0.5 text-[11px] text-[#94A3B8]">
                      {b.payment.method} • {PAYMENT_ID[b.payment.status] ?? b.payment.status}
                    </p>
                  )}
                </div>
                <Link
                  href={`/payment?code=${b.id}`}
                  className="inline-flex h-8 items-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f]"
                >
                  <Wallet className="h-3.5 w-3.5" /> Bayar Sekarang
                </Link>
              </li>
            ))}
          </ul>
        )}
      </SectionCard>

      <SectionCard title="Riwayat Pembayaran">
        {history.length === 0 ? (
          <p className="py-6 text-center text-sm text-[#94A3B8]">Belum ada riwayat pembayaran.</p>
        ) : (
          <div className="panel-scroll rounded-md border border-[#e5edf3]">
            <table className="panel-table">
              <thead>
                <tr>
                  <th>Kode Booking</th>
                  <th>Tanggal Kunjungan</th>
                  <th>Total</th>
                  <th>Metode</th>
                  <th>Status Pembayaran</th>
                  <th>Aksi</th>
                </tr>
              </thead>
              <tbody>
                {history.map((b) => (
                  <tr key={b.id}>
                    <td className="font-medium text-[#17324d]">{b.booking_code}</td>
                    <td className="tabular-nums">{b.visit_date}</td>
                    <td className="tabular-nums">{rupiah(b.total)}</td>
                    <td>{b.payment?.method ?? "-"}</td>
                    <td>
                      <StatusPill
                        status={b.payment ? PAYMENT_ID[b.payment.status] ?? b.payment.status : "Belum ada"}
                      />
                    </td>
                    <td>
                      <Link
                        href={
                          ["PAID", "CONFIRMED", "USED"].includes(b.status)
                            ? `/e-ticket?code=${b.id}`
                            : `/payment?code=${b.id}`
                        }
                        className="inline-flex h-7 items-center gap-1 rounded-md border border-[#d7e3ed] px-2 text-xs font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
                      >
                        <CreditCard className="h-3.5 w-3.5" /> Detail
                      </Link>
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
