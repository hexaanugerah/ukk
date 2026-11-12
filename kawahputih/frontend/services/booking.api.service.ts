// Booking service — operasi booking ke backend API.

import { apiClient, ApiError } from "@/lib/api/client";
import type {
  Booking,
  CreateBookingRequest,
  CreateBookingResponse,
  BookingsListResponse,
  BookingStatus,
  ApiResponse,
} from "@/types/api";

export class BookingError extends Error {}

export interface BookingQueryParams {
  page?: number;
  limit?: number;
  status?: string;
  search?: string;
}

export const bookingService = {
  /** Get all bookings (admin). */
  async list(params: BookingQueryParams = {}): Promise<Booking[]> {
    const response = await apiClient.get<BookingsListResponse>("/admin/bookings", {
      page: String(params.page ?? 1),
      limit: String(params.limit ?? 10),
      status: params.status ?? "",
      search: params.search ?? "",
    });
    return response.data ?? [];
  },

  /** Riwayat booking + pembayaran milik seorang pengunjung (admin/manager). */
  async listForVisitor(visitorId: string, params: BookingQueryParams = {}): Promise<Booking[]> {
    const response = await apiClient.get<BookingsListResponse>(`/admin/visitors/${visitorId}/bookings`, {
      page: String(params.page ?? 1),
      limit: String(params.limit ?? 100),
      status: params.status ?? "",
    });
    return response.data ?? [];
  },

  /** Get today's bookings (admin). */
  async listToday(): Promise<Booking[]> {
    const response = await apiClient.get<BookingsListResponse>("/admin/bookings/today");
    return response.data ?? [];
  },

  /** Get bookings for current user. */
  async listMine(params: BookingQueryParams = {}): Promise<Booking[]> {
    const response = await apiClient.get<BookingsListResponse>("/bookings/mine", {
      page: String(params.page ?? 1),
      limit: String(params.limit ?? 10),
    });
    return response.data ?? [];
  },

  /** Get booking by ID. */
  async getById(id: string): Promise<Booking | null> {
    try {
      const response = await apiClient.get<ApiResponse<Booking>>(`/bookings/${id}`);
      return response.data ?? null;
    } catch {
      return null;
    }
  },

  /** Create a new booking. */
  async create(payload: CreateBookingRequest): Promise<CreateBookingResponse> {
    // Backend memakai envelope standar: { success, message, data: { booking, payment } }.
    // booking/payment ada di dalam `data`, bukan di root respons.
    const response = await apiClient.post<ApiResponse<Pick<CreateBookingResponse, "booking" | "payment">>>(
      "/bookings",
      payload
    );
    if (!response.success || !response.data) {
      throw new BookingError(response.message || "Failed to create booking");
    }
    return {
      success: response.success,
      message: response.message,
      booking: response.data.booking,
      payment: response.data.payment,
    };
  },

  /** Cancel a booking. */
  async cancel(id: string, reason?: string): Promise<Booking> {
    const response = await apiClient.delete<{ success: boolean; message?: string; data: Booking }>(`/bookings/${id}`);
    if (!response.success) throw new BookingError(response.message || "Failed to cancel booking");
    return response.data!;
  },

  /** Update booking status (admin). */
  async updateStatus(id: string, status: BookingStatus): Promise<Booking> {
    const response = await apiClient.patch<{ success: boolean; message?: string; data: Booking }>(`/admin/bookings/${id}/status`, { status });
    if (!response.success) throw new BookingError(response.message || "Failed to update booking status");
    return response.data!;
  },
};