// Auth service — autentikasi ke backend API.

import { apiClient, ApiError } from "@/lib/api/client";
import { POST_LOGIN_HOME, type AppRole } from "@/config/auth.config";
import { useAuthStore, type SessionUser } from "@/store/auth.store";
import type {
  ApiResponse,
  LoginRequest,
  LoginResponse,
  LogoutResponse,
  MeResponse,
  User,
} from "@/types/api";

export class AuthError extends Error {
  constructor(message: string, public status?: number) {
    super(message);
    this.name = "AuthError";
  }
}

function mapBackendUserToSessionUser(user: User): SessionUser {
  return {
    id: user.id,
    name: user.name,
    email: user.email,
    phone: user.phone ?? "",
    role: user.role,
    gate: user.gate,
    shift: user.shift,
  };
}

export const authService = {
  /** Login dengan email/password ke backend API. */
  async login(email: string, password: string, remember: boolean): Promise<{
    user: SessionUser;
    redirectTo: string;
  }> {
    let response: LoginResponse;
    try {
      response = await apiClient.post<LoginResponse>("/auth/login", { email, password });
    } catch (err) {
      // Teruskan pesan backend (mis. "Email atau password tidak valid.") ke form.
      throw new AuthError(err instanceof ApiError ? err.message : "Terjadi kesalahan saat login.", err instanceof ApiError ? err.status : undefined);
    }

    // Backend mengembalikan envelope standar: { success, message, data: { access_token, refresh_token, expires_at, user } }
    if (!response.success || !response.data?.user || !response.data?.access_token) {
      throw new AuthError(response.message || "Login gagal");
    }

    // Simpan token di API client (access + refresh untuk auto-refresh)
    apiClient.setToken(response.data.access_token);
    apiClient.setRefreshToken(response.data.refresh_token ?? null);

    const user = mapBackendUserToSessionUser(response.data.user);
    useAuthStore.getState().login(user, remember);

    // Tujuan default setelah login: pengunjung ke beranda (dasbor pengunjung dibuka manual), staf ke panelnya. Halaman login menimpanya dengan ?redirect= bila pengguna datang dari suatu halaman.
    return {
      user,
      redirectTo: POST_LOGIN_HOME[user.role as AppRole] ?? "/",
    };
  },

  /** Registrasi pengunjung baru via backend API. */
  async register(payload: {
    name: string;
    email: string;
    phone: string;
    password: string;
  }): Promise<{
    user: SessionUser;
    redirectTo: string;
  }> {
    let response: LoginResponse;
    try {
      response = await apiClient.post<LoginResponse>("/auth/register", payload);
    } catch (err) {
      throw new AuthError(err instanceof ApiError ? err.message : "Terjadi kesalahan saat registrasi.", err instanceof ApiError ? err.status : undefined);
    }

    if (!response.success || !response.data?.user || !response.data?.access_token) {
      throw new AuthError(response.message || "Registrasi gagal");
    }

    apiClient.setToken(response.data.access_token);
    apiClient.setRefreshToken(response.data.refresh_token ?? null);
    const user = mapBackendUserToSessionUser(response.data.user);
    useAuthStore.getState().login(user, true);

    return { user, redirectTo: POST_LOGIN_HOME[user.role as AppRole] ?? "/" };
  },

  /** Update profil pengunjung via backend API. */
  async updateProfile(patch: {
    name?: string;
    email?: string;
    phone?: string;
  }): Promise<SessionUser> {
    const current = useAuthStore.getState().user;
    if (!current) throw new AuthError("Belum login.");

    const response = await apiClient.patch<ApiResponse<User>>("/auth/profile", patch);
    if (!response.success || !response.data) {
      throw new AuthError(response.message || "Gagal memperbarui profil.");
    }

    const user = mapBackendUserToSessionUser(response.data);
    useAuthStore.getState().updateUser(user);
    return user;
  },

  /**
   * Ganti password akun yang sedang login (berlaku untuk semua role).
   * Backend memverifikasi password lama, menyimpan hash bcrypt baru, lalu
   * mencabut seluruh refresh token — jadi sesi lokal juga harus dibersihkan.
   */
  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    try {
      await apiClient.post<ApiResponse<null>>("/auth/password/change", {
        current_password: currentPassword,
        new_password: newPassword,
      });
    } catch (err) {
      throw new AuthError(
        err instanceof ApiError ? err.message : "Gagal mengganti password.",
        err instanceof ApiError ? err.status : undefined
      );
    }

    // Hanya setelah backend sukses: seluruh sesi dicabut, jadi bersihkan sesi lokal.
    apiClient.clearTokens();
    useAuthStore.getState().logout();
  },

  /** Logout dari backend (mencabut refresh token aktif). */
  async logout(): Promise<void> {
    try {
      const refreshToken = apiClient.getRefreshToken();
      await apiClient.post<LogoutResponse>("/auth/logout", { refresh_token: refreshToken ?? "" });
    } catch {
      // Abaikan error logout
    } finally {
      apiClient.clearTokens();
      useAuthStore.getState().logout();
    }
  },

  /**
   * Lupa password — backend mengirim link reset ke email.
   * Respons selalu sukses (tidak membocorkan keberadaan email).
   */
  async forgotPassword(email: string): Promise<void> {
    try {
      const response = await apiClient.post<ApiResponse<null>>("/auth/password/forgot", { email });
      if (!response.success) throw new AuthError(response.message || "Gagal mengirim link reset.");
    } catch (err) {
      throw new AuthError(
        err instanceof ApiError ? err.message : "Gagal mengirim link reset.",
        err instanceof ApiError ? err.status : undefined
      );
    }
  },

  /** Reset password dengan token dari email (POST /auth/password/reset). */
  async resetPassword(token: string, newPassword: string): Promise<void> {
    try {
      const response = await apiClient.post<ApiResponse<null>>("/auth/password/reset", {
        token,
        new_password: newPassword,
      });
      if (!response.success) throw new AuthError(response.message || "Gagal mereset password.");
    } catch (err) {
      throw new AuthError(
        err instanceof ApiError ? err.message : "Gagal mereset password.",
        err instanceof ApiError ? err.status : undefined
      );
    }
  },

  /** Minta ulang email verifikasi (POST /auth/email/verify/request). */
  async requestEmailVerification(email: string): Promise<void> {
    try {
      const response = await apiClient.post<ApiResponse<null>>("/auth/email/verify/request", { email });
      if (!response.success) throw new AuthError(response.message || "Gagal mengirim email verifikasi.");
    } catch (err) {
      throw new AuthError(
        err instanceof ApiError ? err.message : "Gagal mengirim email verifikasi.",
        err instanceof ApiError ? err.status : undefined
      );
    }
  },

  /** Verifikasi email dengan token dari email (POST /auth/email/verify). */
  async verifyEmail(token: string): Promise<void> {
    try {
      const response = await apiClient.post<ApiResponse<null>>("/auth/email/verify", { token });
      if (!response.success) throw new AuthError(response.message || "Verifikasi email gagal.");
    } catch (err) {
      throw new AuthError(
        err instanceof ApiError ? err.message : "Verifikasi email gagal.",
        err instanceof ApiError ? err.status : undefined
      );
    }
  },

  /** Cek sesi saat mount — panggil /auth/me. */
  async checkSession(): Promise<SessionUser | null> {
    const token = typeof window !== "undefined" ? localStorage.getItem("auth_token") : null;
    if (!token) return null;

    try {
      apiClient.setToken(token);
      const response = await apiClient.get<MeResponse>("/auth/me");

      if (response.success && response.data) {
        const user = mapBackendUserToSessionUser(response.data);
        useAuthStore.getState().login(user, true);
        return user;
      }
    } catch (error) {
      // Token tidak valid atau kedaluwarsa
      apiClient.clearTokens();
      useAuthStore.getState().logout();
    }

    return null;
  },
};

/** Helper untuk komponen: halaman default yang dituju setelah login, per role. */
export function homeForRole(role: SessionUser["role"]): string {
  return POST_LOGIN_HOME[role as AppRole] ?? "/";
}