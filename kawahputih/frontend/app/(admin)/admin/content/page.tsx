"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { PageHeader } from "@/components/panel/panel-ui";
import { Button } from "@/components/ui/button";
import {
  ArrowRight,
  Building2,
  FileText,
  HelpCircle,
  Image as ImageIcon,
  Loader2,
  Shield,
  type LucideIcon,
} from "lucide-react";
import { toast } from "sonner";
import { destinationInfoService } from "@/services/destination-info.service";
import { faqService } from "@/services/faq.service";
import { facilityService } from "@/services/facility.service";
import { visitorRuleService } from "@/services/visitor-rule.service";
import { galleryService } from "@/services/gallery.service";

interface ModuleStat {
  total: number;
  published: number;
  draft: number;
}

type ModuleKey = "destinationInfo" | "gallery" | "faq" | "facilities" | "visitorRules";

type ContentStats = Record<ModuleKey, ModuleStat>;

interface ContentModule {
  key: ModuleKey;
  title: string;
  desc: string;
  href: string;
  icon: LucideIcon;
  color: string;
}

const MODULES: ContentModule[] = [
  {
    key: "destinationInfo",
    title: "Informasi Kawah Putih",
    desc: "Informasi lengkap mengenai destinasi, sejarah, dan daya tarik Kawah Putih.",
    href: "/admin/destination-info",
    icon: FileText,
    color: "bg-blue-100 text-blue-600",
  },
  {
    key: "gallery",
    title: "Galeri",
    desc: "Kumpulan foto dan dokumentasi keindahan Kawah Putih dan area wisata.",
    href: "/admin/gallery",
    icon: ImageIcon,
    color: "bg-green-100 text-green-600",
  },
  {
    key: "faq",
    title: "FAQ",
    desc: "Kumpulan pertanyaan yang sering diajukan oleh pengunjung.",
    href: "/admin/faq",
    icon: HelpCircle,
    color: "bg-amber-100 text-amber-600",
  },
  {
    key: "facilities",
    title: "Informasi Fasilitas",
    desc: "Informasi fasilitas dan layanan yang tersedia untuk kenyamanan pengunjung.",
    href: "/admin/facilities",
    icon: Building2,
    color: "bg-purple-100 text-purple-600",
  },
  {
    key: "visitorRules",
    title: "Aturan Kunjungan",
    desc: "Panduan dan ketentuan yang perlu diperhatikan selama berkunjung.",
    href: "/admin/visitor-rules",
    icon: Shield,
    color: "bg-red-100 text-red-600",
  },
];

export default function AdminContentPage() {
  const [stats, setStats] = useState<ContentStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState("-");

  const fetchStats = useCallback(async () => {
    setLoading(true);
    const results = await Promise.allSettled([
      destinationInfoService.list({ limit: 100 }),
      galleryService.list({ limit: 500 }),
      faqService.list({ limit: 500 }),
      facilityService.list({ limit: 500 }),
      visitorRuleService.list({ limit: 500 }),
    ]);

    const valueOf = <T,>(result: PromiseSettledResult<T[]>): T[] =>
      result.status === "fulfilled" ? result.value : [];

    const destInfo = valueOf(results[0] as PromiseSettledResult<Awaited<ReturnType<typeof destinationInfoService.list>>>);
    const galleryItems = valueOf(results[1] as PromiseSettledResult<Awaited<ReturnType<typeof galleryService.list>>>);
    const faqs = valueOf(results[2] as PromiseSettledResult<Awaited<ReturnType<typeof faqService.list>>>);
    const facilities = valueOf(results[3] as PromiseSettledResult<Awaited<ReturnType<typeof facilityService.list>>>);
    const visitorRules = valueOf(results[4] as PromiseSettledResult<Awaited<ReturnType<typeof visitorRuleService.list>>>);

    setStats({
      destinationInfo: {
        total: destInfo.length,
        published: destInfo.filter((d) => d.status === "published").length,
        draft: destInfo.filter((d) => d.status !== "published").length,
      },
      gallery: {
        total: galleryItems.length,
        published: galleryItems.length,
        draft: 0,
      },
      faq: {
        total: faqs.length,
        published: faqs.filter((f) => f.is_published).length,
        draft: faqs.filter((f) => !f.is_published).length,
      },
      facilities: {
        total: facilities.length,
        published: facilities.filter((f) => f.is_published).length,
        draft: facilities.filter((f) => !f.is_published).length,
      },
      visitorRules: {
        total: visitorRules.length,
        published: visitorRules.filter((r) => r.is_published).length,
        draft: visitorRules.filter((r) => !r.is_published).length,
      },
    });

    setLastUpdated(
      new Date().toLocaleString("id-ID", {
        day: "numeric",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      })
    );

    if (results.some((r) => r.status === "rejected")) {
      toast.error("Sebagian statistik konten gagal dimuat");
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  return (
    <div>
      <div className="mb-6 flex items-center justify-between gap-4">
        <PageHeader title="Kelola Konten" subtitle="Atur informasi dan konten yang ditampilkan di website" />
        <Button variant="outline" onClick={fetchStats} loading={loading}>
          {!loading && <Loader2 className="mr-1 h-4 w-4" />}
          Refresh
        </Button>
      </div>

      <div className="space-y-3">
        {MODULES.map((module) => {
          const Icon = module.icon;
          const stat = stats?.[module.key];
          return (
            <div
              key={module.key}
              className="flex flex-col gap-4 rounded-lg border border-[#c8dded] bg-white p-4 sm:flex-row sm:items-center"
            >
              <div className={`grid h-12 w-12 shrink-0 place-items-center rounded-lg ${module.color}`}>
                <Icon className="h-6 w-6" />
              </div>
              <div className="min-w-0 flex-1">
                <h3 className="text-sm font-extrabold text-[#14548f]">{module.title}</h3>
                <p className="mt-0.5 text-xs leading-relaxed text-[#698097]">{module.desc}</p>
                <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs">
                  <span className="font-semibold text-[#29445e]">
                    {stat ? `${stat.total.toLocaleString("id-ID")} item` : loading ? "Memuat..." : "0 item"}
                  </span>
                  {stat && (
                    <>
                      <span className="text-green-600">{stat.published} dipublikasikan</span>
                      <span className="text-amber-600">{stat.draft} draft</span>
                    </>
                  )}
                </div>
              </div>
              <Link
                href={module.href}
                className="inline-flex shrink-0 items-center justify-center gap-2 rounded-btn bg-brand px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-brand-dark"
              >
                Kelola Konten
                <ArrowRight className="h-4 w-4" />
              </Link>
            </div>
          );
        })}
      </div>

      <div className="mt-8 rounded-lg border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-900">
        <h3 className="mb-2 text-sm font-semibold text-slate-900 dark:text-white">Informasi</h3>
        <p className="text-sm text-slate-600 dark:text-slate-400">
          Terakhir diperbarui: <span className="font-mono">{lastUpdated}</span>
        </p>
        <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">
          Setiap modul memiliki halaman kelola sendiri untuk menambah, mengubah, mempublikasikan, dan menghapus
          konten. Perubahan pada konten yang dipublikasikan langsung tercermin di website publik.
        </p>
      </div>
    </div>
  );
}
