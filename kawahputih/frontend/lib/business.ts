import type { BookingItem } from "@/types/domain";
import { BOOKING_CONFIG } from "@/config/app.config";

// Business logic — utilitas perhitungan terpusat. Tidak ada komponen yang menghitung total/harga/paginasi sendiri; semua lewat fungsi di sini agar konsisten.

/** Format angka ke Rupiah. */
export function rupiah(value: number): string {
  return `Rp ${Math.round(value).toLocaleString("id-ID")}`;
}

/** Tanggal hari ini dalam format ISO yyyy-MM-dd (zona lokal). */
export function todayISO(): string {
  const d = new Date();
  const off = d.getTimezoneOffset();
  return new Date(d.getTime() - off * 60_000).toISOString().slice(0, 10);
}

export function addDaysISO(iso: string, days: number): string {
  const d = new Date(`${iso}T00:00:00`);
  d.setDate(d.getDate() + days);
  return d.toISOString().slice(0, 10);
}

/** Selisih hari (integer) antara dua tanggal ISO. */
export function daysBetween(a: string, b: string): number {
  const da = new Date(`${a}T00:00:00`).getTime();
  const db = new Date(`${b}T00:00:00`).getTime();
  return Math.round((db - da) / 86_400_000);
}

// Generator ID

function randomSeq(len: number): string {
  let s = "";
  for (let i = 0; i < len; i++) s += Math.floor(Math.random() * 10);
  return s;
}

/** KP-20261004-000123 */
export function generateBookingId(dateISO: string = todayISO()): string {
  return `KP-${dateISO.replace(/-/g, "")}-${randomSeq(6)}`;
}

/** PAY-20261004-000123 */
export function generatePaymentId(dateISO: string = todayISO()): string {
  return `PAY-${dateISO.replace(/-/g, "")}-${randomSeq(6)}`;
}

/** TK-7F3K9Q */
export function generateTicketId(): string {
  const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789";
  let s = "";
  for (let i = 0; i < 6; i++) s += chars[Math.floor(Math.random() * chars.length)];
  return `TK-${s}`;
}

// Perhitungan harga — satu-satunya tempat total dipesan dihitung

export interface BookingTotals {
  subtotal: number;
  serviceFee: number;
  discount: number;
  tax: number;
  total: number;
  totalTickets: number;
}

export function calculateBookingTotals(
  items: Pick<BookingItem, "price" | "quantity">[],
  overrides?: { discountPercent?: number; taxPercent?: number; serviceFee?: number }
): BookingTotals {
  const subtotal = items.reduce((sum, it) => sum + it.price * it.quantity, 0);
  const totalTickets = items.reduce((sum, it) => sum + it.quantity, 0);
  const serviceFee = totalTickets > 0 ? (overrides?.serviceFee ?? BOOKING_CONFIG.serviceFee) : 0;
  const discount = Math.round((subtotal * (overrides?.discountPercent ?? BOOKING_CONFIG.discountPercent)) / 100);
  const tax = Math.round(((subtotal - discount) * (overrides?.taxPercent ?? BOOKING_CONFIG.taxPercent)) / 100);
  const total = Math.max(0, subtotal - discount + serviceFee + tax);
  return { subtotal, serviceFee, discount, tax, total, totalTickets };
}

// Statistik

/** Persentase perubahan dari periode sebelumnya, dibulatkan 1 desimal. */
export function calculatePercentageChange(current: number, previous: number): number {
  if (previous === 0) return current > 0 ? 100 : 0;
  return Math.round(((current - previous) / previous) * 1000) / 10;
}

// Table engine — search / filter / sort / pagination yang bisa dipakai ulang

export interface PaginateOptions {
  page: number;
  pageSize: number;
}

export function paginate<T>(rows: T[], page: number, pageSize: number) {
  const total = rows.length;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const safePage = Math.min(Math.max(1, page), totalPages);
  const start = (safePage - 1) * pageSize;
  return { rows: rows.slice(start, start + pageSize), total, totalPages, page: safePage };
}

export function searchIn<T>(rows: T[], needle: string, keys: (keyof T)[]): T[] {
  const n = needle.trim().toLowerCase();
  if (!n) return rows;
  return rows.filter((row) => keys.some((k) => String(row[k] ?? "").toLowerCase().includes(n)));
}

export function sortData<T>(rows: T[], key: keyof T | string, dir: "asc" | "desc" = "asc"): T[] {
  const copy = [...rows];
  copy.sort((a, b) => {
    const av = (a as Record<string, unknown>)[key as string];
    const bv = (b as Record<string, unknown>)[key as string];
    if (typeof av === "number" && typeof bv === "number") return dir === "asc" ? av - bv : bv - av;
    const as = String(av ?? "").toLowerCase();
    const bs = String(bv ?? "").toLowerCase();
    return dir === "asc" ? as.localeCompare(bs) : bs.localeCompare(as);
  });
  return copy;
}

/** Ambil angka acak dengan distribusi (untuk seed data & simulasi). */
export function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

export function pick<T>(arr: readonly T[]): T {
  const item = arr[Math.floor(Math.random() * arr.length)];
  return item as T;
}
