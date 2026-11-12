export interface Booking {
  id: string;
  user_id: string;
  package_id: string;
  visit_date: string;
  quantity: number;
  total_cents: number;
  currency: string;
  status: "pending_payment" | "paid" | "cancelled" | "expired" | "checked_in";
  ticket_code: string;
  checked_in_at?: string;
}

export interface CreateBookingPayload {
  package_id: string;
  visit_date: string;
  quantity: number;
  customer_name: string;
  customer_email: string;
}

export interface CreateBookingResult {
  booking_id: string;
  redirect_url: string;
  status: string;
}
