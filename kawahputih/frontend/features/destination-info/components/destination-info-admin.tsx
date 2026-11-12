"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/empty/empty-state";
import { Skeleton } from "@/components/loading/skeleton";
import { DataTable } from "@/components/tables/data-table";
import { Plus, Edit, Trash2, Eye, Archive, Send } from "lucide-react";
import { destinationInfoSchema, type DestinationInfoFormValues } from "../schemas/destination-info.schema";
import type { DestinationInfo } from "@/types/destination-info";
import { useDestinationInfoList, usePublishedDestinationInfo, useCreateDestinationInfo, useUpdateDestinationInfo, usePublishDestinationInfo, useArchiveDestinationInfo, useDeleteDestinationInfo } from "@/hooks/use-destination-info";

export function DestinationInfoAdmin() {
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  // Item yang menunggu konfirmasi hapus — dialog muncul sebelum delete dibuat.
  const [pendingDelete, setPendingDelete] = useState<DestinationInfo | null>(null);
  const { data, isLoading, isError } = useDestinationInfoList({ limit: 20 });
  const publishedInfo = usePublishedDestinationInfo();
  const createItem = useCreateDestinationInfo();
  const updateItem = useUpdateDestinationInfo();
  const publishItem = usePublishDestinationInfo();
  const archiveItem = useArchiveDestinationInfo();
  const deleteItem = useDeleteDestinationInfo();

  const { register, handleSubmit, reset, setValue, formState: { errors } } = useForm<DestinationInfoFormValues>({
    resolver: zodResolver(destinationInfoSchema),
    defaultValues: {
      history: "",
      destination_desc: "",
      location: "",
      operating_hours: "",
      contact: "",
      additional_info: "",
    },
  });

  function onSubmit(values: DestinationInfoFormValues) {
    if (editingId) {
      updateItem.mutate(
        { id: editingId, payload: values },
        {
          onSuccess: () => {
            toast.success("Informasi berhasil diperbarui");
            reset();
            setShowForm(false);
            setEditingId(null);
          },
          onError: () => toast.error("Gagal memperbarui informasi"),
        }
      );
    } else {
      createItem.mutate(values, {
        onSuccess: () => {
          toast.success("Informasi berhasil dibuat sebagai draft");
          reset();
          setShowForm(false);
        },
        onError: () => toast.error("Gagal membuat informasi"),
      });
    }
  }

  function handleEdit(item: DestinationInfo) {
    setEditingId(item.id);
    setShowForm(true);
    reset({
      title: item.title,
      description: item.description,
      history: item.history || "",
      destination_desc: item.destination_desc || "",
      location: item.location || "",
      operating_hours: item.operating_hours || "",
      contact: item.contact || "",
      additional_info: item.additional_info || "",
    });
  }

  function handleCancel() {
    setShowForm(false);
    setEditingId(null);
    reset();
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-4">
        <Button onClick={() => { handleCancel(); setShowForm(!showForm); }}>
          <Plus className="mr-1 h-4 w-4" /> {showForm ? "Tutup Form" : "Tambah Informasi"}
        </Button>
        {publishedInfo.data && (
          <Badge status={publishedInfo.data.status === "published" ? "success" : "default"}>
            {publishedInfo.data.status === "published" ? "Sudah Dipublikasikan" : "Belum Dipublikasikan"}
          </Badge>
        )}
      </div>

      {showForm && (
        <form onSubmit={handleSubmit(onSubmit)} className="max-w-4xl space-y-4 rounded-2xl border border-slate-200 p-6 dark:border-slate-800">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">{editingId ? "Edit Informasi" : "Tambah Informasi Baru"}</h3>
            <Button type="button" variant="ghost" size="sm" onClick={handleCancel}>Batal</Button>
          </div>
          
          <div className="grid gap-4 md:grid-cols-2">
            <div className="md:col-span-2">
              <label className="mb-1 block text-sm font-medium">Judul Informasi *</label>
              <Input {...register("title")} />
              {errors.title && <p className="mt-1 text-xs text-red-600">{errors.title.message}</p>}
            </div>
            <div className="md:col-span-2">
              <label className="mb-1 block text-sm font-medium">Deskripsi Utama *</label>
              <Textarea rows={4} {...register("description")} />
              {errors.description && <p className="mt-1 text-xs text-red-600">{errors.description.message}</p>}
            </div>
            <div className="md:col-span-2">
              <label className="mb-1 block text-sm font-medium">Sejarah Kawah Putih</label>
              <Textarea rows={4} {...register("history")} />
            </div>
            <div className="md:col-span-2">
              <label className="mb-1 block text-sm font-medium">Deskripsi Destinasi</label>
              <Textarea rows={4} {...register("destination_desc")} />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Lokasi</label>
              <Input {...register("location")} />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Jam Operasional</label>
              <Input {...register("operating_hours")} />
            </div>
            <div className="md:col-span-2">
              <label className="mb-1 block text-sm font-medium">Kontak</label>
              <Input {...register("contact")} />
            </div>
            <div className="md:col-span-2">
              <label className="mb-1 block text-sm font-medium">Informasi Tambahan</label>
              <Textarea rows={3} {...register("additional_info")} />
            </div>
          </div>

          <div className="flex gap-2 pt-4 border-t">
            <Button type="submit" loading={createItem.isPending || updateItem.isPending}>
              {editingId ? (updateItem.isPending ? "Menyimpan..." : "Simpan Perubahan") : (createItem.isPending ? "Menyimpan..." : "Simpan sebagai Draft")}
            </Button>
            <Button type="button" variant="outline" onClick={handleCancel}>Batal</Button>
          </div>
        </form>
      )}

      {isLoading ? (
        <Skeleton className="h-64 w-full" />
      ) : isError ? (
        <EmptyState title="Gagal memuat informasi" />
      ) : !data?.length ? (
        <EmptyState title="Belum ada informasi" />
      ) : (
        <DataTable<DestinationInfo>
          rows={data}
          columns={[
            { header: "Judul", cell: (d) => d.title },
            { 
              header: "Status", 
              cell: (d) => <Badge status={d.status === "published" ? "success" : d.status === "archived" ? "warning" : "default"}>{d.status}</Badge> 
            },
            { header: "Terakhir Diperbarui", cell: (d) => new Date(d.updated_at).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" }) },
            {
              header: "Aksi",
              cell: (d) => (
                <div className="flex gap-2">
                  {d.status === "draft" && (
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() =>
                        publishItem.mutate(d.id, {
                          onSuccess: () => toast.success("Informasi dipublikasikan"),
                          onError: () => toast.error("Gagal mempublikasikan"),
                        })
                      }
                    >
                      <Send className="mr-1 h-3.5 w-3.5" /> Publikasikan
                    </Button>
                  )}
                  {d.status === "published" && (
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() =>
                        archiveItem.mutate(d.id, {
                          onSuccess: () => toast.success("Informasi diarsipkan"),
                          onError: () => toast.error("Gagal mengarsipkan"),
                        })
                      }
                    >
                      <Archive className="mr-1 h-3.5 w-3.5" /> Arsipkan
                    </Button>
                  )}
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleEdit(d)}
                  >
                    <Edit className="mr-1 h-3.5 w-3.5" /> Edit
                  </Button>
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={() => setPendingDelete(d)}
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
        title="Hapus informasi destinasi ini?"
        description={pendingDelete ? `"${pendingDelete.title}" akan dihapus permanen dan tidak muncul lagi di halaman publik.` : undefined}
        loading={deleteItem.isPending}
        onConfirm={() =>
          pendingDelete &&
          deleteItem.mutate(pendingDelete.id, {
            onSuccess: () => {
              toast.success("Informasi dihapus");
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