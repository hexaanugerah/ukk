import { z } from "zod";

export const articleSchema = z.object({
  title: z.string().min(5, "Judul minimal 5 karakter").max(200),
  excerpt: z.string().max(500).optional().default(""),
  content: z.string().min(20, "Konten minimal 20 karakter"),
  cover_image: z.string().url("URL gambar tidak valid").optional().or(z.literal("")),
});
export type ArticleFormValues = z.infer<typeof articleSchema>;
