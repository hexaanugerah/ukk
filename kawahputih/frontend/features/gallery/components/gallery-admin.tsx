"use client";

import { useEffect, useState } from "react";
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
import { Plus, Trash2, Edit, ImagePlus } from "lucide-react";
import { mediaUrl } from "@/lib/media";
import { galleryItemSchema, type GalleryItemFormValues } from "../schemas/gallery.schema";
import type { GalleryItem } from "@/types/gallery";

const MAX_IMAGE_SIZE = 5 * 1024 * 1024; // 5 MB, sama dengan batas backend
const ALLOWED_IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp", "image/gif"];

export function GalleryAdmin() {
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState<string | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  // Item yang menunggu konfirmasi hapus — dialog muncul sebelum delete dibuat.
  const [pendingDelete, setPendingDelete] = useState<GalleryItem | null>(null);
  const { data, isLoading, isError } = useGallery({ limit: 24 });
  const createItem = useCreateGalleryItem();
  const updateItem = useUpdateGalleryItem();
  const deleteItem = useDeleteGalleryItem();

  const { register, handleSubmit, reset, formState: { errors } } = useForm<GalleryItemFormValues>({
    resolver: zodResolver(galleryItemSchema),
  });

  // Bersihkan object URL preview agar tidak bocor memori.
  useEffect(() => {
    return () => {
      if (previewUrl) URL.revokeObjectURL(previewUrl);
    };
  }, [previewUrl]);

  function pickFile(selected: File | null) {
    setFileError(null);
    setFile(null);
    if (previewUrl) URL.revokeObjectURL(previewUrl);
    setPreviewUrl(null);
    if (!selected) return;
    if (!ALLOWED_IMAGE_TYPES.includes(selected.type)) {
      setFileError("Format tidak didukung. Gunakan JPG, PNG, WEBP, atau GIF.");
      return;
    }
    if (selected.size > MAX_IMAGE_SIZE) {
      setFileError("Ukuran file maksimal 5 MB.");
      return;
    }
    setFile(selected);
    setPreviewUrl(URL.createObjectURL(selected));
  }

  function onSubmit(values: GalleryItemFormValues) {
    if (!editingId && !file) {
      setFileError("Pilih file gambar terlebih dahulu.");
      return;
    }
    const formData = new FormData();
    formData.append("title", values.title);
    if (values.category) formData.append("category", values.category);
    if (file) formData.append("image", file);

    if (editingId) {
      updateItem.mutate(
        { id: editingId, payload: formData },
        {
          onSuccess: () => {
            toast.success("Item galeri diperbarui");
            handleCancel();
          },
          onError: () => toast.error("Gagal memperbarui item"),
        }
      );
    } else {
      createItem.mutate(formData, {
        onSuccess: () => {
          toast.success("Item galeri ditambahkan");
          handleCancel();
        },
        onError: () => toast.error("Gagal menambahkan item"),
      });
    }
  }

  function handleEdit(item: GalleryItem) {
    setEditingId(item.id);
    setShowForm(true);
    setFile(null);
    setFileError(null);
    if (previewUrl) URL.revokeObjectURL(previewUrl);
    setPreviewUrl(mediaUrl(item.image_url) ?? null);
    reset({ title: item.title, category: item.category });
  }

  function handleCancel() {
    setShowForm(false);
    setEditingId(null);
    setFile(null);
    setFileError(null);
    if (previewUrl) URL.revokeObjectURL(previewUrl);
    setPreviewUrl(null);
    reset();
  }

  const submitting = createItem.isPending || updateItem.isPending;

  return (
    <div className="space-y-4">
      <Button onClick={() => (showForm ? handleCancel() : setShowForm(true))}>
        <Plus className="mr-1 h-4 w-4" /> {showForm ? "Tutup Form" : "Tambah Foto"}
      </Button>
      {showForm && (
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
            <label className="mb-1 block text-sm font-medium">
              Gambar {editingId ? "(pilih file baru untuk mengganti)" : "(upload file)"}
            </label>
            <label className="flex cursor-pointer items-center gap-2 rounded-xl border border-dashed border-slate-300 px-3 py-2 text-sm text-slate-600 hover:border-slate-400 dark:border-slate-700 dark:text-slate-300">
              <ImagePlus className="h-4 w-4" />
              <span className="truncate">{file ? file.name : "Pilih file gambar (JPG/PNG/WEBP/GIF, maks 5 MB)"}</span>
              <input
                type="file"
                accept="image/jpeg,image/png,image/webp,image/gif"
                className="hidden"
                onChange={(e) => pickFile(e.target.files?.[0] ?? null)}
              />
            </label>
            {fileError && <p className="mt-1 text-xs text-red-600">{fileError}</p>}
            {previewUrl && (
              <div className="relative mt-2 h-32 w-32 overflow-hidden rounded-lg border border-slate-200 dark:border-slate-700">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src={previewUrl} alt="Pratinjau gambar" className="h-full w-full object-cover" />
              </div>
            )}
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Kategori</label>
            <Input placeholder="mis. Pemandangan, Fasilitas" {...register("category")} />
          </div>
          <Button type="submit" loading={submitting}>
            {editingId ? (submitting ? "Menyimpan..." : "Simpan Perubahan") : (submitting ? "Mengunggah..." : "Simpan")}
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
              {mediaUrl(item.image_url) && (
                <Image src={mediaUrl(item.image_url)!} alt={item.title} fill className="object-cover" />
              )}
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
                  onClick={() => handleEdit(item)}
                >
                  <Edit className="h-3.5 w-3.5" />
                </Button>
                <Button
                  size="sm"
                  variant="secondary"
                  className="bg-red-600/90 text-white hover:bg-red-700"
                  aria-label={`Hapus ${item.title}`}
                  onClick={() => setPendingDelete(item)}
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
