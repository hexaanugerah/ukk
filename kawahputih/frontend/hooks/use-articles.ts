import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { articleService } from "@/services/article.service";
import type { CreateArticlePayload } from "@/types/article";

export function useArticles(params?: { page?: number; limit?: number; status?: string }) {
  return useQuery({ queryKey: ["articles", params], queryFn: () => articleService.list(params) });
}

export function useArticle(id: string) {
  return useQuery({ queryKey: ["articles", id], queryFn: () => articleService.get(id), enabled: !!id });
}

export function useCreateArticle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateArticlePayload) => articleService.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["articles"] }),
  });
}

export function useUpdateArticle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: CreateArticlePayload }) => articleService.update(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["articles"] }),
  });
}

export function usePublishArticle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => articleService.publish(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["articles"] }),
  });
}

export function useDeleteArticle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => articleService.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["articles"] }),
  });
}
