// API client — HTTP client untuk berkomunikasi dengan backend.

// Prioritas env sama seperti lib/axios.ts, agar kedua HTTP client
// menghasilkan base URL yang sama apa pun nama env yang dipakai deployment.
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ??
  process.env.NEXT_PUBLIC_API_URL ??
  "http://localhost:8080/api/v1";
const DEFAULT_TIMEOUT = 30000; // 30 seconds

interface RequestOptions extends RequestInit {
  params?: Record<string, string>;
  timeout?: number;
}

const ACCESS_TOKEN_KEY = "auth_token";
const REFRESH_TOKEN_KEY = "auth_refresh_token";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: any
  ) {
    super(message);
    this.name = "ApiError";
  }
}

class ApiClient {
  private baseUrl: string;
  private token: string | null = null;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
    // Ambil token dari localStorage saat inisialisasi (hanya di sisi klien)
    if (typeof window !== "undefined") {
      this.token = localStorage.getItem(ACCESS_TOKEN_KEY);
    }
  }

  setToken(token: string | null) {
    this.token = token;
    if (typeof window !== "undefined") {
      if (token) {
        localStorage.setItem(ACCESS_TOKEN_KEY, token);
      } else {
        localStorage.removeItem(ACCESS_TOKEN_KEY);
      }
    }
  }

  getToken(): string | null {
    return this.token;
  }

  /** Simpan/hapus refresh token (localStorage) — dipakai auto-refresh & logout. */
  setRefreshToken(token: string | null) {
    if (typeof window !== "undefined") {
      if (token) {
        localStorage.setItem(REFRESH_TOKEN_KEY, token);
      } else {
        localStorage.removeItem(REFRESH_TOKEN_KEY);
      }
    }
  }

  getRefreshToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(REFRESH_TOKEN_KEY);
  }

  /** Hapus semua kredensial lokal (dipanggil saat refresh gagal). */
  clearTokens() {
    this.setToken(null);
    this.setRefreshToken(null);
  }

  /**
   * Auto-refresh: rotasi refresh token sekali lewat POST /auth/refresh, lalu
   * ulangi permintaan asli dengan access token baru. Endpoint /auth/* tidak
   * ikut di-refresh agar tidak terjadi loop (login/refresh/logout sendiri).
   */
  private refreshPromise: Promise<boolean> | null = null;

  private tryRefresh(): Promise<boolean> {
    // Satu refresh bersama untuk semua permintaan 401 yang terjadi bersamaan —
    // refresh token bersifat rotasi sekali-pakai, jadi tidak boleh dua klien
    // memakai token yang sama secara paralel.
    if (!this.refreshPromise) {
      this.refreshPromise = this.doRefresh().finally(() => {
        this.refreshPromise = null;
      });
    }
    return this.refreshPromise;
  }

  private async doRefresh(): Promise<boolean> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) return false;
    try {
      const res = await fetch(`${this.baseUrl}/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
      if (!res.ok) {
        // Refresh token kedaluwarsa/dicabut — sesi tidak bisa diselamatkan.
        this.clearTokens();
        return false;
      }
      const data = await res.json();
      const accessToken: string | undefined = data?.data?.access_token;
      const newRefreshToken: string | undefined = data?.data?.refresh_token;
      if (!accessToken) {
        this.clearTokens();
        return false;
      }
      this.setToken(accessToken);
      if (newRefreshToken) this.setRefreshToken(newRefreshToken);
      return true;
    } catch {
      return false;
    }
  }

  private static isAuthEndpoint(endpoint: string): boolean {
    return endpoint.startsWith("/auth/");
  }

  private buildUrl(endpoint: string, params?: Record<string, string>): string {
    const url = new URL(`${this.baseUrl}${endpoint}`);
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        url.searchParams.append(key, value);
      });
    }
    return url.toString();
  }

  private async request<T>(endpoint: string, options: RequestOptions = {}, isRetry = false): Promise<T> {
    const { params, headers, timeout = DEFAULT_TIMEOUT, ...fetchOptions } = options;

    // FormData harus dibiarkan tanpa Content-Type agar browser menuliskan
    // boundary multipart sendiri.
    const isFormData =
      typeof FormData !== "undefined" && fetchOptions.body instanceof FormData;

    const requestHeaders: HeadersInit = isFormData
      ? { ...headers }
      : {
          "Content-Type": "application/json",
          ...headers,
        };

    if (this.token) {
      (requestHeaders as Record<string, string>)["Authorization"] = `Bearer ${this.token}`;
    }

    const url = this.buildUrl(endpoint, params);

    // Buat abort controller untuk batas waktu
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
      const response = await fetch(url, {
        ...fetchOptions,
        headers: requestHeaders,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      let data: any;
      try {
        data = await response.json();
      } catch {
        data = null;
      }
      if (!response.ok) {
        // Access token kedaluwarsa → coba rotasi refresh token, lalu ulangi
        // permintaan ini sekali dengan token baru.
        if (
          response.status === 401 &&
          !isRetry &&
          !ApiClient.isAuthEndpoint(endpoint) &&
          (await this.tryRefresh())
        ) {
          return this.request<T>(endpoint, options, true);
        }
        const message = data?.message || data?.error || "Terjadi kesalahan. Silakan coba kembali.";
        throw new ApiError(message, response.status, data);
      }

      // Backend mengembalikan { success, message?, data?, meta? }.
      // Envelope utuh dikembalikan agar pemanggil bisa mengakses success/message/meta.
      return data as T;
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof ApiError) throw error;	      if (error instanceof DOMException && error.name === "AbortError") {
	        throw new ApiError("Koneksi terlalu lama. Silakan coba kembali.", 408);
	      }
	      throw new ApiError("Tidak dapat terhubung ke server. Silakan coba kembali.", 0);
    }
  }

  async get<T>(endpoint: string, params?: Record<string, string>, timeout?: number): Promise<T> {
    return this.request<T>(endpoint, { method: "GET", params, timeout });
  }

  async post<T>(endpoint: string, body: any, timeout?: number): Promise<T> {
    return this.request<T>(endpoint, {
      method: "POST",
      body: JSON.stringify(body),
      timeout,
    });
  }

  async put<T>(endpoint: string, body: any, timeout?: number): Promise<T> {
    return this.request<T>(endpoint, {
      method: "PUT",
      body: JSON.stringify(body),
      timeout,
    });
  }

  async patch<T>(endpoint: string, body: any, timeout?: number): Promise<T> {
    return this.request<T>(endpoint, {
      method: "PATCH",
      body: JSON.stringify(body),
      timeout,
    });
  }

  async delete<T>(endpoint: string, timeout?: number): Promise<T> {
    return this.request<T>(endpoint, { method: "DELETE", timeout });
  }

  /** Unggah berkas multipart (FormData) — dipakai unggah bukti pembayaran. */
  async postForm<T>(endpoint: string, form: FormData, timeout?: number): Promise<T> {
    return this.request<T>(endpoint, { method: "POST", body: form, timeout });
  }
}

export const apiClient = new ApiClient(API_BASE_URL);