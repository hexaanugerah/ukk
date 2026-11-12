"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Download, Printer, Info, CheckCircle2, XCircle, Loader2 } from "lucide-react";
import QRCode from "react-qr-code";
import { toast } from "sonner";
import { bookingService, eTicketService, BookingError } from "@/services/api";
import { downloadETicketPDF } from "@/lib/eticket-pdf";
import { rupiah } from "@/lib/business";
import type { Booking, ETicket } from "@/types/api";

// Halaman e-tiket — menampilkan e-tiket dari backend API, QR code asli berisi data tiket, dan unduhan PDF.

export default function ETicketPage() {
  const [isNew, setIsNew] = useState(false);

  const [data, setData] = useState<{ ticket: ETicket; booking: Booking } | null>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);
  const [downloading, setDownloading] = useState(false);

  // Query string hanya terbaca di browser — prerender SSR tidak punya window.
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const q = params.get("code");
    setIsNew(params.get("new") === "1");

    let alive = true;
    (async () => {
      if (!q) {
        setNotFound(true);
        setLoading(false);
        return;
      }
      try {
        const res = await eTicketService.getByBookingId(q);
        if (!alive) return;
        if (res) setData(res);
        else setNotFound(true);
      } catch {
        if (alive) setNotFound(true);
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => {
      alive = false;
    };
  }, []);

  async function handleDownload() {
    if (!data) return;
    setDownloading(true);
    try {
      // Lib PDF memakai tipe domain camelCase, jadi petakan dari ETicket snake_case API sebelum render.
      const t = data.ticket;
      const pdfTicket = {
        ...t,
        bookingId: t.booking_id,
        visitorName: t.visitor_name,
        visitDate: t.visit_date,
        ticketName: t.ticket_name,
        generatedAt: t.generated_at,
        checkedInAt: t.checked_in_at,
        checkedInGate: t.checked_in_gate,
      };
      const b = data.booking;
      const pdfBooking = {
        ...b,
        userId: b.user_id ?? null,
        visitorName: b.visitor_name,
        visitDate: b.visit_date,
        items: b.items.map((it) => ({
          ticketId: it.ticket_id,
          ticketName: it.ticket_name,
          price: it.price,
          quantity: it.quantity,
        })),
        serviceFee: b.service_fee,
        totalTickets: b.total_tickets,
        createdAt: b.created_at,
        ticketId: b.ticket_id ?? null,
      };
      await downloadETicketPDF(pdfTicket as any, pdfBooking as any);
      await eTicketService.markDownloaded(data.booking.id);
      toast.success("E-tiket berhasil diunduh");
    } catch {
      toast.error("Gagal mengunduh e-tiket");
    } finally {
      setDownloading(false);
    }
  }

  if (loading) {
    return (
      <div className="page-shell py-10">
        <div className="h-96 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />
      </div>
    );
  }

  if (notFound || !data) {
    return (
      <div className="page-shell py-10">
        <div className="mx-auto max-w-md rounded-md border border-[#d7e3ed] bg-white p-8 text-center">
          <XCircle className="mx-auto h-10 w-10 text-[#c53030]" />
          <h1 className="mt-3 text-lg font-bold text-[#17324d]">E-tiket tidak ditemukan</h1>
          <p className="mt-1 text-sm text-[#698097]">Kode booking tidak valid atau e-tiket belum diterbitkan.</p>
          <Link href="/packages" className="mt-4 inline-flex h-10 items-center rounded-md bg-[#1768ad] px-5 text-sm font-semibold text-white hover:bg-[#14548f]">
            Pesan Tiket
          </Link>
        </div>
      </div>
    );
  }

  const { ticket, booking } = data;

  // Payload QR harus sama dengan qr_data yang tersimpan di backend agar hasil scan
  // petugas dapat diselesaikan menjadi tiket ini.
  const qrPayload = ticket.qr_data || ticket.id;

  const statusTone: Record<string, string> = {
    VALID: "bg-[#e6f7f0] text-[#16876a]",
    USED: "bg-[#e9f3fb] text-[#1768ad]",
    EXPIRED: "bg-[#fdf3e3] text-[#b7791f]",
    CANCELLED: "bg-[#fdeaea] text-[#c53030]",
  };

  return (
    <div className="page-shell py-10">
      {isNew && (
        <div className="mb-5 flex items-center gap-2 rounded-md border border-[#16876a] bg-[#f0fbf7] px-4 py-3 text-sm text-[#16876a]">
          <CheckCircle2 className="h-5 w-5" />
          Pembayaran berhasil — e-tiket Anda telah diterbitkan!
        </div>
      )}

      <div className="mb-7 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="font-heading text-2xl font-bold text-[#17324d]">E-Tiket Kawah Putih</h1>
          <p className="mt-1 max-w-xl text-sm text-[#698097]">
            Tiket kunjunganmu siap digunakan. Simpan e-tiket ini dan tunjukkan saat tiba di lokasi.
          </p>
        </div>
        <button
          onClick={handleDownload}
          disabled={downloading}
          className="inline-flex h-10 items-center gap-2 rounded-md bg-[#1768ad] px-5 text-sm font-semibold text-white hover:bg-[#14548f] disabled:opacity-60"
        >
          {downloading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
          Download PDF
        </button>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <section className="rounded-md border border-[#d7e3ed] bg-white p-5">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#e5edf3] pb-4">
            <div>
              <p className="text-xs text-[#698097]">Kode Booking</p>
              <p className="text-lg font-bold tracking-wide text-[#17324d]">{booking.booking_code}</p>
            </div>
            <span className={`rounded-full px-3 py-1 text-xs font-semibold ${statusTone[ticket.status] ?? "bg-[#e9f3fb] text-[#1768ad]"}`}>
              {ticket.status === "VALID" ? "Valid" : ticket.status === "USED" ? "Digunakan" : ticket.status === "EXPIRED" ? "Kadaluarsa" : "Dibatalkan"}
            </span>
          </div>

          <dl className="mt-4 grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-xs text-[#698097]">Nama Pemesan</dt>
              <dd className="text-sm font-semibold text-[#17324d]">{booking.visitor_name}</dd>
            </div>
            <div>
              <dt className="text-xs text-[#698097]">Email</dt>
              <dd className="text-sm font-semibold text-[#17324d]">{booking.email}</dd>
            </div>
            <div>
              <dt className="text-xs text-[#698097]">Tanggal Kunjungan</dt>
              <dd className="text-sm font-semibold text-[#17324d]">{booking.visit_date}</dd>
            </div>
            <div>
              <dt className="text-xs text-[#698097]">Jenis Tiket</dt>
              <dd className="text-sm font-semibold text-[#17324d]">{ticket.ticket_name} × {ticket.quantity}</dd>
            </div>
            <div>
              <dt className="text-xs text-[#698097]">Total Pembayaran</dt>
              <dd className="text-sm font-semibold text-[#17324d]">{rupiah(booking.total)}</dd>
            </div>
            <div>
              <dt className="text-xs text-[#698097]">Diterbitkan</dt>
              <dd className="text-sm font-semibold text-[#17324d]">{new Date(ticket.generated_at).toLocaleString("id-ID")}</dd>
            </div>
          </dl>

          <div className="mt-5 border-t border-[#e5edf3] pt-4">
            <p className="text-xs font-semibold text-[#29445e]">Rincian Tiket</p>
            <ul className="mt-2 space-y-2 text-sm">
              {booking.items.map((v) => (
                <li key={v.ticket_id} className="flex justify-between">
                  <span className="text-[#698097]">
                    {v.ticket_name} × {v.quantity}
                  </span>
                  <span className="font-medium tabular-nums text-[#17324d]">{rupiah(v.price * v.quantity)}</span>
                </li>
              ))}
            </ul>
          </div>
        </section>

        <div className="space-y-4">
          <section className="rounded-md border border-[#d7e3ed] bg-white p-5 text-center">
            <h2 className="text-sm font-bold text-[#17324d]">QR Code Tiket</h2>
            <div className="mx-auto mt-3 w-fit rounded-md border border-[#e5edf3] bg-white p-3">
              <QRCode value={qrPayload} size={160} aria-label="QR code e-tiket" />
            </div>
            <p className="mt-2 font-mono text-xs font-bold tracking-wider text-[#17324d]">{ticket.qr_data || ticket.id}</p>
            <p className="mt-1 text-[11px] text-[#698097]">Tunjukkan QR ini kepada petugas gerbang.</p>
          </section>

          <section className="rounded-md border border-[#fdf3e3] bg-[#fdf3e3] p-5">
            <h2 className="flex items-center gap-2 text-sm font-bold text-[#b7791f]">
              <Info className="h-4 w-4" /> Informasi Penting
            </h2>
            <p className="mt-2 text-sm text-[#8a6420]">
              Tiket hanya berlaku pada tanggal kunjungan yang tertera. Tiket yang sudah digunakan tidak dapat dipakai kembali.
            </p>
          </section>

          <button
            onClick={() => window.print()}
            className="flex h-10 w-full items-center justify-center gap-2 rounded-md border border-[#d7e3ed] bg-white px-4 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
          >
            <Printer className="h-4 w-4 text-[#1768ad]" /> Cetak E-Tiket
          </button>

          <Link href="/dashboard/booking" className="block text-center text-xs font-semibold text-[#1768ad] hover:underline">
            Lihat semua tiket saya
          </Link>
        </div>
      </div>
    </div>
  );
}