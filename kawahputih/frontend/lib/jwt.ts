// AuthResponse backend tidak menyertakan field `roles` pada objek user,
// tetapi payload JWT access token menyertakannya. Mendekode di sisi klien
// aman karena hanya dipakai untuk tampilan/UX (menu mana yang ditampilkan).
// Backend tetap memverifikasi signature dan role token di setiap panggilan
// API, jadi token palsu hanya menampilkan menu salah, bukan memberi akses.
export function decodeRolesFromToken(token: string): string[] {
  try {
    const payload = token.split(".")[1];
    if (!payload) return [];
    const decoded = JSON.parse(atob(payload.replace(/-/g, "+").replace(/_/g, "/")));
    return Array.isArray(decoded.roles) ? decoded.roles : [];
  } catch {
    return [];
  }
}
