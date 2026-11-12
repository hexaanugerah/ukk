import type { CreateGalleryItemPayload, GalleryItem } from "@/types/gallery";
import { axiosInstance as api } from "@/lib/axios";

// Backend mendaftarkan endpoint admin galeri di path yang sama dengan
// endpoint publik (POST/PUT/DELETE /gallery aman karena cek JWT + RequireRole).
const BASE = "/gallery";

export const galleryService = {
  /** Daftar foto galeri (public endpoint). */
  async list(params?: { page?: number; limit?: number; category?: string }): Promise<GalleryItem[]> {
    const { data } = await api.get(BASE, { params });
    return data.data;
  },

  /** Detail foto galeri berdasarkan ID. */
  async get(id: string): Promise<GalleryItem> {
    const { data } = await api.get(`${BASE}/${id}`);
    return data.data;
  },

  /** Tambah foto ke galeri (admin only). */
  async create(payload: CreateGalleryItemPayload): Promise<GalleryItem> {
    const { data } = await api.post(BASE, payload);
    return data.data;
  },

  /** Perbarui foto galeri (admin only). */
  async update(id: string, payload: Partial<CreateGalleryItemPayload>): Promise<GalleryItem> {
    const { data } = await api.put(`${BASE}/${id}`, payload);
    return data.data;
  },

  /** Hapus foto galeri (admin only). */
  async remove(id: string): Promise<void> {
    await api.delete(`${BASE}/${id}`);
  },
};