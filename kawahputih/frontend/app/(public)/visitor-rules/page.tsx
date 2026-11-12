import type { Metadata } from "next";
import { Shield, AlertTriangle } from "lucide-react";
import { visitorRuleService } from "@/services/visitor-rule.service";

export const metadata: Metadata = {
  title: "Aturan Kunjungan",
  description: "Panduan dan ketentuan yang perlu diperhatikan selama berkunjung ke Kawah Putih.",
};

const PRIORITY_ICONS = {
  low: "🟢",
  normal: "🔵",
  high: "🟠",
  critical: "🔴",
};

export default async function VisitorRulesPage() {
  const rules = await visitorRuleService.listPublished();

  // Kelompokkan aturan berdasarkan kategori
  const groupedRules = rules.reduce((acc, rule) => {
    const bucket = acc[rule.category] ?? (acc[rule.category] = []);
    bucket.push(rule);
    return acc;
  }, {} as Record<string, typeof rules>);

  const categories = Object.keys(groupedRules).sort();

  return (
    <div className="page-shell py-12">
      <div className="mb-8 max-w-2xl">
        <span className="inline-flex items-center gap-2 rounded-full bg-[#e9f3fb] px-3 py-1 text-xs font-semibold text-[#1768ad]">
          <Shield className="h-3.5 w-3.5" /> Panduan Wisata
        </span>
        <h1 className="mt-4 font-heading text-3xl font-bold text-[#17324d]">Aturan Kunjungan</h1>
        <p className="mt-3 text-[15px] leading-relaxed text-[#698097]">
          Panduan dan ketentuan yang perlu diperhatikan selama berkunjung ke Kawah Putih untuk keselamatan dan kenyamanan bersama.
        </p>
      </div>

      {rules.length === 0 ? (
        <div className="rounded-md border border-[#d7e3ed] bg-white p-8 text-center">
          <Shield className="mx-auto h-12 w-12 text-slate-300" />
          <h2 className="mt-4 text-lg font-semibold text-[#17324d]">Belum ada aturan kunjungan</h2>
          <p className="mt-2 text-sm text-[#698097]">Aturan kunjungan akan ditampilkan di sini setelah dipublikasikan oleh admin.</p>
        </div>
      ) : (
        <div className="space-y-6">
          {categories.map((category) => (
            <section key={category}>
              <h2 className="mb-3 flex items-center gap-2 text-lg font-bold text-[#17324d]">
                <AlertTriangle className="h-5 w-5 text-[#1768ad]" />
                {category}
              </h2>
              <div className="divide-y divide-[#e5edf3] rounded-md border border-[#d7e3ed] bg-white">
                {(groupedRules[category] ?? []).map((rule) => (
                  <div key={rule.id} className="px-5 py-4">
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex-1">
                        <h3 className="text-sm font-semibold text-[#17324d]">{rule.title}</h3>
                        <p className="mt-1.5 text-sm leading-relaxed text-[#698097]">{rule.description}</p>
                      </div>
                      <span className="flex shrink-0 items-center gap-1 rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-medium text-slate-600 dark:bg-slate-800 dark:text-slate-400">
                        <span>{PRIORITY_ICONS[rule.priority]}</span>
                        {rule.priority.charAt(0).toUpperCase() + rule.priority.slice(1)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}