import { z } from "zod";

export const packageSchema = z.object({
  code: z
    .string()
    .min(3, "Kode paket minimal 3 karakter")
    .max(20, "Kode paket maksimal 20 karakter")
    .regex(/^[A-Z0-9]+$/, "Kode paket hanya boleh berisi huruf besar dan angka"),
  name: z
    .string()
    .min(3, "Nama paket minimal 3 karakter")
    .max(100, "Nama paket maksimal 100 karakter")
    .regex(/^[^<>]*$/, "Nama tidak boleh mengandung karakter khusus"),
  description: z
    .string()
    .max(1000, "Deskripsi maksimal 1000 karakter")
    .optional(),
  cover_image: z.string().url("URL gambar tidak valid").optional().or(z.literal("")),
  category: z.enum(["ENTRANCE", "TRANSPORT", "SUNRISE", "PACKAGE", "ACTIVITY", "SERVICE"], {
    errorMap: () => ({ message: "Kategori paket tidak valid" }),
  }),
  price: z.coerce.number().int().positive("Harga harus lebih dari 0"),
  duration_hours: z.coerce.number().int().min(1, "Durasi minimal 1 jam").default(1),
  capacity: z.coerce.number().int().min(1, "Kapasitas minimal 1").max(10000, "Kapasitas maksimal 10000"),
  max_capacity: z.coerce.number().int().min(1, "Kapasitas maksimal minimal 1").max(10000, "Kapasitas maksimal 10000").default(100),
  status: z.enum(["AKTIF", "NONAKTIF"], {
    errorMap: () => ({ message: "Status harus AKTIF atau NONAKTIF" }),
  }),
  notes: z
    .string()
    .max(500, "Catatan maksimal 500 karakter")
    .optional(),
});

export type PackageFormValues = z.infer<typeof packageSchema>;