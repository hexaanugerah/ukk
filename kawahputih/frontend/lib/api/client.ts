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
      this.token = localStorage.getItem("auth_token");
    }
  }

  setToken(token: string | null) {
    this.token = token;
    if (typeof window !== "undefined") {
      if (token) {
        localStorage.setItem("auth_token", token);
      } else {
        localStorage.removeItem("auth_token");
      }
    }
  }

  getToken(): string | null {
    return this.token;
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

  private async request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    const { params, headers, timeout = DEFAULT_TIMEOUT, ...fetchOptions } = options;

    const requestHeaders: HeadersInit = {
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
      }	      if (!response.ok) {
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
}

export const apiClient = new ApiClient(API_BASE_URL);