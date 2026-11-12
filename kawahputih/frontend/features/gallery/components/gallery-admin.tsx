"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import Image from "next/image";
import { useGallery, useCreateGalleryItem, useUpdateGalleryItem, useDeleteGalleryItem } from "@/hooks/use-gallery";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { CardGridSkeleton } from "@/components/loading/skeleton";
import { EmptyState } from "@/components/empty/empty-state";
import { Plus, Trash2, Edit, Eye } from "lucide-react";
import { galleryItemSchema, type GalleryItemFormValues } from "../schemas/gallery.schema";

export function GalleryAdmin() {
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  // Item yang menunggu konfirmasi hapus — dialog muncul sebelum delete dibuat.
  const [pendingDelete, setPendingDelete] = useState<GalleryItemFormValues & { id: string } | null>(null);
  const { data, isLoading, isError } = useGallery({ limit: 24 });
  const createItem = useCreateGalleryItem();
  const updateItem = useUpdateGalleryItem();
  const deleteItem = useDeleteGalleryItem();

  const { register, handleSubmit, reset, formState: { errors } } = useForm<GalleryItemFormValues>({
    resolver: zodResolver(galleryItemSchema),
  });

  function onSubmit(values: GalleryItemFormValues) {
    if (editingId) {
      updateItem.mutate(
        { id: editingId, payload: values },
        {
          onSuccess: () => {
            toast.success("Item galeri diperbarui");
            reset();
            setShowForm(false);
            setEditingId(null);
          },
          onError: () => toast.error("Gagal memperbarui item"),
        }
      );
    } else {
      createItem.mutate(values, {
        onSuccess: () => {
          toast.success("Item galeri ditambahkan");
          reset();
          setShowForm(false);
        },
        onError: () => toast.error("Gagal menambahkan item"),
      });
    }
  }

  function handleEdit(item: GalleryItemFormValues & { id: string }) {
    setEditingId(item.id);
    setShowForm(true);
    reset({
      title: item.title,
      image_url: item.image_url,
      category: item.category,
    });
  }

  function handleCancel() {
    setShowForm(false);
    setEditingId(null);
    reset();
  }

  return (
    <div className="space-y-4">
      <Button onClick={() => { handleCancel(); setShowForm(!showForm); }}>
        <Plus className="mr-1 h-4 w-4" /> {showForm ? "Tutup Form" : "Tambah Foto"}
      </Button>      {showForm && (
        <form onSubmit={handleSubmit(onSubmit)} className="max-w-lg space-y-3 rounded-2xl border border-slate-200 p-4 dark:border-slate-800">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">{editingId ? "Edit Foto" : "Tambah Foto Baru"}</h3>
            <Button type="button" variant="ghost" size="sm" onClick={handleCancel}>Batal</Button>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Judul</label>
            <Input {...register("title")} />
            {errors.title && <p className="mt-1 text-xs text-red-600">{errors.title.message}</p>}
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">URL Gambar</label>
            <Input {...register("image_url")} />
            {errors.image_url && <p className="mt-1 text-xs text-red-600">{errors.image_url.message}</p>}
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Kategori</label>
            <Input placeholder="mis. Pemandangan, Fasilitas" {...register("category")} />
          </div>
          <Button type="submit" loading={createItem.isPending || updateItem.isPending}>
            {editingId ? (updateItem.isPending ? "Menyimpan..." : "Simpan Perubahan") : (createItem.isPending ? "Menyimpan..." : "Simpan")}
          </Button>
        </form>
      )}

      {isLoading ? (
        <CardGridSkeleton count={8} />
      ) : isError ? (
        <EmptyState title="Gagal memuat galeri" />
      ) : !data?.length ? (
        <EmptyState title="Belum ada foto" />
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
          {data.map((item) => (
            <div key={item.id} className="group relative aspect-square overflow-hidden rounded-xl bg-slate-100 dark:bg-slate-800">
              <Image src={item.image_url} alt={item.title} fill className="object-cover" />
              {/* Keterangan judul + kategori agar admin tahu isi tiap foto */}
              <div className="absolute inset-x-0 bottom-0 bg-slate-950/60 px-2 py-1.5">
                <p className="truncate text-xs font-semibold text-white">{item.title}</p>
                {item.category && (
                  <p className="truncate text-[11px] text-slate-300">{item.category}</p>
                )}
              </div>
              <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2 p-2">
                <Button
                  size="sm"
                  variant="secondary"
                  className="bg-white/90 text-slate-900 hover:bg-white"
                  onClick={() => handleEdit({ ...item, id: item.id })}
                >
                  <Edit className="h-3.5 w-3.5" />
                </Button>
                <Button
                  size="sm"
                  variant="secondary"
                  className="bg-red-600/90 text-white hover:bg-red-700"
                  aria-label={`Hapus ${item.title}`}
                  onClick={() =>
                    setPendingDelete({ ...item, id: item.id })
                  }
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      <ConfirmDialog
        open={!!pendingDelete}
        title="Hapus foto ini?"
        description={pendingDelete ? `"${pendingDelete.title}" akan dihapus permanen dan tidak muncul lagi di halaman publik.` : undefined}
        loading={deleteItem.isPending}
        onConfirm={() =>
          pendingDelete &&
          deleteItem.mutate(pendingDelete.id, {
            onSuccess: () => {
              toast.success("Item dihapus");
              setPendingDelete(null);
            },
            onError: () => toast.error("Gagal menghapus"),
          })
        }
        onCancel={() => setPendingDelete(null)}
      />
    </div>
  );
}