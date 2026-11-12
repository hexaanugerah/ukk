// Visitor service — operasi pengunjung ke backend API.

import { apiClient, ApiError } from "@/lib/api/client";
import type { Visitor } from "@/types/api";

export class VisitorError extends Error {}

export const visitorService = {
  /** Get all visitors with stats (admin). */
  async list(): Promise<Visitor[]> {
    const response = await apiClient.get<{ success: boolean; message?: string; data: Visitor[] }>("/admin/visitors");
    return response.data ?? [];
  },

  /** Get visitor by ID (admin). */
  async getById(id: string): Promise<Visitor | null> {
    try {
      const response = await apiClient.get<{ success: boolean; message?: string; data: Visitor }>(`/admin/visitors/${id}`);
      return response.data ?? null;
    } catch {
      return null;
    }
  },

  /** Get visitor stats (admin). */
  async getStats(): Promise<Record<string, number>> {
    const response = await apiClient.get<{ success: boolean; message?: string; data: Record<string, number> }>("/admin/visitors/stats");
    return response.data ?? {};
  },
};