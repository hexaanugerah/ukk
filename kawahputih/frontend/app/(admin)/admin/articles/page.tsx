import type { Metadata } from "next";
import { ArticlesAdmin } from "@/features/cms/components/articles-admin";

export const metadata: Metadata = { title: "Kelola Artikel" };

export default function AdminArticlesPage() {
  return (
    <div>
      <h1 className="mb-6 text-xl font-semibold">Kelola Artikel</h1>
      <ArticlesAdmin />
    </div>
  );
}
