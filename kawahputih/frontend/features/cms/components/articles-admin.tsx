"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useArticles, useCreateArticle, usePublishArticle, useDeleteArticle } from "@/hooks/use-articles";
import { DataTable } from "@/components/tables/data-table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/empty/empty-state";
import { Skeleton } from "@/components/loading/skeleton";
import { Plus } from "lucide-react";
import { articleSchema, type ArticleFormValues } from "../schemas/article.schema";
import type { Article } from "@/types/article";

export function ArticlesAdmin() {
  const [showForm, setShowForm] = useState(false);
  const { data, isLoading, isError } = useArticles({ limit: 20 });
  const createArticle = useCreateArticle();
  const publishArticle = usePublishArticle();
  const deleteArticle = useDeleteArticle();

  const { register, handleSubmit, reset, formState: { errors } } = useForm<ArticleFormValues>({
    resolver: zodResolver(articleSchema),
  });

  function onSubmit(values: ArticleFormValues) {
    createArticle.mutate(
      { ...values, cover_image: values.cover_image || undefined },
      {
        onSuccess: () => {
          toast.success("Artikel berhasil dibuat sebagai draft");
          reset();
          setShowForm(false);
        },
        onError: () => toast.error("Gagal membuat artikel"),
      }
    );
  }

  return (
    <div className="space-y-4">
      <Button onClick={() => setShowForm(!showForm)}>
        <Plus className="mr-1 h-4 w-4" /> Artikel Baru
      </Button>

      {showForm && (
        <form onSubmit={handleSubmit(onSubmit)} className="max-w-lg space-y-3 rounded-2xl border border-slate-200 p-4 dark:border-slate-800">
          <div>
            <label className="mb-1 block text-sm font-medium">Judul</label>
            <Input {...register("title")} />
            {errors.title && <p className="mt-1 text-xs text-red-600">{errors.title.message}</p>}
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Ringkasan</label>
            <Input {...register("excerpt")} />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Konten</label>
            <textarea
              className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm dark:border-slate-700 dark:bg-slate-900"
              rows={6}
              {...register("content")}
            />
            {errors.content && <p className="mt-1 text-xs text-red-600">{errors.content.message}</p>}
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">URL Gambar Sampul (opsional)</label>
            <Input {...register("cover_image")} />
          </div>
          <Button type="submit" loading={createArticle.isPending}>
            {createArticle.isPending ? "Menyimpan..." : "Simpan sebagai Draft"}
          </Button>
        </form>
      )}

      {isLoading ? (
        <Skeleton className="h-64 w-full" />
      ) : isError ? (
        <EmptyState title="Gagal memuat artikel" />
      ) : !data?.length ? (
        <EmptyState title="Belum ada artikel" />
      ) : (
        <DataTable<Article>
          rows={data}
          columns={[
            { header: "Judul", cell: (a) => a.title },
            { header: "Status", cell: (a) => <Badge status={a.status} /> },
            {
              header: "Aksi",
              cell: (a) => (
                <div className="flex gap-2">
                  {a.status === "draft" && (
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() =>
                        publishArticle.mutate(a.id, {
                          onSuccess: () => toast.success("Artikel dipublikasikan"),
                          onError: () => toast.error("Gagal mempublikasikan"),
                        })
                      }
                    >
                      Publikasikan
                    </Button>
                  )}
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={() =>
                      deleteArticle.mutate(a.id, {
                        onSuccess: () => toast.success("Artikel dihapus"),
                        onError: () => toast.error("Gagal menghapus"),
                      })
                    }
                  >
                    Hapus
                  </Button>
                </div>
              ),
            },
          ]}
        />
      )}
    </div>
  );
}
