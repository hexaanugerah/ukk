import { apiClient } from "@/lib/api/client";
import type { ApiMeta } from "@/lib/api";
import type { Article, CreateArticlePayload } from "@/types/article";

// Article service — artikel dari backend API. Endpoint: GET /articles (publik), GET /articles/:id, POST/PUT/DELETE /articles (ADMIN), POST /articles/:id/publish, /archive.

type Envelope<T> = { success: boolean; message: string; data: T; meta?: ApiMeta };

export const articleService = {
  /** Daftar artikel, opsional difilter status & dibatasi jumlahnya. */
  async list(params?: { page?: number; limit?: number; status?: string }): Promise<Article[]> {
    const response = await apiClient.get<Envelope<Article[]>>("/articles", {
      page: String(params?.page ?? 1),
      limit: String(params?.limit ?? 20),
      status: params?.status ?? "",
    });
    return response.data ?? [];
  },

  /** Detail artikel berdasarkan id — null bila tidak ditemukan. */
  async get(id: string): Promise<Article | null> {
    try {
      const response = await apiClient.get<Envelope<Article>>(`/articles/${id}`);
      return response.data ?? null;
    } catch {
      return null;
    }
  },

  /** Buat artikel baru (status draft). */
  async create(payload: CreateArticlePayload): Promise<Article> {
    const response = await apiClient.post<Envelope<Article>>("/articles", payload);
    return response.data;
  },

  /** Perbarui isi artikel. */
  async update(id: string, payload: CreateArticlePayload): Promise<Article> {
    const response = await apiClient.put<Envelope<Article>>(`/articles/${id}`, payload);
    return response.data;
  },

  /** Terbitkan artikel draft. */
  async publish(id: string): Promise<void> {
    await apiClient.post<Envelope<null>>(`/articles/${id}/publish`, {});
  },

  /** Arsipkan artikel. */
  async archive(id: string): Promise<void> {
    await apiClient.post<Envelope<null>>(`/articles/${id}/archive`, {});
  },

  /** Hapus artikel. */
  async remove(id: string): Promise<void> {
    await apiClient.delete<Envelope<null>>(`/articles/${id}`);
  },
};
