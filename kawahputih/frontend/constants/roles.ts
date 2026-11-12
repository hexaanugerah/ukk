// Mencerminkan seed role di backend (database/migrations/002_role.up.sql). String ini harus sama dengan yang dikembalikan API pada array `roles` pengguna.
export const ROLES = {
  ADMIN: "admin",
  SUPER_ADMIN: "super_admin",
  OWNER: "owner",
  FINANCE_ADMIN: "finance_admin",
  STAFF_TICKETING: "staff_ticketing",
  STAFF_CONTENT: "staff_content",
  VISITOR: "visitor",
} as const;

export type RoleName = (typeof ROLES)[keyof typeof ROLES];
