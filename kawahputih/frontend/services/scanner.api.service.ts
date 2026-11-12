// Scanner service — operasi scanner ke backend API.

import { apiClient, ApiError } from "@/lib/api/client";
import type {
  ETicket,
  ManagerMonitoring,
  ScanRecord,
  ValidateTicketRequest,
  ValidateTicketResponse,
  ScanHistoryResponse,
  ScanStatsResponse,
  ScanStatus,
} from "@/types/api";

export type { ScanRecord };

export class ScannerError extends Error {}

export interface ScanHistoryParams {
  page?: number;
  limit?: number;
  gate?: string;
  petugas?: string;
  from?: string;
  to?: string;
  status?: string;
}

/** Read-only ticket lookup by ticket UUID, booking code, or QR payload. */
export async function lookupTicket(ref: string): Promise<ETicket | null> {
  try {
    const response = await apiClient.get<{ success: boolean; message?: string; data: ETicket }>(
      "/petugas/scanner/ticket",
      { ref },
    );
    return response.data ?? null;
  } catch {
    return null;
  }
}

export const scannerService = {
  /** Validate a ticket. */
  async validateTicket(payload: ValidateTicketRequest): Promise<ValidateTicketResponse> {
    const response = await apiClient.post<ValidateTicketResponse>("/petugas/scanner/validate", payload);
    return response;
  },

  /** Get scan history. */
  async getHistory(params: ScanHistoryParams = {}): Promise<ScanRecord[]> {
    const response = await apiClient.get<ScanHistoryResponse>("/petugas/scanner/history", {
      page: String(params.page ?? 1),
      limit: String(params.limit ?? 20),
      gate: params.gate ?? "",
      petugas: params.petugas ?? "",
      from: params.from ?? "",
      to: params.to ?? "",
    });
    return response.data ?? [];
  },

  /** Get scan stats. */
  async getStats(): Promise<Record<string, number>> {
    const response = await apiClient.get<ScanStatsResponse>("/petugas/scanner/stats");
    return response.data ?? {};
  },

  /** Agregat monitoring untuk manager (tanpa riwayat scan mentah). */
  async getMonitoring(date?: string): Promise<ManagerMonitoring> {
    const response = await apiClient.get<{ success: boolean; message?: string; data: ManagerMonitoring }>(
      "/manager/monitoring",
      date ? { date } : undefined
    );
    if (!response.success) throw new ScannerError(response.message || "Gagal memuat monitoring");
    return response.data!;
  },
};