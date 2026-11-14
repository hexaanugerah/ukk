import { z } from "zod";

// Berkas gambar tidak masuk skema zod karena tidak lewat react-hook-form;
// validasinya dilakukan manual di komponen (wajib saat create, opsional saat edit).
export const galleryItemSchema = z.object({
  title: z.string().min(3, "Judul minimal 3 karakter").max(200),
  category: z.string().max(100).optional().default(""),
});
export type GalleryItemFormValues = z.infer<typeof galleryItemSchema>;
