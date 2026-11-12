import { z } from "zod";

export const galleryItemSchema = z.object({
  title: z.string().min(3, "Judul minimal 3 karakter").max(200),
  image_url: z.string().min(1, "URL gambar wajib diisi").url("URL gambar tidak valid"),
  category: z.string().max(100).optional().default(""),
});
export type GalleryItemFormValues = z.infer<typeof galleryItemSchema>;
