import { z } from "zod";
import { FAQ_CATEGORIES } from "@/types/faq";

export const faqSchema = z.object({
  question: z.string().min(5, "Pertanyaan minimal 5 karakter").max(500),
  answer: z.string().min(10, "Jawaban minimal 10 karakter"),
  category: z.enum(FAQ_CATEGORIES),
  sort_order: z.number().int().min(0).optional().default(0),
  is_published: z.boolean().optional().default(false),
});

export type FAQFormValues = z.infer<typeof faqSchema>;