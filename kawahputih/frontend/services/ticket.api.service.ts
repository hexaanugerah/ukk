// Ticket service — operasi tiket ke backend API.

import { apiClient, ApiError } from "@/lib/api/client";
import type { Ticket, TicketsListResponse, TicketStatus, TicketCategory } from "@/types/api";

export class TicketError extends Error {}

export interface TicketQueryParams {
  page?: number;
  per_page?: number;
  status?: string;
  category?: string;
}

export const ticketService = {
  /** Get all tickets with optional filtering. */
  async list(params: TicketQueryParams = {}): Promise<Ticket[]> {
    const response = await apiClient.get<TicketsListResponse>("/tickets", {
      page: String(params.page ?? 1),
      // Backend membatasi per_page 100; seed referensi saja sudah 21+ tiket.
      // Pakai nilai maksimal agar tiket baru (urut created_at ASC) tidak
      // terpotong ke "halaman 2" yang tidak pernah dimuat UI.
      per_page: String(params.per_page ?? 100),
      status: params.status ?? "",
      category: params.category ?? "",
    });
    return response.data ?? [];
  },

  /** Get active tickets for booking. */
  async getActive(): Promise<Ticket[]> {
    const response = await apiClient.get<TicketsListResponse>("/tickets/active");
    return response.data ?? [];
  },

  /** Get ticket by ID. */
  async getById(id: string): Promise<Ticket | null> {
    try {
      const response = await apiClient.get<{ success: boolean; message?: string; data: Ticket }>(`/tickets/${id}`);
      return response.data ?? null;
    } catch {
      return null;
    }
  },

  /** Get ticket by code. */
  async getByCode(code: string): Promise<Ticket | null> {
    try {
      const response = await apiClient.get<{ success: boolean; message?: string; data: Ticket }>(`/tickets/code/${code}`);
      return response.data ?? null;
    } catch {
      return null;
    }
  },

  /** Get ticket categories. */
  async getCategories(): Promise<string[]> {
    const response = await apiClient.get<{ success: boolean; message?: string; data: string[] }>("/tickets/categories");
    return response.data ?? [];
  },

  // Operasi admin
  async create(ticket: Omit<Ticket, "id" | "created_at" | "updated_at" | "available">): Promise<Ticket> {
    const response = await apiClient.post<{ success: boolean; message?: string; data: Ticket }>("/admin/tickets", ticket);
    if (!response.success) throw new TicketError(response.message || "Failed to create ticket");
    return response.data!;
  },

  async update(id: string, ticket: Partial<Ticket>): Promise<Ticket> {
    const response = await apiClient.put<{ success: boolean; message?: string; data: Ticket }>(`/admin/tickets/${id}`, ticket);
    if (!response.success) throw new TicketError(response.message || "Failed to update ticket");
    return response.data!;
  },

  async delete(id: string): Promise<void> {
    const response = await apiClient.delete<{ success: boolean; message?: string }>(`/admin/tickets/${id}`);
    if (!response.success) throw new TicketError(response.message || "Failed to delete ticket");
  },

  async updateStatus(id: string, status: TicketStatus): Promise<Ticket> {
    const response = await apiClient.patch<{ success: boolean; message?: string; data: Ticket }>(`/admin/tickets/${id}/status`, { status });
    if (!response.success) throw new TicketError(response.message || "Failed to update ticket status");
    return response.data!;
  },
};