"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Eye, Pencil, Plus, Power, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { SearchInput } from "@/components/panel/panel-toolbar";
import { ticketService, TicketError } from "@/services/ticket.api.service";
import { rupiah, paginate, searchIn, sortData } from "@/lib/business";
import type { Ticket, TicketCategory } from "@/types/api";

// Kelola tiket — CRUD via backend API.

const emptyForm = {
  id: "",
  code: "",
  category: "ENTRANCE" as TicketCategory,
  name: "",
  description: "",
  price: 0,
  capacity: 0,
  status: "AKTIF" as "AKTIF" | "NONAKTIF",
};

export default function AdminTicketsPage() {
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [loading, setLoading] = useState(true);

  const [q, setQ] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [sortKey, setSortKey] = useState<"name" | "price" | "capacity">("name");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("asc");
  const [page, setPage] = useState(1);
  const [form, setForm] = useState<typeof emptyForm | null>(null);
  const [detail, setDetail] = useState<Ticket | null>(null);
  const [delTarget, setDelTarget] = useState<Ticket | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const loadTickets = useCallback(async () => {
    setLoading(true);
    try {
      const data = await ticketService.list({ status: statusFilter || undefined });
      setTickets(data);
    } catch (error: any) {
      toast.error("Gagal memuat tiket", { description: error.message });
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  // Muat tiket saat mount dan ketika filter status berubah.
  useEffect(() => {
    loadTickets();
  }, [loadTickets]);

  function pageSize(total: number) {
    return total > 30 ? 10 : 8;
  }

  // Perhitungan yang di-memoize (harus sebelum early return)
  const rows = useMemo(() => {
    let list = searchIn(tickets, q, ["name", "description"]);
    if (statusFilter) list = list.filter((t) => t.status === statusFilter);
    return sortData(list, sortKey, sortDir);
  }, [tickets, q, statusFilter, sortKey, sortDir]);

  const paged = useMemo(() => paginate(rows, page, pageSize(rows.length)), [rows, page]);

  if (loading) {
    return (
      <div>
        <PageHeader
          title="Kelola Tiket"
          subtitle="Kelola jenis tiket dan harga"
          action={
            <button
              onClick={openAdd}
              className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f]"
            >
              <Plus className="h-4 w-4" /> Tambah Tiket
            </button>
          }
        />
        <SectionCard title="Daftar Jenis Tiket" action={<span className="text-sm font-semibold text-[#1768ad]">Loading...</span>}>
          <div className="animate-pulse space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-14 bg-gray-200 rounded border border-[#e5edf3]" />
            ))}
          </div>
        </SectionCard>
      </div>
    );
  }

  function openAdd() {
    setForm({ ...emptyForm });
    setErrors({});
  }

  function openEdit(t: Ticket) {
    setForm({
      id: t.id,
      code: t.code,
      category: t.category,
      name: t.name,
      description: t.description,
      price: t.price,
      capacity: t.capacity,
      status: t.status,
    });
    setErrors({});
  }

  async function save() {
    if (!form) return;
    const errs: Record<string, string> = {};
    if (form.name.trim().length < 3) errs.name = "Nama minimal 3 karakter";
    if (!Number.isFinite(form.price) || form.price <= 0) errs.price = "Harga harus lebih dari 0";
    if (!Number.isFinite(form.capacity) || form.capacity <= 0) errs.capacity = "Kapasitas harus lebih dari 0";
    setErrors(errs);
    if (Object.keys(errs).length > 0) return;

    setSubmitting(true);
    try {
      if (form.id) {
        await ticketService.update(form.id, {
          code: form.code, // pertahankan kode tiket yang ada
          name: form.name.trim(),
          description: form.description.trim(),
          category: form.category, // pertahankan kategori yang ada
          price: Math.round(form.price),
          capacity: Math.round(form.capacity),
          status: form.status,
        });
        toast.success("Tiket diperbarui");
      } else {
        // Buat kode dari nama
        const code = "TKT-" + form.name.trim().toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 10);
        await ticketService.create({
          code,
          name: form.name.trim(),
          description: form.description.trim(),
          category: "ENTRANCE",
          price: Math.round(form.price),
          capacity: Math.round(form.capacity),
          status: form.status,
          is_active: form.status === "AKTIF",
        });
        toast.success("Tiket baru ditambahkan");
      }
      setForm(null);
      loadTickets();
    } catch (err: any) {
      const message = err instanceof TicketError ? err.message : err.message || "Gagal menyimpan tiket";
      toast.error(message);
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete() {
    if (!delTarget) return;
    setDeleting(true);
    try {
      await ticketService.delete(delTarget.id);
      toast.success(`${delTarget.name} dihapus`);
      setDelTarget(null);
      loadTickets();
    } catch (err: any) {
      toast.error("Gagal menghapus tiket", { description: err.message });
    } finally {
      setDeleting(false);
    }
  }

  async function toggleStatus(t: Ticket) {
    try {
      await ticketService.updateStatus(t.id, t.status === "AKTIF" ? "NONAKTIF" : "AKTIF");
      toast.success(`${t.name} ${t.status === "AKTIF" ? "dinonaktifkan" : "diaktifkan"}`);
      loadTickets();
    } catch (err: any) {
      toast.error(err.message || "Gagal mengubah status");
    }
  }

  async function handleStatusFilterChange(status: string) {
    setStatusFilter(status);
    setPage(1);
    loadTickets();
  }

  async function handleSearchChange(value: string) {
    setQ(value);
    setPage(1);
  }

  return (
    <div>
      <PageHeader
        title="Kelola Tiket"
        subtitle="Kelola jenis tiket dan harga"
        action={
          <button
            onClick={openAdd}
            className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f]"
          >
            <Plus className="h-4 w-4" /> Tambah Tiket
          </button>
        }
      />

      <SectionCard title="Daftar Jenis Tiket" action={<span className="text-sm font-semibold text-[#1768ad]">{rows.length} data</span>}>
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <SearchInput value={q} onChange={handleSearchChange} placeholder="Cari nama tiket…" className="w-full sm:max-w-xs" />
          <select
            value={statusFilter}
            onChange={(e) => handleStatusFilterChange(e.target.value)}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter status tiket"
          >
            <option value="">Semua status</option>
            <option value="AKTIF">Aktif</option>
            <option value="NONAKTIF">Nonaktif</option>
          </select>
        </div>

        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>NO</th>
                <th>Nama Tiket</th>
                <th>
                  <button className="font-bold uppercase" onClick={() => { setSortKey("price"); setSortDir(sortKey === "price" && sortDir === "asc" ? "desc" : "asc"); }}>
                    Harga {sortKey === "price" && (sortDir === "asc" ? "↑" : "↓")}
                  </button>
                </th>
                <th>Status</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {paged.rows.length === 0 ? (
                <tr><td colSpan={5} className="py-10 text-center text-sm text-[#94A3B8]">Belum ada jenis tiket.</td></tr>
              ) : (
                paged.rows.map((t, index) => (
                  <tr key={t.id}>
                    <td className="tabular-nums text-[#698097]">{(paged.page - 1) * pageSize(rows.length) + index + 1}.</td>
                    <td>
                      <p className="font-medium text-[#17324d]">{t.name}</p>
                      <p className="text-xs text-[#698097]">{t.description}</p>
                    </td>
                    <td className="tabular-nums">{rupiah(t.price)}</td>
                    <td><StatusPill status={t.status === "AKTIF" ? "Aktif" : "Nonaktif"} /></td>
                    <td>
                      <div className="flex items-center gap-1.5">
                        <button onClick={() => setDetail(t)} className="grid h-7 w-7 place-items-center rounded-md border border-[#d7e3ed] text-[#29445e] hover:bg-[#f2f7fb]" aria-label={`Detail ${t.name}`}>
                          <Eye className="h-3.5 w-3.5" />
                        </button>
                        <button onClick={() => openEdit(t)} className="grid h-7 w-7 place-items-center rounded-md border border-[#d7e3ed] text-[#29445e] hover:bg-[#f2f7fb]" aria-label={`Edit ${t.name}`}>
                          <Pencil className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => toggleStatus(t)}
                          className={`grid h-7 w-7 place-items-center rounded-md border ${t.status === "AKTIF" ? "border-[#fdeaea] text-[#c53030]" : "border-[#e6f7f0] text-[#16876a]"} hover:bg-[#f7fbff}`}
                          aria-label={t.status === "AKTIF" ? `Nonaktifkan ${t.name}` : `Aktifkan ${t.name}`}
                        >
                          <Power className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => setDelTarget(t)}
                          className="grid h-7 w-7 place-items-center rounded-md border border-[#fdeaea] text-[#c53030] hover:bg-[#fdeaea]"
                          aria-label={`Hapus ${t.name}`}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        <TablePagination page={paged.page} totalPages={paged.totalPages} onPage={setPage} />
      </SectionCard>

      {/* Form dialog */}
      {form && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" onClick={() => setForm(null)} role="dialog" aria-modal="true">
          <div className="w-full max-w-md rounded-md bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-base font-bold text-[#17324d]">{form.id ? "Edit Tiket" : "Tambah Tiket"}</h3>
            <div className="mt-4 space-y-3">
              <label className="block">
                <span className="mb-1 block text-xs font-semibold text-[#29445e]">Nama Tiket*</span>
                <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} className="h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm outline-none focus:border-[#1768ad]" />
                {errors.name && <span className="text-xs text-red-600">{errors.name}</span>}
              </label>
              <label className="block">
                <span className="mb-1 block text-xs font-semibold text-[#29445e]">Deskripsi</span>
                <input value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} className="h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm outline-none focus:border-[#1768ad]" />
              </label>
              <div className="grid grid-cols-2 gap-3">
                <label className="block">
                  <span className="mb-1 block text-xs font-semibold text-[#29445e]">Harga (Rp)*</span>
                  <input type="number" min={0} value={form.price || ""} onChange={(e) => setForm({ ...form, price: Number(e.target.value) })} className="h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm outline-none focus:border-[#1768ad]" />
                  {errors.price && <span className="text-xs text-red-600">{errors.price}</span>}
                </label>
                <label className="block">
                  <span className="mb-1 block text-xs font-semibold text-[#29445e]">Kapasitas*</span>
                  <input type="number" min={0} value={form.capacity || ""} onChange={(e) => setForm({ ...form, capacity: Number(e.target.value) })} className="h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm outline-none focus:border-[#1768ad]" />
                  {errors.capacity && <span className="text-xs text-red-600">{errors.capacity}</span>}
                </label>
              </div>
              <label className="block">
                <span className="mb-1 block text-xs font-semibold text-[#29445e]">Status</span>
                <select value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value as "AKTIF" | "NONAKTIF" })} className="h-9 w-full rounded-md border border-[#d7e3ed] bg-white px-3 text-sm outline-none focus:border-[#1768ad]">
                  <option value="AKTIF">Aktif</option>
                  <option value="NONAKTIF">Nonaktif</option>
                </select>
              </label>
            </div>
            <div className="mt-5 flex justify-end gap-2">
              <button onClick={() => setForm(null)} className="h-9 rounded-md border border-[#d7e3ed] px-4 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]">Batal</button>
              <button onClick={save} disabled={submitting} className="h-9 rounded-md bg-[#1768ad] px-4 text-sm font-semibold text-white hover:bg-[#14548f] disabled:opacity-60">
                {submitting ? "Menyimpan..." : "Simpan"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Konfirmasi hapus */}
      {delTarget && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" onClick={() => setDelTarget(null)} role="dialog" aria-modal="true">
          <div className="w-full max-w-sm rounded-md bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-base font-bold text-[#17324d]">Hapus tiket ini?</h3>
            <p className="mt-2 text-sm text-[#698097]">
              <span className="font-semibold text-[#17324d]">{delTarget.name}</span> akan dihapus permanen. Tindakan ini tidak bisa dibatalkan.
            </p>
            <div className="mt-5 flex justify-end gap-2">
              <button onClick={() => setDelTarget(null)} className="h-9 rounded-md border border-[#d7e3ed] px-4 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]">Batal</button>
              <button onClick={handleDelete} disabled={deleting} className="h-9 rounded-md bg-[#c53030] px-4 text-sm font-semibold text-white hover:brightness-95 disabled:opacity-60">
                {deleting ? "Menghapus..." : "Hapus"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Detail dialog */}
      {detail && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" onClick={() => setDetail(null)} role="dialog" aria-modal="true">
          <div className="w-full max-w-md rounded-md bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between border-b border-[#e5edf3] pb-3">
              <h3 className="text-base font-bold text-[#17324d]">{detail.name}</h3>
              <button onClick={() => setDetail(null)} className="text-[#698097] hover:text-[#17324d]" aria-label="Tutup">✕</button>
            </div>
            <dl className="mt-3 space-y-2 text-sm">
              <div className="flex justify-between"><dt className="text-[#698097]">ID</dt><dd className="font-mono">{detail.id}</dd></div>
              <div className="flex justify-between"><dt className="text-[#698097]">Kode</dt><dd className="font-mono">{detail.code}</dd></div>
              <div className="flex justify-between"><dt className="text-[#698097]">Harga</dt><dd className="font-semibold tabular-nums">{rupiah(detail.price)}</dd></div>
              <div className="flex justify-between"><dt className="text-[#698097]">Kapasitas</dt><dd className="tabular-nums">{detail.capacity.toLocaleString("id-ID")}</dd></div>
              <div className="flex justify-between"><dt className="text-[#698097]">Status</dt><dd><StatusPill status={detail.status} /></dd></div>
              <div className="flex justify-between"><dt className="text-[#698097]">Dibuat</dt><dd>{new Date(detail.created_at).toLocaleDateString("id-ID")}</dd></div>
            </dl>
          </div>
        </div>
      )}
    </div>
  );
}

function TablePagination({ page, totalPages, onPage }: { page: number; totalPages: number; onPage: (p: number) => void }) {
  if (totalPages <= 1) return null;
  return (
    <div className="mt-3 flex items-center justify-end gap-1 text-xs text-[#698097]">
      <button onClick={() => onPage(page - 1)} disabled={page <= 1} className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40">‹</button>
      <span className="px-2">Hal {page} / {totalPages}</span>
      <button onClick={() => onPage(page + 1)} disabled={page >= totalPages} className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40">›</button>
    </div>
  );
}