import type { Metadata } from "next";
import { VisitorRuleAdmin } from "@/features/visitor-rule/components/visitor-rule-admin";

export const metadata: Metadata = { title: "Kelola Aturan Kunjungan" };

export default function AdminVisitorRulePage() {
  return (
    <div>
      <h1 className="mb-6 text-xl font-semibold">Kelola Aturan Kunjungan</h1>
      <VisitorRuleAdmin />
    </div>
  );
}