// Tipe domain — tipe data terpusat untuk seluruh modul.

export type TicketStatus = "AKTIF" | "NONAKTIF";

/** Jenis tiket yang dijual (dikonfigurasi, bukan hardcoded di UI). */
export interface Ticket {
  id: string;
  name: string;
  description: string;
  price: number;
  capacity: number;
  status: TicketStatus;
  createdAt: string;
}

export type BookingStatus = "PENDING" | "PAID" | "CONFIRMED" | "USED" | "CANCELLED" | "EXPIRED";

export interface BookingItem {
  ticketId: string;
  ticketName: string;
  price: number;
  quantity: number;
}

export interface Booking {
  id: string; // KP-YYYYMMDD-XXXXXX
  userId: string | null; // null = guest (belum login)
  visitorName: string;
  email: string;
  phone: string;
  visitDate: string; // ISO yyyy-MM-dd
  items: BookingItem[];
  subtotal: number;
  serviceFee: number;
  discount: number;
  tax: number;
  total: number;
  /** Total kuantitas semua item — dihitung saat create & ikut tersimpan. */
  totalTickets: number;
  status: BookingStatus;
  note?: string;
  createdAt: string; // ISO datetime
  ticketId: string | null; // tiket e-ticket utama
  checkedInAt?: string;
}

export type PaymentStatus = "PENDING" | "PROCESSING" | "PAID" | "FAILED" | "EXPIRED" | "REFUNDED";

export interface Payment {
  id: string; // PAY-YYYYMMDD-XXXXXX
  bookingId: string;
  method: "QRIS";
  amount: number;
  status: PaymentStatus;
  createdAt: string;
  paidAt?: string;
  expiredAt: string;
}

export type TicketValidityStatus = "VALID" | "USED" | "EXPIRED" | "CANCELLED";

/** E-tiket hasil booking — berisi data yang di-encode ke QR. */
export interface ETicket {
  id: string; // TK-XXXXXX
  bookingId: string;
  visitorName: string;
  email: string;
  visitDate: string;
  ticketName: string;
  quantity: number;
  total: number;
  status: TicketValidityStatus;
  generatedAt: string;
  checkedInAt?: string;
  checkedInGate?: string;
}

export type StaffRole = "PETUGAS GERBANG" | "PETUGAS TIKET" | "SUPERVISOR" | "ADMIN";
export type StaffStatus = "AKTIF" | "NONAKTIF";
export type StaffShift = "PAGI" | "SIANG" | "MALAM";

export interface Staff {
  id: string;
  name: string;
  email: string;
  phone: string;
  role: StaffRole;
  shift: StaffShift;
  gate: string;
  status: StaffStatus;
  lastLogin: string | null;
}

/** Profil pengunjung — diturunkan dari booking (denormalisasi mock). */
export interface Visitor {
  id: string;
  name: string;
  email: string;
  phone: string;
  joinedAt: string;
}

export interface ScanRecord {
  id: string;
  ticketId: string;
  bookingId: string;
  visitorName: string;
  timestamp: string;
  gate: string;
  petugas: string;
  status: "VALID" | "SUDAH DIGUNAKAN" | "INVALID" | "EXPIRED" | "CANCELLED";
}

/** Akun pengguna lokal (pendaftaran pengunjung baru, mode mock). */
export interface LocalAccount {
  id: string;
  name: string;
  email: string;
  password: string;
  phone: string;
  role: "PENGUNJUNG";
  createdAt: string;
}

// Statistik & analytics

export interface DashboardStatistics {
  totalBookings: number;
  totalTickets: number; // tiket terjual (jumlah kuantitas)
  revenue: number;
  visitors: number;
  bookingsToday: number;
  revenueToday: number;
  ticketsSold: number;
  ticketsAvailable: number;
}

export interface AnalyticsPoint {
  label: string;
  value: number;
}

export interface AnalyticsSummary {
  totalRevenue: number;
  revenueToday: number;
  totalBookings: number;
  activeStaff: number;
  totalVisitors: number;
  avgBookingValue: number;
  conversionRate: number;
}

export interface AnalyticsData {
  summary: AnalyticsSummary;
  revenueTrend: AnalyticsPoint[];
  bookingTrend: AnalyticsPoint[];
  weekdayVsWeekend: AnalyticsPoint[];
  ticketTypePerformance: AnalyticsPoint[];
  paymentStatusBreakdown: AnalyticsPoint[];
  revenueChangePct: number;
  bookingChangePct: number;
  visitorsChangePct: number;
}

export type DateRangePreset = "today" | "7d" | "30d" | "month" | "custom";

export interface AppNotification {
  id: string;
  title: string;
  message: string;
  type: "booking" | "payment" | "ticket" | "staff" | "system";
  status: "Berhasil" | "Menunggu" | "Dibatalkan" | "Gagal";
  createdAt: string;
  read: boolean;
}
