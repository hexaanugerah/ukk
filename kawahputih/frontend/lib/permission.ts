// Pengecekan role di sisi klien hanya untuk UX (menyembunyikan/menampilkan
// tombol); penegakan sebenarnya ada di backend lewat middleware RequireRole.
// Jangan andalkan hasil fungsi ini untuk hal sensitif keamanan, ini hanya
// agar UI tidak menampilkan tombol yang toh akan ditolak API.
export function hasRole(userRoles: string[], ...allowed: string[]): boolean {
  return userRoles.some((r) => allowed.includes(r));
}

export function isAdmin(userRoles: string[]): boolean {
  return hasRole(userRoles, "admin", "super_admin");
}
