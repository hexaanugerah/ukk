import type { GalleryItem } from "@/types/gallery";
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

  /** Tambah foto ke galeri (admin only) — upload file multipart. */
  async create(payload: FormData): Promise<GalleryItem> {
    const { data } = await api.post(BASE, payload, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return data.data;
  },

  /** Perbarui foto galeri (admin only) — upload file multipart (gambar opsional). */
  async update(id: string, payload: FormData): Promise<GalleryItem> {
    const { data } = await api.put(`${BASE}/${id}`, payload, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return data.data;
  },

  /** Hapus foto galeri (admin only). */
  async remove(id: string): Promise<void> {
    await api.delete(`${BASE}/${id}`);
  },
};
