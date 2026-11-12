import { create } from "zustand";

// Booking store — membawa draft pemesanan melewati alur multi-halaman: Booking -> Checkout/Payment -> Processing -> Success -> E-Tiket. Bersifat sementara (tidak dipersist).

export interface BookingDraft {
  visitDate: string;
  items: { ticketId: string; ticketName: string; price: number; quantity: number }[];
  visitorName: string;
  email: string;
  phone: string;
  note?: string;
  totals: { subtotal: number; serviceFee: number; discount: number; tax: number; total: number; totalTickets: number };
}

interface BookingState {
  draft: BookingDraft | null;
  /** Kode booking terakhir yang diproses (untuk halaman payment/success/e-tiket). */
  lastBookingId: string | null;
  setDraft: (draft: BookingDraft) => void;
  setLastBookingId: (id: string | null) => void;
  clear: () => void;
}

export const useBookingStore = create<BookingState>((set) => ({
  draft: null,
  lastBookingId: null,
  setDraft: (draft) => set({ draft }),
  setLastBookingId: (id) => set({ lastBookingId: id }),
  clear: () => set({ draft: null, lastBookingId: null }),
}));
