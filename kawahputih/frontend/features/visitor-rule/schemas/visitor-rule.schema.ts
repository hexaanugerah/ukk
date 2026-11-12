import { z } from "zod";
import { VISITOR_RULE_CATEGORIES, VISITOR_RULE_PRIORITIES } from "@/types/visitor-rule";

export const visitorRuleSchema = z.object({
  title: z.string().min(3, "Judul aturan minimal 3 karakter").max(200),
  description: z.string().min(10, "Deskripsi aturan minimal 10 karakter"),
  category: z.enum(VISITOR_RULE_CATEGORIES),
  priority: z.enum(VISITOR_RULE_PRIORITIES).optional().default("normal"),
  sort_order: z.number().int().min(0).optional().default(0),
  is_published: z.boolean().optional().default(false),
});

export type VisitorRuleFormValues = z.infer<typeof visitorRuleSchema>;