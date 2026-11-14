import { z } from "zod";
import { joinPhone, splitPhone, validatePhoneFull } from "@/lib/phone";

// Skema pengunjung
export const visitorSchema = z.object({
  name: z
    .string()
    .min(3, "Nama minimal 3 karakter")
    .max(100, "Nama maksimal 100 karakter")
    .regex(/^[^<>]*$/, "Nama tidak boleh mengandung karakter khusus"),
  email: z
    .string()
    .min(1, "Email wajib diisi")
    .email("Format email tidak valid")
    .max(150, "Email maksimal 150 karakter"),
  phone: z
    .string()
    .min(1, "Nomor telepon wajib diisi")
    .regex(/^\+[1-9]\d{6,14}$/, "Format nomor telepon tidak valid (contoh: +6281234567890)"),
});

// Skema item booking
export const bookingItemSchema = z.object({
  ticket_id: z.string().uuid("ID tiket tidak valid"),
  quantity: z.coerce.number().int().min(1, "Minimal 1 tiket").max(20, "Maksimal 20 tiket per jenis"),
});

// Skema booking utama
export const bookingSchema = z.object({
  visit_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, "Format tanggal tidak valid (YYYY-MM-DD)"),
  visitor: visitorSchema,
  items: z.array(bookingItemSchema).min(1, "Minimal 1 jenis tiket").max(10, "Maksimal 10 jenis tiket"),
  note: z.string().max(500, "Catatan maksimal 500 karakter").optional(),
});

export type BookingFormValues = z.infer<typeof bookingSchema>;
export type VisitorFormValues = z.infer<typeof visitorSchema>;
export type BookingItemFormValues = z.infer<typeof bookingItemSchema>;

// Helper validasi
export function validateVisitDate(date: string): { valid: boolean; message?: string } {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  
  const selected = new Date(date);
  selected.setHours(0, 0, 0, 0);
  
  if (selected < today) {
    return { valid: false, message: "Tanggal kunjungan tidak boleh di masa lalu" };
  }
  
  const maxDate = new Date(today);
  maxDate.setDate(today.getDate() + 60);
  
  if (selected > maxDate) {
    return { valid: false, message: "Booking maksimal 60 hari ke depan" };
  }
  
  return { valid: true };
}

export function validatePhone(phone: string): { valid: boolean; message?: string } {
  // Menerima format lama "08..." maupun E.164 "+62..." / "+xx...".
  return validatePhoneFull(phone, true);
}

export function sanitizeInput(input: string): string {
  return input
    .replace(/\0/g, "")
    .trim()
    .slice(0, 10000);
}

export function sanitizeEmail(email: string): string {
  return email.toLowerCase().trim();
}

export function sanitizePhone(phone: string): string {
  // Normalisasi ke E.164: "0812..." / "62812..." / "+62..." → "+62812...".
  const { dial, national } = splitPhone(phone);
  return joinPhone(dial, national);
}