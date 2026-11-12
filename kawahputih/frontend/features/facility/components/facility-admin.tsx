"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Select } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/empty/empty-state";
import { Skeleton } from "@/components/loading/skeleton";
import { DataTable } from "@/components/tables/data-table";
import { Plus, Edit, Trash2, Image as ImageIcon } from "lucide-react";
import Image from "next/image";
import { facilitySchema, type FacilityFormValues } from "../schemas/facility.schema";
import type { Facility } from "@/types/facility";
import { FACILITY_STATUSES } from "@/types/facility";
import { useFacilityList, usePublishedFacilityList, useCreateFacility, useUpdateFacility, useDeleteFacility } from "@/hooks/use-facility";

export function FacilityAdmin() {
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [previewImage, setPreviewImage] = useState<string | null>(null);
  // Item yang menunggu konfirmasi hapus — dialog muncul sebelum delete dibuat.
  const [pendingDelete, setPendingDelete] = useState<Facility | null>(null);
  const { data, isLoading, isError } = useFacilityList({ limit: 50 });
  const publishedFacilities = usePublishedFacilityList();
  const createItem = useCreateFacility();
  const updateItem = useUpdateFacility();
  const deleteItem = useDeleteFacility();

  const { register, handleSubmit, reset, watch, formState: { errors } } = useForm<FacilityFormValues>({
    resolver: zodResolver(facilitySchema),
    defaultValues: {
      status: "active",
      is_published: false,
      sort_order: 0,
      icon: "",
      image_url: "",
      description: "",
      location: "",
      operating_hours: "",
    },
  });

  const watchedImageUrl = watch("image_url");

  function onSubmit(values: FacilityFormValues) {
    if (editingId) {
      updateItem.mutate(
        { id: editingId, payload: values },
        {
          onSuccess: () => {
            toast.success("Fasilitas berhasil diperbarui");
            reset();
            setShowForm(false);
            setEditingId(null);
            setPreviewImage(null);
          },
          onError: () => toast.error("Gagal memperbarui fasilitas"),
        }
      );
    } else {
      createItem.mutate(values, {
        onSuccess: () => {
          toast.success("Fasilitas berhasil dibuat");
          reset();
          setShowForm(false);
          setPreviewImage(null);
        },
        onError: () => toast.error("Gagal membuat fasilitas"),
      });
    }
  }

  function handleEdit(item: Facility) {
    setEditingId(item.id);
    setShowForm(true);
    setPreviewImage(item.image_url || null);
    reset({
      name: item.name,
      description: item.description,
      icon: item.icon,
      image_url: item.image_url,
      location: item.location,
      operating_hours: item.operating_hours,
      status: item.status,
      is_published: item.is_published,
      sort_order: item.sort_order,
    });
  }

  function handleCancel() {
    setShowForm(false);
    setEditingId(null);
    setPreviewImage(null);
    reset({
      status: "active",
      is_published: false,
      sort_order: 0,
      icon: "",
      image_url: "",
      description: "",
      location: "",
      operating_hours: "",
    });
  }

  const publishedCount = publishedFacilities.data?.filter(f => f.is_published).length ?? 0;
  const totalCount = data?.length ?? 0;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-4 flex-wrap">
        <Button onClick={() => { handleCancel(); setShowForm(!showForm); }}>
          <Plus className="mr-1 h-4 w-4" /> {showForm ? "Tutup Form" : "Tambah Fasilitas"}
        </Button>
        <div className="flex items-center gap-4 text-sm text-slate-600 dark:text-slate-400">
          <span>Total: <strong className="text-slate-900 dark:text-white">{totalCount}</strong></span>
          <span>Dipublikasikan: <strong className="text-green-600 dark:text-green-400">{publishedCount}</strong></span>
          <span>Draft: <strong className="text-amber-600 dark:text-amber-400">{totalCount - publishedCount}</strong></span>
        </div>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit(onSubmit)} className="max-w-3xl space-y-4 rounded-2xl border border-slate-200 p-6 dark:border-slate-800">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">{editingId ? "Edit Fasilitas" : "Tambah Fasilitas Baru"}</h3>
            <Button type="button" variant="ghost" size="sm" onClick={handleCancel}>Batal</Button>
          </div>
          
          <div className="space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium">Nama Fasilitas *</label>
              <Input {...register("name")} />
              {errors.name && <p className="mt-1 text-xs text-red-600">{errors.name.message}</p>}
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <label className="mb-1 block text-sm font-medium">Ikon (Lucide icon name)</label>
                <Input placeholder="Car, Coffee, ShieldCheck..." {...register("icon")} />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">URL Gambar</label>
                <Input placeholder="https://..." {...register("image_url")} />
                {errors.image_url && <p className="mt-1 text-xs text-red-600">{errors.image_url.message}</p>}
              </div>
            </div>
            
            {watchedImageUrl && (
              <div className="relative aspect-video w-full max-w-md overflow-hidden rounded-xl bg-slate-100 dark:bg-slate-800">
                <Image src={watchedImageUrl} alt="Preview" fill className="object-cover" />
              </div>
            )}
            
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <label className="mb-1 block text-sm font-medium">Lokasi Fasilitas</label>
                <Input {...register("location")} />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Jam Operasional</label>
                <Input {...register("operating_hours")} />
              </div>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Deskripsi</label>
              <Textarea rows={3} {...register("description")} />
            </div>
            <div className="grid gap-4 md:grid-cols-3">
              <div>
                <label className="mb-1 block text-sm font-medium">Status *</label>
                <Select {...register("status")}>
                  {FACILITY_STATUSES.map((status) => (
                    <option key={status} value={status}>{status.charAt(0).toUpperCase() + status.slice(1)}</option>
                  ))}
                </Select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Urutan Tampilan</label>
                <Input type="number" min={0} {...register("sort_order", { valueAsNumber: true })} />
              </div>
              <div className="flex items-end">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" {...register("is_published")} className="rounded border-slate-300" />
                  <span className="text-sm">Publikasikan</span>
                </label>
              </div>
            </div>
          </div>

          <div className="flex gap-2 pt-4 border-t">
            <Button type="submit" loading={createItem.isPending || updateItem.isPending}>
              {editingId ? (updateItem.isPending ? "Menyimpan..." : "Simpan Perubahan") : (createItem.isPending ? "Menyimpan..." : "Simpan Fasilitas")}
            </Button>
            <Button type="button" variant="outline" onClick={handleCancel}>Batal</Button>
          </div>
        </form>
      )}

      {isLoading ? (
        <Skeleton className="h-64 w-full" />
      ) : isError ? (
        <EmptyState title="Gagal memuat fasilitas" />
      ) : !data?.length ? (
        <EmptyState title="Belum ada fasilitas" />
      ) : (
        <DataTable<Facility>
          rows={data}
          columns={[
            { 
              header: "Fasilitas", 
              cell: (f) => (
                <div className="flex items-center gap-3">
                  {f.image_url && (
                    <Image src={f.image_url} alt={f.name} width={40} height={40} className="rounded-lg object-cover" />
                  )}
                  <div>
                    <p className="font-medium">{f.name}</p>
                    <p className="text-xs text-slate-500">{f.location || "Tanpa lokasi"}</p>
                  </div>
                </div>
              ) 
            },
            { header: "Status", cell: (f) => <Badge variant="outline">{f.status}</Badge> },
            { 
              header: "Publikasi", 
              cell: (f) => <Badge status={f.is_published ? "success" : "default"}>{f.is_published ? "Ya" : "Tidak"}</Badge> 
            },
            { header: "Urutan", cell: (f) => f.sort_order },
            {
              header: "Aksi",
              cell: (f) => (
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleEdit(f)}
                  >
                    <Edit className="mr-1 h-3.5 w-3.5" /> Edit
                  </Button>
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={() => setPendingDelete(f)}
                  >
                    <Trash2 className="mr-1 h-3.5 w-3.5" /> Hapus
                  </Button>
                </div>
              ),
            },
          ]}
        />
      )}

      <ConfirmDialog
        open={!!pendingDelete}
        title="Hapus fasilitas ini?"
        description={pendingDelete ? `"${pendingDelete.name}" akan dihapus permanen dan tidak muncul lagi di halaman publik.` : undefined}
        loading={deleteItem.isPending}
        onConfirm={() =>
          pendingDelete &&
          deleteItem.mutate(pendingDelete.id, {
            onSuccess: () => {
              toast.success("Fasilitas dihapus");
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