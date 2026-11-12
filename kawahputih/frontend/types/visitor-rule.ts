export interface VisitorRule {
  id: string;
  title: string;
  description: string;
  category: string;
  priority: "low" | "normal" | "high" | "critical";
  sort_order: number;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateVisitorRulePayload {
  title: string;
  description: string;
  category: string;
  priority?: "low" | "normal" | "high" | "critical";
  sort_order?: number;
  is_published?: boolean;
}

export type UpdateVisitorRulePayload = CreateVisitorRulePayload;

export const VISITOR_RULE_CATEGORIES = [
  "Keselamatan",
  "Lingkungan",
  "Kebersihan",
  "Kendaraan",
  "Aktivitas Pengunjung",
  "Larangan",
  "Informasi Umum",
] as const;

export const VISITOR_RULE_PRIORITIES = ["low", "normal", "high", "critical"] as const;

export type VisitorRuleCategory = (typeof VISITOR_RULE_CATEGORIES)[number];
export type VisitorRulePriority = (typeof VISITOR_RULE_PRIORITIES)[number];