// Payment service — operasi pembayaran ke backend API.

import { apiClient, ApiError } from "@/lib/api/client";
import type {
  ApiResponse,
  Payment,
  ProcessPaymentRequest,
  ProcessPaymentResponse,
  PaymentsListResponse,
  PaymentStatus,
} from "@/types/api";

export class PaymentError extends Error {}

export interface PaymentQueryParams {
  page?: number;
  limit?: number;
  status?: string;
}

export const paymentService = {
  /** Get payment by booking ID. */
  async getByBookingId(bookingId: string): Promise<Payment | null> {
    try {
      const response = await apiClient.get<{ success: boolean; message?: string; data: Payment }>(`/payments/booking/${bookingId}`);
      return response.data ?? null;
    } catch {
      return null;
    }
  },

  /** Process payment (simulate QRIS). */
  async processPayment(payload: ProcessPaymentRequest): Promise<ProcessPaymentResponse> {
    // Backend memakai envelope standar: { success, message, data: { booking, payment, outcome } }.
    const response = await apiClient.post<ApiResponse<Pick<ProcessPaymentResponse, "booking" | "payment" | "outcome">>>(
      "/payments/process",
      payload
    );
    if (!response.success || !response.data) {
      throw new PaymentError(response.message || "Payment processing failed");
    }
    return {
      success: response.success,
      message: response.message,
      booking: response.data.booking,
      payment: response.data.payment,
      outcome: response.data.outcome,
    };
  },

  /** Expire a pending payment. */
  async expirePayment(bookingId: string): Promise<void> {
    const response = await apiClient.post<{ success: boolean; message?: string }>("/payments/expire?booking_id=" + encodeURIComponent(bookingId), {});
    if (!response.success) throw new PaymentError(response.message || "Failed to expire payment");
  },

  /** Get all payments (admin). */
  async list(params: PaymentQueryParams = {}): Promise<Payment[]> {
    const response = await apiClient.get<PaymentsListResponse>("/admin/payments", {
      page: String(params.page ?? 1),
      limit: String(params.limit ?? 10),
      status: params.status ?? "",
    });
    return response.data ?? [];
  },

  /** Update payment status (admin). */
  async updateStatus(id: string, status: PaymentStatus): Promise<Payment> {
    const response = await apiClient.patch<{ success: boolean; message?: string; data: Payment }>(`/admin/payments/${id}/status`, { status });
    if (!response.success) throw new PaymentError(response.message || "Failed to update payment status");
    return response.data!;
  },
};