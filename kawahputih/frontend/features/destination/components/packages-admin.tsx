"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { usePackages, useCreatePackage, useUpdatePackage, useDeletePackage } from "@/hooks/use-destinations";
import { DataTable } from "@/components/tables/data-table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { CardGridSkeleton } from "@/components/loading/skeleton";
import { EmptyState } from "@/components/empty/empty-state";
import { formatCurrency } from "@/lib/currency";
import { Plus } from "lucide-react";
import { packageSchema, type PackageFormValues } from "../schemas/package.schema";
import type { TourismPackage } from "@/types/destination";

export function PackagesAdmin() {
  const [showForm, setShowForm] = useState(false);
  const { data, isLoading, isError } = usePackages({ limit: 20 });
  const createPackage = useCreatePackage();
  const updatePackage = useUpdatePackage();
  const deletePackage = useDeletePackage();

  const { register, handleSubmit, reset, formState: { errors } } = useForm<PackageFormValues>({
    resolver: zodResolver(packageSchema),
  });

  function onSubmit(values: PackageFormValues) {
    createPackage.mutate(
      {
        name: values.name,
        description: values.description ?? "",
        cover_image: values.cover_image || undefined,
        price_cents: Math.round(values.price * 100),
        duration_hours: values.duration_hours,
        max_capacity: values.max_capacity,
      },
      {
        onSuccess: () => {
          toast.success("Paket wisata dibuat");
          reset();
          setShowForm(false);
        },
        onError: () => toast.error("Gagal membuat paket"),
      }
    );
  }

  function toggleActive(pkg: TourismPackage) {
    updatePackage.mutate(
      {
        id: pkg.id,
        payload: {
          name: pkg.name,
          description: pkg.description,
          cover_image: pkg.cover_image,
          price_cents: pkg.price_cents,
          duration_hours: pkg.duration_hours,
          max_capacity: pkg.max_capacity,
          is_active: !pkg.is_active,
        },
      },
      {
        onSuccess: () => toast.success("Status paket diperbarui"),
        onError: () => toast.error("Gagal memperbarui status"),
      }
    );
  }

  return (
    <div className="space-y-4">
      <Button onClick={() => setShowForm(!showForm)}>
        <Plus className="mr-1 h-4 w-4" /> Paket Baru
      </Button>

      {showForm && (
        <form onSubmit={handleSubmit(onSubmit)} className="max-w-lg space-y-3 rounded-2xl border border-slate-200 p-4 dark:border-slate-800">
          <div>
            <label className="mb-1 block text-sm font-medium">Nama Paket</label>
            <Input {...register("name")} />
            {errors.name && <p className="mt-1 text-xs text-red-600">{errors.name.message}</p>}
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Deskripsi</label>
            <textarea className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm dark:border-slate-700 dark:bg-slate-900" rows={4} {...register("description")} />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">URL Gambar Sampul</label>
            <Input {...register("cover_image")} />
          </div>
          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="mb-1 block text-sm font-medium">Harga (Rp)</label>
              <Input type="number" {...register("price")} />
              {errors.price && <p className="mt-1 text-xs text-red-600">{errors.price.message}</p>}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Durasi (jam)</label>
              <Input type="number" {...register("duration_hours")} />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Kapasitas/hari</label>
              <Input type="number" {...register("max_capacity")} />
              <p className="mt-1 text-xs text-slate-400">0 = tidak terbatas</p>
            </div>
          </div>
          <Button type="submit" loading={createPackage.isPending}>
            {createPackage.isPending ? "Menyimpan..." : "Simpan"}
          </Button>
        </form>
      )}

      {isLoading ? (
        <CardGridSkeleton count={4} />
      ) : isError ? (
        <EmptyState title="Gagal memuat paket" />
      ) : !data?.length ? (
        <EmptyState title="Belum ada paket wisata" />
      ) : (
        <DataTable<TourismPackage>
          rows={data}
          columns={[
            { header: "Nama", cell: (p) => p.name },
            { header: "Harga", cell: (p) => formatCurrency(p.price_cents, p.currency) },
            { header: "Kapasitas", cell: (p) => (p.max_capacity > 0 ? `${p.max_capacity}/hari` : "Tidak terbatas") },
            {
              header: "Status",
              cell: (p) => <span className={p.is_active ? "text-green-600" : "text-slate-400"}>{p.is_active ? "Aktif" : "Nonaktif"}</span>,
            },
            {
              header: "Aksi",
              cell: (p) => (
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" onClick={() => toggleActive(p)} loading={updatePackage.isPending}>
                    {p.is_active ? "Nonaktifkan" : "Aktifkan"}
                  </Button>
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={() =>
                      deletePackage.mutate(p.id, {
                        onSuccess: () => toast.success("Paket dihapus"),
                        onError: () => toast.error("Gagal menghapus paket"),
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
