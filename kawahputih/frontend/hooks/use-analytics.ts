import { useQuery } from "@tanstack/react-query";
import { analyticsService } from "@/services/analytics.api.service";
import type { ManagerAnalytics } from "@/types/api";

export type DateRangePreset = "today" | "7d" | "30d" | "month" | "custom";

// useManagerAnalytics — analytics manager dari backend API.

export function useManagerAnalytics(range: DateRangePreset = "7d", custom?: { from: string; to: string }) {
  return useQuery<ManagerAnalytics>({
    queryKey: ["manager", "analytics", range, custom?.from, custom?.to],
    queryFn: () =>
      analyticsService.getManagerAnalytics(
        range === "custom" ? { range: "custom", custom } : { range },
      ),
    staleTime: 5_000,
  });
}
