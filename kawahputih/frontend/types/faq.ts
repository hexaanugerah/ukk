export interface FAQ {
  id: string;
  question: string;
  answer: string;
  category: string;
  sort_order: number;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateFAQPayload {
  question: string;
  answer: string;
  category: string;
  sort_order?: number;
  is_published?: boolean;
}

export type UpdateFAQPayload = CreateFAQPayload;

export const FAQ_CATEGORIES = [
  "Tiket",
  "Jam Operasional",
  "Lokasi",
  "Fasilitas",
  "Aturan Kunjungan",
  "Parkir",
  "Pembayaran",
  "Umum",
] as const;

export type FAQCategory = (typeof FAQ_CATEGORIES)[number];