// api.ts meniru pkg/response.Envelope backend persis. Semua tipe balikan
// service dibangun di atas ApiResponse<T>, sehingga perubahan bentuk
// envelope backend hanya perlu diubah di satu file ini.
export interface ApiMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
  meta?: ApiMeta;
  errors?: unknown;
}
