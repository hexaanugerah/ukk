import type { Metadata } from "next";
import { FAQAdmin } from "@/features/faq/components/faq-admin";

export const metadata: Metadata = { title: "Kelola FAQ" };

export default function AdminFAQPage() {
  return (
    <div>
      <h1 className="mb-6 text-xl font-semibold">Kelola FAQ</h1>
      <FAQAdmin />
    </div>
  );
}