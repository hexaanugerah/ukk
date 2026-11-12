"use client";

import { useCallback } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore, type SessionUser } from "@/store/auth.store";
import { authService, AuthError } from "@/services/auth.api.service";
import type { AppRole } from "@/config/auth.config";

// Auth hooks — jembatan komponen ke auth service. Komponen tidak memanggil service langsung.

export function useUser(): SessionUser | null {
  return useAuthStore((s) => s.user);
}

export function useIsAuthenticated(): boolean {
  return useAuthStore((s) => s.user !== null);
}

export function useHasRole(): (...roles: AppRole[]) => boolean {
  const role = useAuthStore((s) => s.user?.role ?? null);
  return useCallback((...roles: AppRole[]) => role !== null && roles.includes(role), [role]);
}

interface LoginCredentials {
  email: string;
  password: string;
  remember: boolean;
  /** Halaman tujuan setelah login (mis. `/booking?date=...` dari alur pemesanan). */
  redirectTo?: string;
}

/**
 * Baca ?redirect= dari URL login. Hanya menerima path internal (diawali "/"
 * dan bukan "//") supaya tidak bisa dipakai untuk open-redirect ke situs lain.
 */
function readRedirectParam(): string | null {
  if (typeof window === "undefined") return null;
  const value = new URLSearchParams(window.location.search).get("redirect");
  if (!value || !value.startsWith("/") || value.startsWith("//")) return null;
  return value;
}

export function useLogin(options?: { redirectTo?: string }) {
  const router = useRouter();
  return useCallback(
    async (creds: LoginCredentials): Promise<{ ok: true; redirectTo: string } | { ok: false; error: string }> => {
      try {
        const result = await authService.login(creds.email, creds.password, creds.remember);
        // Prioritas: tujuan eksplisit pemanggil → ?redirect= di URL → default role.
        // Inilah yang membuat login dari halaman tiket kembali ke halaman itu,
        // bukan mendarat di dasbor pengunjung.
        const target = creds.redirectTo ?? options?.redirectTo ?? readRedirectParam() ?? result.redirectTo;
        router.push(target);
        return { ok: true, redirectTo: target };
      } catch (err) {
        const message = err instanceof AuthError ? err.message : "Terjadi kesalahan saat login.";
        return { ok: false, error: message };
      }
    },
    [router, options?.redirectTo]
  );
}

export function useRegister() {
  const router = useRouter();
  return useCallback(
    async (payload: { name: string; email: string; phone: string; password: string }) => {
      try {
        const result = await authService.register(payload);
        router.push(result.redirectTo);
        return { ok: true as const };
      } catch (err) {
        const message = err instanceof AuthError ? err.message : "Registrasi gagal.";
        return { ok: false as const, error: message };
      }
    },
    [router]
  );
}

export function useChangePassword() {
  const router = useRouter();
  return useCallback(
    async (
      currentPassword: string,
      newPassword: string
    ): Promise<{ ok: true } | { ok: false; error: string }> => {
      try {
        await authService.changePassword(currentPassword, newPassword);
        // Backend mencabut semua refresh token; pengguna wajib login ulang.
        router.push("/login");
        return { ok: true };
      } catch (err) {
        const message = err instanceof AuthError ? err.message : "Gagal mengganti password.";
        return { ok: false, error: message };
      }
    },
    [router]
  );
}

export function useLogout() {
  const router = useRouter();
  return useCallback(() => {
    authService.logout();
    router.push("/login");
  }, [router]);
}

// Re-export untuk kompatibilitas dengan kode lama yang mengimpor useAuthStore.
export { useAuthStore };
export type { SessionUser };
