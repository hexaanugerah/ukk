import { apiClient } from "@/lib/api/client";
import type { ApiMeta } from "@/lib/api";
import type { CreatePackagePayload, TourismPackage, UpdatePackagePayload } from "@/types/destination";

// Destination service — paket wisata dari backend API. Endpoint: GET /packages (publik), GET /packages/:id, POST/PUT/DELETE /packages (ADMIN).

type Envelope<T> = { success: boolean; message: string; data: T; meta?: ApiMeta };

export const destinationService = {
  /** Daftar paket wisata, opsional hanya yang aktif. */
  async list(params?: { page?: number; limit?: number; active_only?: boolean }): Promise<TourismPackage[]> {
    const response = await apiClient.get<Envelope<TourismPackage[]>>("/packages", {
      page: String(params?.page ?? 1),
      limit: String(params?.limit ?? 20),
      active_only: String(params?.active_only ?? false),
    });
    return response.data ?? [];
  },

  /** Detail paket berdasarkan id — null bila tidak ditemukan. */
  async get(id: string): Promise<TourismPackage | null> {
    try {
      const response = await apiClient.get<Envelope<TourismPackage>>(`/packages/${id}`);
      return response.data ?? null;
    } catch {
      return null;
    }
  },

  /** Buat paket baru (aktif secara default). */
  async create(payload: CreatePackagePayload): Promise<TourismPackage> {
    const response = await apiClient.post<Envelope<TourismPackage>>("/packages", payload);
    return response.data;
  },

  /** Perbarui paket (termasuk aktif/nonaktif). */
  async update(id: string, payload: UpdatePackagePayload): Promise<TourismPackage> {
    const response = await apiClient.put<Envelope<TourismPackage>>(`/packages/${id}`, payload);
    return response.data;
  },

  /** Hapus paket. */
  async remove(id: string): Promise<void> {
    await apiClient.delete<Envelope<null>>(`/packages/${id}`);
  },
};
