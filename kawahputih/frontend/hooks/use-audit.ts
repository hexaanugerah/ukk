import { useQuery } from "@tanstack/react-query";
import { auditService, type AuditFeed, type AuditQueryParams } from "@/services/audit.api.service";

// useAuditFeed — feed aktivitas admin dari backend (GET /admin/audit).

export function useAuditFeed(params: AuditQueryParams = {}) {
  return useQuery<AuditFeed>({
    queryKey: [
      "admin",
      "audit",
      params.page ?? 1,
      params.limit ?? 20,
      params.action ?? "",
      params.target_type ?? "",
      params.status ?? "",
      params.from ?? "",
      params.to ?? "",
    ],
    queryFn: () => auditService.list(params),
    staleTime: 10_000,
  });
}
