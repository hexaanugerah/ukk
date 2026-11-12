// Analytics service — operasi analytics ke backend API.

import { apiClient, ApiError } from "@/lib/api/client";
import type {
  AdminDashboardStats,
  ManagerAnalytics,
} from "@/types/api";

export class AnalyticsError extends Error {}

export type AnalyticsRange = "today" | "7d" | "30d" | "month" | "custom";
export type DateRangePreset = AnalyticsRange;

export interface AnalyticsParams {
  range?: AnalyticsRange;
  custom?: { from: string; to: string };
}

export const analyticsService = {
  /** Get admin dashboard stats. */
  async getAdminDashboard(): Promise<AdminDashboardStats> {
    const response = await apiClient.get<{ success: boolean; message?: string; data: AdminDashboardStats }>("/admin/dashboard");
    if (!response.success) throw new AnalyticsError(response.message || "Failed to get dashboard stats");
    return response.data!;
  },

  /** Get manager analytics. */
  async getManagerAnalytics(params: AnalyticsParams = {}): Promise<ManagerAnalytics> {
    const queryParams: Record<string, string> = {
      range: params.range ?? "7d",
    };
    if (params.custom) {
      queryParams.custom = JSON.stringify(params.custom);
    }

    const response = await apiClient.get<{ success: boolean; message?: string; data: ManagerAnalytics }>("/manager/analytics", queryParams);
    if (!response.success) throw new AnalyticsError(response.message || "Failed to get analytics");
    return response.data!;
  },

  /** Get manager dashboard (alias for analytics with default range). */
  async getManagerDashboard(): Promise<ManagerAnalytics> {
    return this.getManagerAnalytics({ range: "7d" });
  },
};