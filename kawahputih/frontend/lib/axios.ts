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
