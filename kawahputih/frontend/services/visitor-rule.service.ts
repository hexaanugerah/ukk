import type { VisitorRule, CreateVisitorRulePayload, UpdateVisitorRulePayload } from "@/types/visitor-rule";
import { axiosInstance as api } from "@/lib/axios";

const BASE = "/admin/visitor-rules";

export const visitorRuleService = {
  async list(params?: { page?: number; limit?: number; category?: string; published_only?: boolean }): Promise<VisitorRule[]> {
    const { data } = await api.get(BASE, { params });
    return data.data;
  },

  async get(id: string): Promise<VisitorRule> {
    const { data } = await api.get(`${BASE}/${id}`);
    return data.data;
  },

  async listPublished(category?: string): Promise<VisitorRule[]> {
    const { data } = await api.get("/visitor-rules", { params: { category } });
    return data.data;
  },

  async create(payload: CreateVisitorRulePayload): Promise<VisitorRule> {
    const { data } = await api.post(BASE, payload);
    return data.data;
  },

  async update(id: string, payload: UpdateVisitorRulePayload): Promise<VisitorRule> {
    const { data } = await api.put(`${BASE}/${id}`, payload);
    return data.data;
  },

  async remove(id: string): Promise<void> {
    await api.delete(`${BASE}/${id}`);
  },
};