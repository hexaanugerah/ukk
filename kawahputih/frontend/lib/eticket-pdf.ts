"use client";

import { jsPDF } from "jspdf";
import QRCode from "qrcode";
import { rupiah } from "@/lib/business";
import type { Booking, ETicket } from "@/types/domain";

// E-ticket PDF — generate PDF e-tiket di sisi klien (jsPDF + QR PNG). Dipanggil dari tombol Download PDF.

export async function downloadETicketPDF(ticket: ETicket, booking: Booking) {
  const doc = new jsPDF({ unit: "pt", format: "a4", orientation: "portrait" });
  const W = doc.internal.pageSize.getWidth();

  // Header brand
  doc.setFillColor(23, 104, 173); // #1768ad
  doc.rect(0, 0, W, 90, "F");
  doc.setTextColor(255, 255, 255);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(20);
  doc.text("KAWAH PUTIH RANCABALI", 40, 42);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(11);
  doc.text("E-Tiket Kunjungan — Jl. Raya Soreang-Ciwidey KM.25, Bandung", 40, 62);

  // Judul & status
  let y = 130;
  doc.setTextColor(23, 50, 77); // #17324d
  doc.setFont("helvetica", "bold");
  doc.setFontSize(16);
  doc.text("E-TIKET", 40, y);

  doc.setFontSize(10);
  const statusColor: Record<string, [number, number, number]> = {
    VALID: [22, 135, 106],
    USED: [23, 104, 173],
    EXPIRED: [181, 76, 26],
    CANCELLED: [197, 48, 48],
  };
  const [r, g, b] = statusColor[ticket.status] ?? [23, 50, 77];
  doc.setTextColor(r, g, b);
  doc.text(`STATUS: ${ticket.status}`, W - 40, y, { align: "right" });

  // QR Code
  const qrPayload = JSON.stringify({
    t: ticket.id,
    b: ticket.bookingId,
    v: ticket.visitDate,
    n: ticket.quantity,
    s: ticket.status,
  });
  const qrDataUrl = await QRCode.toDataURL(qrPayload, { width: 360, margin: 1 });
  const qrSize = 150;
  const qrX = W - 40 - qrSize;
  doc.addImage(qrDataUrl, "PNG", qrX, y + 16, qrSize, qrSize);

  // Info booking
  doc.setTextColor(23, 50, 77);
  const rows: [string, string][] = [
    ["Kode Booking", booking.id],
    ["Kode Tiket", ticket.id],
    ["Nama Pemesan", booking.visitorName],
    ["Email", booking.email],
    ["Tanggal Kunjungan", booking.visitDate],
    ["Jenis Tiket", ticket.ticketName],
    ["Jumlah Tiket", `${ticket.quantity} tiket`],
    ["Total Pembayaran", rupiah(booking.total)],
    ["Metode Pembayaran", "QRIS"],
    ["Dibuat", new Date(ticket.generatedAt).toLocaleString("id-ID")],
  ];

  y += 20;
  doc.setFontSize(11);
  rows.forEach(([label, value], i) => {
    const ry = y + i * 22;
    doc.setFont("helvetica", "normal");
    doc.setTextColor(105, 128, 151);
    doc.text(label, 40, ry);
    doc.setFont("helvetica", "bold");
    doc.setTextColor(23, 50, 77);
    doc.text(String(value).slice(0, 55), 190, ry);
  });

  // Catatan
  const noteY = y + rows.length * 22 + 30;
  doc.setDrawColor(215, 227, 237);
  doc.roundedRect(40, noteY, W - 80 - qrSize - 20, 70, 6, 6, "S");
  doc.setFont("helvetica", "normal");
  doc.setFontSize(9);
  doc.setTextColor(105, 128, 151);
  doc.text("Tunjukkan e-tiket ini (cetak atau dari layar) saat tiba di gerbang.", 52, noteY + 20);
  doc.text("Tiket hanya berlaku pada tanggal kunjungan yang tertera.", 52, noteY + 36);
  doc.text("Tiket yang sudah digunakan tidak dapat dipakai kembali.", 52, noteY + 52);

  // Footer
  doc.setFontSize(8);
  doc.setTextColor(148, 163, 184);
  doc.text("Dokumen ini dihasilkan otomatis oleh sistem Kawah Putih Rancabali.", 40, 800);

  doc.save(`e-tiket-${booking.id}.pdf`);
}
