import type { Metadata } from "next";
import Link from "next/link";
import { HelpCircle } from "lucide-react";
import { FaqAccordion } from "@/features/public/components/faq-accordion";
import { faqService } from "@/services/faq.service";

export const metadata: Metadata = {
  title: "FAQ",
  description: "Pertanyaan yang sering diajukan pengunjung Kawah Putih.",
};

export default async function FaqPage() {
  const faqs = await faqService.listPublished();

  return (
    <div className="page-shell py-12">
      <div className="mb-8 max-w-2xl">
        <span className="inline-flex items-center gap-2 rounded-full bg-[#e9f3fb] px-3 py-1 text-xs font-semibold text-[#1768ad]">
          <HelpCircle className="h-3.5 w-3.5" /> Pusat Bantuan
        </span>
        <h1 className="mt-4 font-heading text-3xl font-bold text-[#17324d]">FAQ</h1>
        <p className="mt-3 text-[15px] leading-relaxed text-[#698097]">
          Pertanyaan yang sering diajukan pengunjung Kawah Putih
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <FaqAccordion items={faqs} />

        <aside className="h-fit rounded-md border border-[#1768ad] bg-[#1768ad] p-6 text-white lg:sticky lg:top-24">
          <h2 className="text-lg font-bold">Masih ada pertanyaan?</h2>
          <p className="mt-2 text-sm text-blue-100">
            Tim kami siap membantu. Hubungi pusat informasi atau pesan tiketmu sekarang.
          </p>
          <Link
            href="/contact"
            className="mt-5 inline-flex h-10 items-center rounded-md border border-white/40 px-5 text-sm font-semibold text-white hover:bg-white/10"
          >
            Hubungi Kami
          </Link>
          <Link
            href="/packages"
            className="mt-2 flex h-10 items-center justify-center rounded-md bg-white px-5 text-sm font-semibold text-[#1768ad] hover:bg-blue-50"
          >
            Pesan Tiket
          </Link>
        </aside>
      </div>
    </div>
  );
}