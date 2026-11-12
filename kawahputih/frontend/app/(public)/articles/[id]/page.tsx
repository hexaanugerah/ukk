"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useArticle } from "@/hooks/use-articles";

// Detail artikel (publik) — dibaca lewat client hook agar perubahan CMS admin langsung terlihat. Menampilkan state loading/error.

export default function ArticleDetailPage() {
  const params = useParams<{ id: string }>();
  const id = typeof params?.id === "string" ? params.id : "";
  const { data: article, isLoading, isError } = useArticle(id);

  if (isLoading) {
    return <p className="mx-auto max-w-2xl px-4 py-10 text-sm text-slate-500">Memuat artikel...</p>;
  }

  if (isError || !article || article.status !== "published") {
    return (
      <div className="mx-auto max-w-2xl px-4 py-10">
        <h1 className="text-2xl font-semibold">Artikel tidak ditemukan</h1>
        <p className="mt-2 text-sm text-slate-500">Artikel yang Anda cari tidak tersedia atau belum dipublikasikan.</p>
        <Link href="/articles" className="mt-4 inline-block text-sm text-brand-600 underline">
          &larr; Kembali ke daftar artikel
        </Link>
      </div>
    );
  }

  return (
    <article className="mx-auto max-w-2xl px-4 py-10">
      <h1 className="mb-4 text-2xl font-semibold">{article.title}</h1>
      <div className="prose prose-slate dark:prose-invert max-w-none whitespace-pre-line">{article.content}</div>
    </article>
  );
}
