import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { faqService } from "@/services/faq.service";
import type { CreateFAQPayload, UpdateFAQPayload } from "@/types/faq";

export function useFAQList(params?: { page?: number; limit?: number; category?: string; published_only?: boolean }) {
  return useQuery({
    queryKey: ["faq", params],
    queryFn: () => faqService.list(params),
  });
}

export function useFAQ(id: string) {
  return useQuery({
    queryKey: ["faq", id],
    queryFn: () => faqService.get(id),
    enabled: !!id,
  });
}

export function usePublishedFAQList(category?: string) {
  return useQuery({
    queryKey: ["faq", "published", category],
    queryFn: () => faqService.listPublished(category),
  });
}

export function useCreateFAQ() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateFAQPayload) => faqService.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["faq"] }),
  });
}

export function useUpdateFAQ() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateFAQPayload }) => faqService.update(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["faq"] }),
  });
}

export function useDeleteFAQ() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => faqService.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["faq"] }),
  });
}