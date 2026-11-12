"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuthStore } from "@/store/auth.store";
import { bookingService } from "@/services/booking.api.service";
import {
  PageHeader,
  SectionCard,
  StatCard,
  StatGrid,
  StatusPill,
} from "@/components/panel/panel-ui";
import { rupiah } from "@/lib/business";
import type { Booking } from "@/types/api";

// Dashboard pengunjung — ringkasan aktivitas booking pengguna, dari backend API.

const STATUS_ID: Record<string, string> = {
  PENDING: "Menunggu Pembayaran",
  PAID: "Dibayar",
  CONFIRMED: "Terkonfirmasi",
  USED: "Digunakan",
  CANCELLED: "Dibatalkan",
  EXPIRED: "Kadaluarsa",
};

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user);
  const [bookings, setBookings] = useState<Booking[] | null>(null);

  useEffect(() => {
    bookingService
      .listMine()
      .then(setBookings)
      .catch((e) => console.error("Failed to load bookings:", e));
  }, []);

  const list = bookings ?? [];
  const pending = list.filter((b) => b.status === "PENDING").length;
  const confirmed = list.filter((b) => ["PAID", "CONFIRMED"].includes(b.status)).length;
  const used = list.filter((b) => b.status === "USED").length;

  return (
    <div>
      <PageHeader
        title="Dashboard"
        subtitle={`Ringkasan Aktivitas — ${user?.name ?? "Pengunjung"}`}
      />

      <StatGrid cols={4}>
        <StatCard label="Total Booking" value={list.length.toLocaleString("id-ID")} />
        <StatCard label="Menunggu Pembayaran" value={pending.toLocaleString("id-ID")} />
        <StatCard label="Terkonfirmasi" value={confirmed.toLocaleString("id-ID")} />
        <StatCard label="Tiket Digunakan" value={used.toLocaleString("id-ID")} />
      </StatGrid>

      <SectionCard
        title="Booking Terbaru"
        subtitle="5 transaksi terakhir"
        action={
          <Link
            href="/dashboard/booking"
            className="text-[11px] font-bold text-[#1768ad] hover:underline"
          >
            Lihat semua
          </Link>
        }
        className="mt-2.5"
      >
        {list.length === 0 ? (
          <p className="py-4 text-center text-[11px] text-[#94A3B8]">
            Belum ada booking.{" "}
            <Link href="/packages" className="font-semibold text-[#1768ad] hover:underline">
              Pesan tiket sekarang
            </Link>
            .
          </p>
        ) : (
          <div className="panel-scroll">
            <table className="panel-table">
              <thead>
                <tr>
                  <th>Kode</th>
                  <th>Tanggal Kunjungan</th>
                  <th>Total</th>
                  <th>Status</th>
                  <th>Aksi</th>
                </tr>
              </thead>
              <tbody>
                {list.slice(0, 5).map((b) => (
                  <tr key={b.id}>
                    <td className="font-medium text-[#17324d]">{b.booking_code}</td>
                    <td className="tabular-nums">{b.visit_date}</td>
                    <td className="tabular-nums">{rupiah(b.total)}</td>
                    <td>
                      <StatusPill status={STATUS_ID[b.status] ?? b.status} />
                    </td>
                    <td>
                      {["PAID", "CONFIRMED", "USED"].includes(b.status) && (
                        <Link
                          href={`/e-ticket?code=${b.id}`}
                          className="font-semibold text-[#1768ad] hover:underline"
                        >
                          E-Tiket
                        </Link>
                      )}
                      {b.status === "PENDING" && (
                        <Link
                          href={`/payment?code=${b.id}`}
                          className="font-semibold text-[#1768ad] hover:underline"
                        >
                          Bayar
                        </Link>
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
