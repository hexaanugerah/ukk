// Auth config — konfigurasi sesi & role aplikasi. TIDAK ADA kredensial di sini; akun sungguhan dibuat & divalidasi di backend. Frontend hanya menyimpan hasil login di store.

/** Role aplikasi — dipakai lintas modul (auth, guard, sidebar, redirect). */
export const APP_ROLES = ["PENGUNJUNG", "ADMIN", "MANAGER", "PETUGAS"] as const;
export type AppRole = (typeof APP_ROLES)[number];

/** Tautan "Dasbor" per role (navigasi manual dari navbar). */
export const ROLE_HOME: Record<AppRole, string> = {
  PENGUNJUNG: "/dashboard",
  ADMIN: "/admin/dashboard",
  MANAGER: "/manager/dashboard",
  PETUGAS: "/petugas/dashboard",
};

/**
 * Tujuan default SETELAH login (bukan tautan navbar). Pengunjung selalu
 * dikembalikan ke beranda — dasbor pengunjung hanya dibuka secara manual.
 * Jika ada ?redirect= di URL login, itu yang dipakai lebih dulu.
 */
export const POST_LOGIN_HOME: Record<AppRole, string> = {
  PENGUNJUNG: "/",
  ADMIN: "/admin/dashboard",
  MANAGER: "/manager/dashboard",
  PETUGAS: "/petugas/dashboard",
};

/** Durasi sesi (ms). */
export const SESSION_DURATION_MS = 1000 * 60 * 60 * 8; // 8 jam

/** Nama key localStorage sesi auth. */
export const AUTH_STORAGE_KEY = "kpr-auth";
export const AUTH_COOKIE_NAME = "kpr_access_token";
