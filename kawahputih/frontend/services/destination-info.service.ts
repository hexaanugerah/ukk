import type { DestinationInfo, CreateDestinationInfoPayload, UpdateDestinationInfoPayload } from "@/types/destination-info";
import { axiosInstance as api } from "@/lib/axios";

const BASE = "/admin/destination-info";

export const destinationInfoService = {
  async list(params?: { page?: number; limit?: number; status?: string }): Promise<DestinationInfo[]> {
    const { data } = await api.get(BASE, { params });
    return data.data;
  },

  async get(id: string): Promise<DestinationInfo> {
    const { data } = await api.get(`${BASE}/${id}`);
    return data.data;
  },

  async getPublished(): Promise<DestinationInfo | null> {
    const { data } = await api.get("/destination-info");
    return data.data;
  },

  async create(payload: CreateDestinationInfoPayload): Promise<DestinationInfo> {
    const { data } = await api.post(BASE, payload);
    return data.data;
  },

  async update(id: string, payload: UpdateDestinationInfoPayload): Promise<DestinationInfo> {
    const { data } = await api.put(`${BASE}/${id}`, payload);
    return data.data;
  },

  async publish(id: string): Promise<void> {
    await api.post(`${BASE}/${id}/publish`);
  },

  async archive(id: string): Promise<void> {
    await api.post(`${BASE}/${id}/archive`);
  },

  async remove(id: string): Promise<void> {
    await api.delete(`${BASE}/${id}`);
  },
};