"use client";

import { jsPDF } from "jspdf";
import * as XLSX from "xlsx";
import { rupiah } from "@/lib/business";

// Ekspor laporan — ekspor laporan manager (PDF & Excel) sepenuhnya di sisi klien.

export interface ReportData {
  title: string;
  dateRangeLabel: string;
  summary: { label: string; value: string }[];
  revenueSeries: { label: string; value: number }[];
  bookingSeries: { label: string; value: number }[];
  ticketPerformance: { label: string; value: number }[];
  paymentStatus: { label: string; value: number }[];
  /** Baris harian untuk Excel: Date, Booking, Tickets Sold, Revenue, Visitors, Payment Status */
  dailyRows: { date: string; bookings: number; tickets: number; revenue: number; visitors: number; paymentStatus: string }[];
}

export function exportReportPDF(data: ReportData) {
  const doc = new jsPDF({ unit: "pt", format: "a4" });
  const W = doc.internal.pageSize.getWidth();

  doc.setFillColor(23, 104, 173);
  doc.rect(0, 0, W, 70, "F");
  doc.setTextColor(255, 255, 255);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(16);
  doc.text("KAWAH PUTIH RANCABALI", 40, 32);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(10);
  doc.text("Laporan Analitik Manajemen", 40, 50);

  let y = 100;
  doc.setTextColor(23, 50, 77);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(14);
  doc.text(data.title, 40, y);
  y += 18;
  doc.setFont("helvetica", "normal");
  doc.setFontSize(10);
  doc.setTextColor(105, 128, 151);
  doc.text(`Periode: ${data.dateRangeLabel}`, 40, y);

  y += 26;
  doc.setFont("helvetica", "bold");
  doc.setFontSize(11);
  doc.setTextColor(23, 50, 77);
  doc.text("Ringkasan", 40, y);
  y += 16;
  doc.setFontSize(10);
  data.summary.forEach((s, i) => {
    const ry = y + i * 18;
    doc.setFont("helvetica", "normal");
    doc.setTextColor(105, 128, 151);
    doc.text(s.label, 40, ry);
    doc.setFont("helvetica", "bold");
    doc.setTextColor(23, 50, 77);
    doc.text(s.value, 240, ry);
  });

  y += data.summary.length * 18 + 20;
  doc.setFont("helvetica", "bold");
  doc.setFontSize(11);
  doc.setTextColor(23, 50, 77);
  doc.text("Tren Pendapatan", 40, y);
  y += 14;
  drawBarChartPDF(doc, data.revenueSeries, 40, y, W - 80, 110);

  y += 140;
  doc.setFont("helvetica", "bold");
  doc.text("Statistik Booking", 40, y);
  y += 14;
  drawBarChartPDF(doc, data.bookingSeries, 40, y, W - 80, 110);

  y += 140;
  doc.setFont("helvetica", "bold");
  doc.text("Kinerja Jenis Tiket", 40, y);
  y += 16;
  data.ticketPerformance.forEach((t, i) => {
    const ry = y + i * 16;
    doc.setFont("helvetica", "normal");
    doc.setTextColor(105, 128, 151);
    doc.text(`${t.label}`, 40, ry);
    doc.setTextColor(23, 50, 77);
    doc.text(`${t.value} tiket`, 240, ry);
  });

  doc.setFontSize(8);
  doc.setTextColor(148, 163, 184);
  doc.text("Laporan dihasilkan otomatis oleh sistem Kawah Putih.", 40, 800);

  doc.save(`laporan-manager-${new Date().toISOString().slice(0, 10)}.pdf`);
}

function drawBarChartPDF(doc: jsPDF, series: { label: string; value: number }[], x: number, y: number, w: number, h: number) {
  if (series.length === 0) return;
  const max = Math.max(...series.map((s) => s.value), 1);
  const gap = 8;
  const bw = (w - gap * (series.length - 1)) / series.length;
  doc.setFillColor(23, 104, 173);
  series.forEach((s, i) => {
    const bh = Math.max(4, (s.value / max) * h);
    const bx = x + i * (bw + gap);
    doc.rect(bx, y + h - bh, bw, bh, "F");
    doc.setFontSize(7);
    doc.setTextColor(105, 128, 151);
    doc.text(s.label, bx + bw / 2, y + h + 10, { align: "center" });
  });
}

export function exportReportExcel(data: ReportData) {
  const wb = XLSX.utils.book_new();

  // Sheet 1: Ringkasan
  const summaryRows = [
    ["Laporan", data.title],
    ["Periode", data.dateRangeLabel],
    ["Dihasilkan", new Date().toLocaleString("id-ID")],
    [],
    ["Metrik", "Nilai"],
    ...data.summary.map((s) => [s.label, s.value]),
  ];
  const ws1 = XLSX.utils.aoa_to_sheet(summaryRows);
  XLSX.utils.book_append_sheet(wb, ws1, "Ringkasan");

  // Sheet 2: Data harian
  const dailyRows = [
    ["Tanggal", "Booking", "Tiket Terjual", "Pendapatan", "Pengunjung", "Status Pembayaran"],
    ...data.dailyRows.map((r) => [r.date, r.bookings, r.tickets, r.revenue, r.visitors, r.paymentStatus]),
  ];
  const ws2 = XLSX.utils.aoa_to_sheet(dailyRows);
  XLSX.utils.book_append_sheet(wb, ws2, "Data Harian");

  XLSX.writeFile(wb, `laporan-manager-${new Date().toISOString().slice(0, 10)}.xlsx`);
}
