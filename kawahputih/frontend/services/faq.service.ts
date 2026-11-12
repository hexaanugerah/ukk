import type { FAQ, CreateFAQPayload, UpdateFAQPayload } from "@/types/faq";
import { axiosInstance as api } from "@/lib/axios";

const BASE = "/admin/faq";

export const faqService = {
  async list(params?: { page?: number; limit?: number; category?: string; published_only?: boolean }): Promise<FAQ[]> {
    const { data } = await api.get(BASE, { params });
    return data.data;
  },

  async get(id: string): Promise<FAQ> {
    const { data } = await api.get(`${BASE}/${id}`);
    return data.data;
  },

  async listPublished(category?: string): Promise<FAQ[]> {
    const { data } = await api.get("/faq", { params: { category } });
    return data.data;
  },

  async create(payload: CreateFAQPayload): Promise<FAQ> {
    const { data } = await api.post(BASE, payload);
    return data.data;
  },

  async update(id: string, payload: UpdateFAQPayload): Promise<FAQ> {
    const { data } = await api.put(`${BASE}/${id}`, payload);
    return data.data;
  },

  async remove(id: string): Promise<void> {
    await api.delete(`${BASE}/${id}`);
  },
};