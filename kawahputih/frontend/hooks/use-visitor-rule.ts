import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { visitorRuleService } from "@/services/visitor-rule.service";
import type { CreateVisitorRulePayload, UpdateVisitorRulePayload } from "@/types/visitor-rule";

export function useVisitorRuleList(params?: { page?: number; limit?: number; category?: string; published_only?: boolean }) {
  return useQuery({
    queryKey: ["visitor-rule", params],
    queryFn: () => visitorRuleService.list(params),
  });
}

export function useVisitorRule(id: string) {
  return useQuery({
    queryKey: ["visitor-rule", id],
    queryFn: () => visitorRuleService.get(id),
    enabled: !!id,
  });
}

export function usePublishedVisitorRuleList(category?: string) {
  return useQuery({
    queryKey: ["visitor-rule", "published", category],
    queryFn: () => visitorRuleService.listPublished(category),
  });
}

export function useCreateVisitorRule() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateVisitorRulePayload) => visitorRuleService.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["visitor-rule"] }),
  });
}

export function useUpdateVisitorRule() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateVisitorRulePayload }) => visitorRuleService.update(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["visitor-rule"] }),
  });
}

export function useDeleteVisitorRule() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => visitorRuleService.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["visitor-rule"] }),
  });
}