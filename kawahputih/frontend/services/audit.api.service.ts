// Audit service — feed aktivitas nyata dari backend (GET /admin/audit). Menggantikan mock DB sebagai sumber notifikasi admin.

import { apiClient } from "@/lib/api/client";
import type { ApiMeta } from "@/lib/api";

export interface AuditLog {
  id: string;
  actor_id: string;
  actor_role: string;
  action: string;
  target_type: string;
  target_id: string;
  status: string;
  metadata?: Record<string, unknown>;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}

export interface AuditQueryParams {
  page?: number;
  limit?: number;
  action?: string;
  target_type?: string;
  status?: string;
  from?: string;
  to?: string;
}

export interface AuditFeed {
  items: AuditLog[];
  meta?: ApiMeta;
}

export const auditService = {
  /** Get paginated audit logs (admin/manager). */
  async list(params: AuditQueryParams = {}): Promise<AuditFeed> {
    const response = await apiClient.get<{ success: boolean; message?: string; data: AuditLog[]; meta?: ApiMeta }>(
      "/admin/audit",
      {
        page: String(params.page ?? 1),
        limit: String(params.limit ?? 20),
        action: params.action ?? "",
        target_type: params.target_type ?? "",
        status: params.status ?? "",
        from: params.from ?? "",
        to: params.to ?? "",
      }
    );
    return { items: response.data ?? [], meta: response.meta };
  },
};

/** Human label for a raw audit action string (e.g. "booking.create"). */
export function auditActionLabel(action: string): string {
  const map: Record<string, string> = {
    "auth.login": "Login pengguna",
    "auth.logout": "Logout pengguna",
    "auth.register": "Pendaftaran akun",
    "booking.create": "Booking baru",
    "booking.cancel": "Booking dibatalkan",
    "booking.status_update": "Status booking diubah",
    "payment.process": "Pembayaran diproses",
    "payment.expire": "Pembayaran kedaluwarsa",
    "payment.status_update": "Status pembayaran diubah",
    "ticket.validate": "Validasi tiket",
    "ticket.issue": "E-tiket diterbitkan",
    "ticket.lookup": "Pencarian tiket",
    "ticket.sync_status": "Sinkronisasi status tiket",
    "tickettype.create": "Jenis tiket ditambah",
    "tickettype.update": "Jenis tiket diubah",
    "tickettype.delete": "Jenis tiket dihapus",
  };
  return map[action] ?? action;
}
