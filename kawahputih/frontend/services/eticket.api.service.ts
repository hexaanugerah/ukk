// E-ticket service — operasi e-tiket ke backend API.

import { apiClient, ApiError } from "@/lib/api/client";
import { BookingError } from "@/services/booking.api.service";
import type { ETicket, Booking } from "@/types/api";

export interface ETicketResponse {
  success: boolean;
  message?: string;
  data?: { ticket: ETicket; booking: Booking };
}

export interface MarkDownloadedResponse {
  success: boolean;
  message?: string;
}

export const eTicketService = {
  /** Get all e-tickets owned by the current visitor. */
  async listMine(): Promise<ETicket[]> {
    const response = await apiClient.get<{ success: boolean; message?: string; data: ETicket[] }>("/e-tickets/mine");
    return response.data ?? [];
  },

  /** Get e-ticket by booking ID. */
  async getByBookingId(bookingId: string): Promise<{ ticket: ETicket; booking: Booking } | null> {
    try {
      const response = await apiClient.get<ETicketResponse>(`/payments/booking/${bookingId}`);
      return response.data ?? null;
    } catch {
      return null;
    }
  },

  /** Mark e-ticket as downloaded. */
  async markDownloaded(bookingId: string): Promise<void> {
    const response = await apiClient.post<MarkDownloadedResponse>(`/payments/booking/${bookingId}/download`, {});
    if (!response.success) throw new BookingError(response.message || "Failed to mark as downloaded");
  },
};