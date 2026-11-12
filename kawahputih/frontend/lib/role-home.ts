// Helper kompatibilitas untuk kode lama. Role backend memakai huruf besar
// (PENGUNJUNG/ADMIN/MANAGER/PETUGAS), jadi pemetaan mengikuti nilai itu.
export function roleHome(roles: string[] = []) {
  if (roles.some((role) => ["ADMIN", "SUPER_ADMIN", "super_admin"].includes(role))) return "/admin/dashboard";
  if (roles.includes("MANAGER")) return "/manager/dashboard";
  if (roles.includes("PETUGAS")) return "/petugas/dashboard";
  return "/dashboard";
}
