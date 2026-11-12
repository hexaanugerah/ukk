// Staff service — operasi staf ke backend API.

import { apiClient, ApiError } from "@/lib/api/client";
import type { Staff } from "@/types/api";

export class StaffError extends Error {}

export const staffService = {
  /** Get all staff (admin). */
  async list(): Promise<Staff[]> {
    const response = await apiClient.get<{ success: boolean; message?: string; data: Staff[] }>("/admin/staff");
    return response.data ?? [];
  },

  /** Get staff by ID (admin). */
  async getById(id: string): Promise<Staff | null> {
    try {
      const response = await apiClient.get<{ success: boolean; message?: string; data: Staff }>(`/admin/staff/${id}`);
      return response.data ?? null;
    } catch {
      return null;
    }
  },

  /** Get staff stats (admin). */
  async getStats(): Promise<Record<string, number>> {
    const response = await apiClient.get<{ success: boolean; message?: string; data: Record<string, number> }>("/admin/staff/stats");
    return response.data ?? {};
  },

  /** Update staff status (admin). */
  async updateStatus(id: string, status: "AKTIF" | "NONAKTIF"): Promise<void> {
    const response = await apiClient.patch<{ success: boolean; message?: string }>(`/admin/staff/${id}/status`, { status });
    if (!response.success) throw new StaffError(response.message || "Failed to update staff status");
  },
};