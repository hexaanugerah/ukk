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
import { Plus, Edit, Trash2, Shield, AlertTriangle, Flag } from "lucide-react";
import { visitorRuleSchema, type VisitorRuleFormValues } from "../schemas/visitor-rule.schema";
import type { VisitorRule } from "@/types/visitor-rule";
import { VISITOR_RULE_CATEGORIES, VISITOR_RULE_PRIORITIES } from "@/types/visitor-rule";
import { useVisitorRuleList, usePublishedVisitorRuleList, useCreateVisitorRule, useUpdateVisitorRule, useDeleteVisitorRule } from "@/hooks/use-visitor-rule";

const PRIORITY_ICONS = {
  low: Flag,
  normal: Shield,
  high: AlertTriangle,
  critical: AlertTriangle,
};

export function VisitorRuleAdmin() {
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  // Item yang menunggu konfirmasi hapus — dialog muncul sebelum delete dibuat.
  const [pendingDelete, setPendingDelete] = useState<VisitorRule | null>(null);
  const { data, isLoading, isError } = useVisitorRuleList({ limit: 50 });
  const publishedRules = usePublishedVisitorRuleList();
  const createItem = useCreateVisitorRule();
  const updateItem = useUpdateVisitorRule();
  const deleteItem = useDeleteVisitorRule();

  const { register, handleSubmit, reset, formState: { errors } } = useForm<VisitorRuleFormValues>({
    resolver: zodResolver(visitorRuleSchema),
    defaultValues: {
      priority: "normal",
      sort_order: 0,
      is_published: false,
    },
  });

  function onSubmit(values: VisitorRuleFormValues) {
    if (editingId) {
      updateItem.mutate(
        { id: editingId, payload: values },
        {
          onSuccess: () => {
            toast.success("Aturan kunjungan berhasil diperbarui");
            reset();
            setShowForm(false);
            setEditingId(null);
          },
          onError: () => toast.error("Gagal memperbarui aturan"),
        }
      );
    } else {
      createItem.mutate(values, {
        onSuccess: () => {
          toast.success("Aturan kunjungan berhasil dibuat");
          reset();
          setShowForm(false);
        },
        onError: () => toast.error("Gagal membuat aturan"),
      });
    }
  }

  function handleEdit(item: VisitorRule) {
    setEditingId(item.id);
    setShowForm(true);
    reset({
      title: item.title,
      description: item.description,
      category: item.category as VisitorRuleFormValues["category"],
      priority: item.priority,
      sort_order: item.sort_order,
      is_published: item.is_published,
    });
  }

  function handleCancel() {
    setShowForm(false);
    setEditingId(null);
    reset({
      priority: "normal",
      sort_order: 0,
      is_published: false,
    });
  }

  const publishedCount = publishedRules.data?.filter(r => r.is_published).length ?? 0;
  const totalCount = data?.length ?? 0;

  function PriorityBadge({ priority }: { priority: VisitorRule["priority"] }) {
    const Icon = PRIORITY_ICONS[priority];
    return (
      <Badge status={priority}>
        <Icon className="mr-1 h-3 w-3" />
        {priority.charAt(0).toUpperCase() + priority.slice(1)}
      </Badge>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-4 flex-wrap">
        <Button onClick={() => { handleCancel(); setShowForm(!showForm); }}>
          <Plus className="mr-1 h-4 w-4" /> {showForm ? "Tutup Form" : "Tambah Aturan"}
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
            <h3 className="text-lg font-semibold">{editingId ? "Edit Aturan Kunjungan" : "Tambah Aturan Baru"}</h3>
            <Button type="button" variant="ghost" size="sm" onClick={handleCancel}>Batal</Button>
          </div>
          
          <div className="space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium">Judul Aturan *</label>
              <Input {...register("title")} />
              {errors.title && <p className="mt-1 text-xs text-red-600">{errors.title.message}</p>}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Deskripsi Aturan *</label>
              <Textarea rows={4} {...register("description")} />
              {errors.description && <p className="mt-1 text-xs text-red-600">{errors.description.message}</p>}
            </div>
            <div className="grid gap-4 md:grid-cols-3">
              <div>
                <label className="mb-1 block text-sm font-medium">Kategori *</label>
                <Select {...register("category")}>
                  {VISITOR_RULE_CATEGORIES.map((cat) => (
                    <option key={cat} value={cat}>{cat}</option>
                  ))}
                </Select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Prioritas *</label>
                <Select {...register("priority")}>
                  {VISITOR_RULE_PRIORITIES.map((priority) => (
                    <option key={priority} value={priority}>
                      {priority.charAt(0).toUpperCase() + priority.slice(1)}
                    </option>
                  ))}
                </Select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Urutan Tampilan</label>
                <Input type="number" min={0} {...register("sort_order", { valueAsNumber: true })} />
              </div>
            </div>
            <div className="flex items-end">
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" {...register("is_published")} className="rounded border-slate-300" />
                <span className="text-sm">Publikasikan</span>
              </label>
            </div>
          </div>

          <div className="flex gap-2 pt-4 border-t">
            <Button type="submit" loading={createItem.isPending || updateItem.isPending}>
              {editingId ? (updateItem.isPending ? "Menyimpan..." : "Simpan Perubahan") : (createItem.isPending ? "Menyimpan..." : "Simpan Aturan")}
            </Button>
            <Button type="button" variant="outline" onClick={handleCancel}>Batal</Button>
          </div>
        </form>
      )}

      {isLoading ? (
        <Skeleton className="h-64 w-full" />
      ) : isError ? (
        <EmptyState title="Gagal memuat aturan kunjungan" />
      ) : !data?.length ? (
        <EmptyState title="Belum ada aturan kunjungan" />
      ) : (
        <DataTable<VisitorRule>
          rows={data}
          columns={[
            { header: "Judul Aturan", cell: (r) => r.title },
            { header: "Kategori", cell: (r) => <Badge variant="outline">{r.category}</Badge> },
            { header: "Prioritas", cell: (r) => <PriorityBadge priority={r.priority} /> },
            { 
              header: "Status", 
              cell: (r) => <Badge status={r.is_published ? "success" : "default"}>{r.is_published ? "Dipublikasikan" : "Draft"}</Badge> 
            },
            { header: "Urutan", cell: (r) => r.sort_order },
            {
              header: "Aksi",
              cell: (r) => (
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleEdit(r)}
                  >
                    <Edit className="mr-1 h-3.5 w-3.5" /> Edit
                  </Button>
                  <Button
                    size="sm"
                    variant="destructive"
                    onClick={() => setPendingDelete(r)}
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
        title="Hapus aturan kunjungan ini?"
        description={pendingDelete ? `"${pendingDelete.title}" akan dihapus permanen dan tidak muncul lagi di halaman publik.` : undefined}
        loading={deleteItem.isPending}
        onConfirm={() =>
          pendingDelete &&
          deleteItem.mutate(pendingDelete.id, {
            onSuccess: () => {
              toast.success("Aturan dihapus");
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