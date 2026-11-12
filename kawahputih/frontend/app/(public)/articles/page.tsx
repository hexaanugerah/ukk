"use client";

import Link from "next/link";
import { formatDate } from "@/lib/date";
import { useArticles } from "@/hooks/use-articles";

// Artikel (publik) — daftar artikel published. Client component agar perubahan dari CMS admin langsung tampil tanpa build.

export default function ArticlesPage() {
  const { data, isLoading, isError } = useArticles({ status: "published", limit: 20 });
  const articles = data ?? [];

  return (
    <div className="page-shell py-10">
      <h1 className="mb-6 text-2xl font-semibold">Artikel</h1>
      {isLoading ? (
        <p className="py-10 text-center text-sm text-slate-500">Memuat artikel...</p>
      ) : isError ? (
        <p className="py-10 text-center text-sm text-slate-500">Terjadi kesalahan saat memuat artikel.</p>
      ) : articles.length === 0 ? (
        <p className="text-sm text-slate-500">Belum ada artikel yang dipublikasikan.</p>
      ) : (
        <div className="space-y-6">
          {articles.map((a) => (
            <Link key={a.id} href={`/articles/${a.id}`} className="block rounded-2xl border border-slate-200 p-5 hover:shadow-sm dark:border-slate-800">
              <h2 className="font-semibold">{a.title}</h2>
              <p className="mt-1 line-clamp-2 text-sm text-slate-500">{a.excerpt}</p>
              {a.published_at && <p className="mt-2 text-xs text-slate-400">{formatDate(new Date(a.published_at * 1000).toISOString())}</p>}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
