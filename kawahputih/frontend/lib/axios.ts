// Axios — instans tunggal untuk memanggil backend (endpoint terproteksi).

import axios from "axios";

export const axiosInstance = axios.create({
  baseURL:
    process.env.NEXT_PUBLIC_API_BASE_URL ??
    process.env.NEXT_PUBLIC_API_URL ??
    "http://localhost:8080/api/v1",
  headers: { "Content-Type": "application/json" },
});

// Lampirkan JWT dari login (kunci `auth_token`, sama seperti apiClient) agar
// permintaan ke endpoint admin/konten yang terproteksi ikut terautentikasi.
axiosInstance.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = window.localStorage.getItem("auth_token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

// Auto-refresh: bila access token kedaluwarsa (401), rotasi refresh token via
// POST /auth/refresh sekali, simpan token baru, lalu ulangi permintaan asli.
// Endpoint /auth/* dikecualikan agar tidak terjadi loop.
let refreshing: Promise<string | null> | null = null;

function requestRefresh(): Promise<string | null> {
  if (!refreshing) {
    refreshing = (async () => {
      const refreshToken = window.localStorage.getItem("auth_refresh_token");
      if (!refreshToken) return null;
      try {
        const res = await axios.post(
          `${axiosInstance.defaults.baseURL}/auth/refresh`,
          { refresh_token: refreshToken }
        );
        const access = res.data?.data?.access_token;
        const refresh = res.data?.data?.refresh_token;
        if (!access) {
          window.localStorage.removeItem("auth_token");
          window.localStorage.removeItem("auth_refresh_token");
          return null;
        }
        window.localStorage.setItem("auth_token", access);
        if (refresh) window.localStorage.setItem("auth_refresh_token", refresh);
        return access as string;
      } catch {
        window.localStorage.removeItem("auth_token");
        window.localStorage.removeItem("auth_refresh_token");
        return null;
      }
    })().finally(() => {
      refreshing = null;
    });
  }
  return refreshing;
}

axiosInstance.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config ?? {};
    const status = error.response?.status;
    if (
      status === 401 &&
      !original._retried &&
      !String(original.url ?? "").includes("/auth/")
    ) {
      original._retried = true;
      const token = await requestRefresh();
      if (token) {
        original.headers = { ...original.headers, Authorization: `Bearer ${token}` };
        return axiosInstance(original);
      }
    }
    return Promise.reject(error);
  }
);
