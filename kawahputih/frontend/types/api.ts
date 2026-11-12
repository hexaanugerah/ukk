// Tipe API — tipe TypeScript yang cocok dengan respons backend.

export type BookingStatus = "PENDING" | "PAID" | "CONFIRMED" | "USED" | "CANCELLED" | "EXPIRED";

export type PaymentStatus = "PENDING" | "PROCESSING" | "PAID" | "FAILED" | "EXPIRED" | "REFUNDED";

export type TicketStatus = "AKTIF" | "NONAKTIF";

export type TicketCategory = "ENTRANCE" | "TRANSPORT" | "SUNRISE" | "PACKAGE" | "ACTIVITY" | "SERVICE";

export type ETicketStatus = "VALID" | "USED" | "CANCELLED" | "EXPIRED";

export type ScanStatus = "VALID" | "USED" | "INVALID" | "EXPIRED" | "CANCELLED";

export type UserRole = "ADMIN" | "MANAGER" | "PETUGAS" | "PENGUNJUNG";

// Envelope respons standar backend
export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
  meta?: ApiMeta;
  errors?: any;
}

export interface ApiMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  gate?: string;
  shift?: string;
  phone?: string;
}

// Tipe auth
export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthData {
  access_token: string;
  refresh_token: string;
  expires_at: number;
  user: User;
}

export interface LoginResponse {
  success: boolean;
  message?: string;
  data: AuthData;
}

export interface LogoutResponse {
  success: boolean;
  message: string;
}

export interface MeResponse {
  success: boolean;
  message?: string;
  data: User;
}

// Tipe tiket
export interface Ticket {
  id: string;
  code: string;
  name: string;
  description: string;
  category: TicketCategory;
  price: number;
  status: TicketStatus;
  capacity: number;
  available: number;
  is_active: boolean;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface TicketsListResponse extends ApiResponse<Ticket[]> {}

// Tipe booking
export interface BookingItem {
  ticket_id: string;
  ticket_name: string;
  price: number;
  quantity: number;
}

export interface BookingTotals {
  subtotal: number;
  discount: number;
  service_fee: number;
  tax: number;
  total: number;
  total_tickets: number;
}

export interface Booking {
  id: string;
  /** Human-readable booking reference (e.g. KP-20261014-71b571). `id` is the
   *  UUID used for API/navigation; `booking_code` is what users should see. */
  booking_code: string;
  user_id?: string;
  visitor_name: string;
  email: string;
  phone: string;
  visit_date: string;
  items: BookingItem[];
  subtotal: number;
  discount: number;
  service_fee: number;
  tax: number;
  total: number;
  total_tickets: number;
  status: BookingStatus;
  note?: string;
  created_at: string;
  updated_at: string;
  ticket_id?: string;
  checked_in_at?: string;	checked_in_by?: string;
	checked_in_gate?: string;
	payment?: Payment;
}

export interface CreateBookingRequest {
  visit_date: string;
  visitor: {
    name: string;
    email: string;
    phone: string;
  };
  items: Array<{
    ticket_id: string;
    quantity: number;
  }>;
  note?: string;
}

export interface CreateBookingResponse {
  success: boolean;
  message?: string;
  booking: Booking;
  payment: Payment;
}

export interface BookingsListResponse extends ApiResponse<Booking[]> {}

export interface UpdateBookingStatusRequest {
  status: BookingStatus;
}

// Tipe pembayaran
export interface Payment {
  id: string;
  booking_id: string;
  /** Kode metode pembayaran, mis. "QRIS", "GOPAY", "BCA_VA". */
  method: string;
  amount: number;
  status: PaymentStatus;
  created_at: string;
  paid_at?: string;
  expired_at: string;
}

export interface ProcessPaymentRequest {
  booking_id: string;
  /** Metode yang dipilih pengunjung; kosong = pertahankan metode yang ada. */
  method?: string;
  forced?: "SUCCESS" | "FAILED" | "EXPIRED";
}

export interface ProcessPaymentResponse {
  success: boolean;
  message?: string;
  booking: Booking;
  payment: Payment;
  outcome: "SUCCESS" | "FAILED" | "EXPIRED";
}

export interface PaymentsListResponse extends ApiResponse<Payment[]> {}

// Tipe e-tiket
export interface ETicket {
  id: string;
  booking_id: string;
  booking_code: string;
  visitor_name: string;
  email: string;
  visit_date: string;
  ticket_name: string;
  quantity: number;
  total: number;
  status: ETicketStatus;
  generated_at: string;
  checked_in_at?: string;
  checked_in_gate?: string;
  qr_data: string;
}

// Tipe pengunjung
export interface Visitor {
  id: string;
  name: string;
  email: string;
  phone: string;
  joined_at: string;
  total_bookings?: number;
  total_spent?: number;
  last_visit?: string;
}

export interface VisitorsListResponse extends ApiResponse<Visitor[]> {}
export interface VisitorStatsResponse extends ApiResponse<Record<string, number>> {}

// Tipe staf
export interface Staff {
  id: string;
  name: string;
  email: string;
  phone: string;
  role: string;
  shift: string;
  gate: string;
  status: "AKTIF" | "NONAKTIF";
  last_login?: string;
}

export interface StaffListResponse extends ApiResponse<Staff[]> {}
export interface StaffStatsResponse extends ApiResponse<Record<string, number>> {}

// Tipe scanner
export interface ScanRecord {
  id: string;
  ticket_id: string;
  booking_id: string;
  visitor_name: string;
  timestamp: string;
  gate: string;
  petugas: string;
  status: ScanStatus;
}

export interface ValidateTicketRequest {
  ticket_id: string;
  gate?: string;
  petugas?: string;
}

export interface ValidateTicketResponse {
  success: boolean;
  message?: string;
  status: ScanStatus;
  ticket?: ETicket;
}

export interface ScanHistoryResponse extends ApiResponse<ScanRecord[]> {}
export interface ScanStatsResponse extends ApiResponse<Record<string, number>> {}

export interface MonitoringPetugasRow {
  name: string;
  gate: string;
  total: number;
  valid: number;
  last_activity?: string;
}

export interface MonitoringHourlyRow {
  hour: number;
  count: number;
}

export interface ManagerMonitoring {
  totals: Record<string, number>;
  total_valid: number;
  valid_today: number;
  per_petugas: MonitoringPetugasRow[];
  hourly: MonitoringHourlyRow[];
}

// Tipe analytics
export interface AdminDashboardStats {
  total_bookings: number;
  total_tickets_sold: number;
  revenue: number;
  total_visitors: number;
  visitors_today: number;
  bookings_today: number;
  revenue_today: number;
  tickets_available: number;
  pending_payments: number;
  active_staff: number;
}

export interface TrendPoint {
  label: string;
  value: number;
}

export interface DailyRow {
  date: string;
  bookings: number;
  tickets: number;
  revenue: number;
  visitors: number;
  payment_status: string;
}

export interface ManagerAnalytics {
  summary: {
    total_revenue: number;
    revenue_today: number;
    total_bookings: number;
    active_staff: number;
    total_tickets: number;
    total_visitors: number;
    avg_booking_value: number;
    conversion_rate: number;
  };
  changes: {
    revenue: number;
    bookings: number;
    visitors: number;
  };
  revenue_trend: TrendPoint[];
  booking_trend: TrendPoint[];
  weekday_vs_weekend: TrendPoint[];
  ticket_type_performance: TrendPoint[];
  payment_status: TrendPoint[];
  daily_rows: DailyRow[];
  range_label: string;
}