// Role service — manajemen penetapan role untuk super admin.
// Endpoint: GET /super-admin/roles, PUT /super-admin/users/:userId/roles,
// GET /admin/users (daftar pengguna beserta role saat ini).

import { apiClient, ApiError } from "@/lib/api/client";
import type { ApiResponse } from "@/types/api";

export class RoleError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "RoleError";
  }
}

export interface RoleItem {
  id: string;
  name: string;
  description: string;
}

/** Pandangan domain user atas tabel users (AdminUserResponse backend). */
export interface AdminUserItem {
  id: string;
  name: string;
  email: string;
  phone: string;
  is_active: boolean;
  is_verified: boolean;
  roles: string[];
}

export const roleService = {
  /** Daftar role kanonik (termasuk super_admin) — hanya untuk super admin. */
  async listRoles(): Promise<RoleItem[]> {
    const response = await apiClient.get<ApiResponse<RoleItem[]>>("/super-admin/roles");
    return response.data ?? [];
  },

  /** Daftar semua pengguna beserta role aktifnya (halaman penuh hingga 100/halaman). */
  async listUsers(page = 1, limit = 100): Promise<AdminUserItem[]> {
    const response = await apiClient.get<ApiResponse<AdminUserItem[]>>("/admin/users", {
      page: String(page),
      limit: String(limit),
    });
    return response.data ?? [];
  },

  /** Ganti seluruh role seorang pengguna (replace, bukan tambah). */
  async assignRoles(userId: string, roles: string[]): Promise<void> {
    try {
      const response = await apiClient.put<ApiResponse<null>>(
        `/super-admin/users/${userId}/roles`,
        { roles }
      );
      if (!response.success) throw new RoleError(response.message || "Gagal menyimpan role.");
    } catch (err) {
      throw new RoleError(
        err instanceof ApiError ? err.message : "Gagal menyimpan role."
      );
    }
  },
};
