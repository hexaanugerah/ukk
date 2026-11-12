import type { Facility, CreateFacilityPayload, UpdateFacilityPayload } from "@/types/facility";
import { axiosInstance as api } from "@/lib/axios";

const BASE = "/admin/facilities";

export const facilityService = {
  async list(params?: { page?: number; limit?: number; status?: string; published_only?: boolean }): Promise<Facility[]> {
    const { data } = await api.get(BASE, { params });
    return data.data;
  },

  async get(id: string): Promise<Facility> {
    const { data } = await api.get(`${BASE}/${id}`);
    return data.data;
  },

  async listPublished(): Promise<Facility[]> {
    const { data } = await api.get("/facilities");
    return data.data;
  },

  async create(payload: CreateFacilityPayload): Promise<Facility> {
    const { data } = await api.post(BASE, payload);
    return data.data;
  },

  async update(id: string, payload: UpdateFacilityPayload): Promise<Facility> {
    const { data } = await api.put(`${BASE}/${id}`, payload);
    return data.data;
  },

  async remove(id: string): Promise<void> {
    await api.delete(`${BASE}/${id}`);
  },
};