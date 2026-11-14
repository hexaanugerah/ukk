// Payment service — operasi pembayaran ke backend API.

import { apiClient, ApiError } from "@/lib/api/client";
import type {
  ApiResponse,
  ManagerPaymentReport,
  Payment,
  PaymentDestination,
  PaymentDestinationPayload,
  PaymentMethodConfig,
  PaymentSettings,
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
  method?: string;
  from?: string;
  to?: string;
  q?: string;
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

  /** Proses / cek status pembayaran di backend. */
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
      method: params.method ?? "",
      from: params.from ?? "",
      to: params.to ?? "",
      q: params.q ?? "",
    });
    return response.data ?? [];
  },

  /** Unggah bukti pembayaran (multipart) untuk sebuah booking. */
  async uploadProof(bookingId: string, file: File): Promise<Payment> {
    const form = new FormData();
    form.append("proof", file);
    const response = await apiClient.postForm<ApiResponse<Payment>>(
      `/payments/booking/${bookingId}/proof`,
      form,
      60000
    );
    if (!response.success || !response.data) {
      throw new PaymentError(response.message || "Gagal mengunggah bukti pembayaran");
    }
    return response.data;
  },

  /** Konfigurasi payment gateway (admin). */
  async getSettings(): Promise<PaymentSettings> {
    const response = await apiClient.get<ApiResponse<PaymentSettings>>("/admin/payment-settings");
    if (!response.success || !response.data) throw new PaymentError(response.message || "Failed to load settings");
    return response.data;
  },

  /** Ubah konfigurasi gateway (admin). */
  async updateSettings(payload: Partial<Pick<PaymentSettings, "enabled" | "merchant_name" | "settlement_bank" | "settlement_account_masked">>): Promise<PaymentSettings> {
    const response = await apiClient.put<ApiResponse<PaymentSettings>>("/admin/payment-settings", payload);
    if (!response.success || !response.data) throw new PaymentError(response.message || "Failed to update settings");
    return response.data;
  },

  /** Seluruh metode pembayaran (admin). */
  async listMethods(): Promise<PaymentMethodConfig[]> {
    const response = await apiClient.get<ApiResponse<PaymentMethodConfig[]>>("/admin/payment-methods");
    if (!response.success) throw new PaymentError(response.message || "Failed to load methods");
    return response.data ?? [];
  },

  /** Metode pembayaran aktif (publik). */
  async listActiveMethods(): Promise<PaymentMethodConfig[]> {
    try {
      const response = await apiClient.get<ApiResponse<PaymentMethodConfig[]>>("/payment-methods");
      return response.data ?? [];
    } catch {
      return [];
    }
  },

  /** Aktif/nonaktifkan metode pembayaran (admin). */
  async updateMethod(code: string, payload: { enabled?: boolean; display_order?: number }): Promise<PaymentMethodConfig> {
    const response = await apiClient.patch<ApiResponse<PaymentMethodConfig>>(
      `/admin/payment-methods/${encodeURIComponent(code)}`,
      payload
    );
    if (!response.success || !response.data) throw new PaymentError(response.message || "Failed to update method");
    return response.data;
  },

  /** Laporan pembayaran manager. */
  async getManagerReport(range: { from?: string; to?: string } = {}): Promise<ManagerPaymentReport> {
    const response = await apiClient.get<ApiResponse<ManagerPaymentReport>>("/manager/payment-reports", {
      from: range.from ?? "",
      to: range.to ?? "",
    });
    if (!response.success || !response.data) throw new PaymentError(response.message || "Failed to load report");
    return response.data;
  },

  /** Update payment status (admin). */
  async updateStatus(id: string, status: PaymentStatus): Promise<Payment> {
    const response = await apiClient.patch<{ success: boolean; message?: string; data: Payment }>(`/admin/payments/${id}/status`, { status });
    if (!response.success) throw new PaymentError(response.message || "Failed to update payment status");
    return response.data!;
  },

  /**
   * Daftar tujuan pembayaran aktif (endpoint publik) — rekening bank,
   * nomor e-wallet, atau kode retail yang diatur oleh admin.
   */
  async listDestinations(): Promise<PaymentDestination[]> {
    try {
      const response = await apiClient.get<ApiResponse<PaymentDestination[]>>("/payment-destinations");
      return response.data ?? [];
    } catch {
      return [];
    }
  },

  /** Daftar seluruh tujuan pembayaran (admin). */
  async listDestinationsAdmin(): Promise<PaymentDestination[]> {
    const response = await apiClient.get<ApiResponse<PaymentDestination[]>>("/admin/payment-destinations");
    if (!response.success) throw new PaymentError(response.message || "Failed to load destinations");
    return response.data ?? [];
  },

  /** Tambah tujuan pembayaran (admin). */
  async createDestination(payload: PaymentDestinationPayload): Promise<PaymentDestination> {
    const response = await apiClient.post<ApiResponse<PaymentDestination>>("/admin/payment-destinations", payload);
    if (!response.success || !response.data) throw new PaymentError(response.message || "Failed to create destination");
    return response.data;
  },

  /** Ubah tujuan pembayaran (admin). */
  async updateDestination(id: string, payload: PaymentDestinationPayload): Promise<PaymentDestination> {
    const response = await apiClient.put<ApiResponse<PaymentDestination>>(`/admin/payment-destinations/${id}`, payload);
    if (!response.success || !response.data) throw new PaymentError(response.message || "Failed to update destination");
    return response.data;
  },

  /** Hapus tujuan pembayaran (admin). */
  async deleteDestination(id: string): Promise<void> {
    const response = await apiClient.delete<{ success: boolean; message?: string }>(`/admin/payment-destinations/${id}`);
    if (!response.success) throw new PaymentError(response.message || "Failed to delete destination");
  },
};