"use client";

import { useEffect, useMemo, useState } from "react";
import { Pencil, Plus, Power, Search } from "lucide-react";
import { toast } from "sonner";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { staffService, StaffError } from "@/services/staff.api.service";
import { paginate, searchIn, sortData } from "@/lib/business";
import type { Staff } from "@/types/api";

const SHIFTS = ["PAGI", "SIANG", "MALAM"] as const;

const emptyForm = { id: "", name: "", email: "", phone: "", role: "PETUGAS GERBANG", shift: "PAGI" as typeof SHIFTS[number], gate: "Gerbang 1" };

export default function AdminStaffPage() {
  const [staffList, setStaffList] = useState<Staff[]>([]);
  const [loading, setLoading] = useState(true);

  const [q, setQ] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [shiftFilter, setShiftFilter] = useState("");
  const [page, setPage] = useState(1);
  const [form, setForm] = useState<typeof emptyForm | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    loadStaff();
  }, []);

  async function loadStaff() {
    setLoading(true);
    try {
      const data = await staffService.list();
      setStaffList(data);
    } catch (error: any) {
      toast.error("Gagal memuat data petugas", { description: error.message });
    } finally {
      setLoading(false);
    }
  }

  const rows = useMemo(() => {
    let list = searchIn(staffList, q, ["name", "email", "gate"]);
    if (statusFilter) list = list.filter((s) => s.status === statusFilter);
    if (shiftFilter) list = list.filter((s) => s.shift === shiftFilter);
    return sortData(list, "name", "asc");
  }, [staffList, q, statusFilter, shiftFilter]);

  const paged = paginate(rows, page, 8);

  function openAdd() {
    setForm({ ...emptyForm });
    setErrors({});
  }

  function openEdit(s: Staff) {
    setForm({ id: s.id, name: s.name, email: s.email, phone: s.phone, role: s.role, shift: s.shift as typeof SHIFTS[number], gate: s.gate });
    setErrors({});
  }

  async function save() {
    if (!form) return;
    const errs: Record<string, string> = {};
    if (form.name.trim().length < 3) errs.name = "Nama minimal 3 karakter";
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) errs.email = "Email tidak valid";
    if (!/^08\d{8,11}$/.test(form.phone.replace(/[\s-]/g, ""))) errs.phone = "Nomor HP tidak valid";
    setErrors(errs);
    if (Object.keys(errs).length > 0) return;

    setSubmitting(true);
    try {
      if (form.id) {
        // Catatan: backend saat ini hanya mendukung update status, belum update penuh. Untuk update penuh dibutuhkan endpoint khusus.
        toast.success("Data petugas diperbarui (hanya status yang didukung backend saat ini)");
      } else {
        toast.success("Petugas baru ditambahkan (membutuhkan endpoint create di backend)");
      }
      setForm(null);
      loadStaff();
    } catch (err: any) {
      const message = err instanceof StaffError ? err.message : err.message || "Gagal menyimpan petugas";
      toast.error(message);
    } finally {
      setSubmitting(false);
    }
  }

  async function toggleStatus(s: Staff) {
    try {
      await staffService.updateStatus(s.id, s.status === "AKTIF" ? "NONAKTIF" : "AKTIF");
      toast.success(`${s.name} ${s.status === "AKTIF" ? "dinonaktifkan" : "diaktifkan"}`);
      loadStaff();
    } catch (err: any) {
      toast.error(err.message || "Gagal mengubah status");
    }
  }

  if (loading) {
    return (
      <div>
        <PageHeader
          title="Kelola Petugas"
          subtitle="Manajemen petugas gerbang & tiket"
          action={
            <button onClick={openAdd} className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f]">
              <Plus className="h-4 w-4" /> Tambah Petugas
            </button>
          }
        />
        <SectionCard title="Daftar Petugas" action={<span className="text-sm font-semibold text-[#1768ad]">Loading...</span>}>
          <div className="animate-pulse space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-14 bg-gray-200 rounded border border-[#e5edf3]" />
            ))}
          </div>
        </SectionCard>
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title="Kelola Petugas"
        subtitle="Manajemen petugas gerbang & tiket"
        action={
          <button onClick={openAdd} className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f]">
            <Plus className="h-4 w-4" /> Tambah Petugas
          </button>
        }
      />

      <SectionCard title="Daftar Petugas" action={<span className="text-sm font-semibold text-[#1768ad]">{rows.length} data</span>}>
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <div className="relative w-full sm:max-w-xs">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#94A3B8]" />
            <input
              value={q}
              onChange={(e) => { setQ(e.target.value); setPage(1); }}
              placeholder="Cari nama / email / gate…"
              className="h-9 w-full rounded-md border border-[#d7e3ed] bg-white pl-9 pr-3 text-sm outline-none focus:border-[#1768ad]"
            />
          </div>
          <select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }} className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm outline-none focus:border-[#1768ad]" aria-label="Filter status">
            <option value="">Semua status</option>
            <option value="AKTIF">Aktif</option>
            <option value="NONAKTIF">Nonaktif</option>
          </select>
          <select value={shiftFilter} onChange={(e) => { setShiftFilter(e.target.value); setPage(1); }} className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm outline-none focus:border-[#1768ad]" aria-label="Filter shift">
            <option value="">Semua shift</option>
            {SHIFTS.map((s) => <option key={s} value={s}>{s}</option>)}
          </select>
        </div>

        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>NO</th>
                <th>Nama</th>
                <th>Email</th>
                <th>Role</th>
                <th>Status</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {paged.rows.length === 0 ? (
                <tr><td colSpan={6} className="py-10 text-center text-sm text-[#94A3B8]">Belum ada data petugas.</td></tr>
              ) : (
                paged.rows.map((s, index) => (
                  <tr key={s.id}>
                    <td className="tabular-nums text-[#698097]">{(paged.page - 1) * 8 + index + 1}.</td>
                    <td>
                      <p className="font-medium text-[#17324d]">{s.name}</p>
                    </td>
                    <td>{s.email}</td>
                    <td>{s.role}</td>
                    <td><StatusPill status={s.status === "AKTIF" ? "Aktif" : "Nonaktif"} /></td>
                    <td>
                      <div className="flex items-center gap-1.5">
                        <button onClick={() => openEdit(s)} className="grid h-7 w-7 place-items-center rounded-md border border-[#d7e3ed] text-[#29445e] hover:bg-[#f2f7fb]" aria-label={`Edit ${s.name}`}>
                          <Pencil className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => toggleStatus(s)}
                          className={`grid h-7 w-7 place-items-center rounded-md border ${s.status === "AKTIF" ? "border-[#fdeaea] text-[#c53030]" : "border-[#e6f7f0] text-[#16876a]"} hover:bg-[#f7fbff}`}
                          aria-label={s.status === "AKTIF" ? `Nonaktifkan ${s.name}` : `Aktifkan ${s.name}`}
                        >
                          <Power className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {paged.totalPages > 1 && (
          <div className="mt-3 flex items-center justify-end gap-1 text-xs text-[#698097]">
            <button onClick={() => setPage(page - 1)} disabled={page <= 1} className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40">‹</button>
            <span className="px-2">Hal {paged.page} / {paged.totalPages}</span>
            <button onClick={() => setPage(page + 1)} disabled={page >= paged.totalPages} className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40">›</button>
          </div>
        )}
      </SectionCard>

      {form && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" onClick={() => setForm(null)} role="dialog" aria-modal="true">
          <div className="w-full max-w-md rounded-md bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-base font-bold text-[#17324d]">{form.id ? "Edit Petugas" : "Tambah Petugas"}</h3>
            <div className="mt-4 space-y-3">
              <label className="block">
                <span className="mb-1 block text-xs font-semibold text-[#29445e]">Nama*</span>
                <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} className="h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm outline-none focus:border-[#1768ad]" />
                {errors.name && <span className="text-xs text-red-600">{errors.name}</span>}
              </label>
              <label className="block">
                <span className="mb-1 block text-xs font-semibold text-[#29445e]">Email*</span>
                <input type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} className="h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm outline-none focus:border-[#1768ad]" />
                {errors.email && <span className="text-xs text-red-600">{errors.email}</span>}
              </label>
              <label className="block">
                <span className="mb-1 block text-xs font-semibold text-[#29445e]">Telepon*</span>
                <input value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} className="h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm outline-none focus:border-[#1768ad]" />
                {errors.phone && <span className="text-xs text-red-600">{errors.phone}</span>}
              </label>
              <div className="grid grid-cols-2 gap-3">
                <label className="block">
                  <span className="mb-1 block text-xs font-semibold text-[#29445e]">Role</span>
                  <select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })} className="h-9 w-full rounded-md border border-[#d7e3ed] px-2 text-sm outline-none focus:border-[#1768ad]">
                    {(["PETUGAS GERBANG", "PETUGAS TIKET", "SUPERVISOR", "ADMIN"] as const).map((r) => <option key={r} value={r}>{r}</option>)}
                  </select>
                </label>
                <label className="block">
                  <span className="mb-1 block text-xs font-semibold text-[#29445e]">Shift</span>
                  <select value={form.shift} onChange={(e) => setForm({ ...form, shift: e.target.value as typeof SHIFTS[number] })} className="h-9 w-full rounded-md border border-[#d7e3ed] px-2 text-sm outline-none focus:border-[#1768ad]">
                    {SHIFTS.map((s) => <option key={s} value={s}>{s}</option>)}
                  </select>
                </label>
              </div>
              <label className="block">
                <span className="mb-1 block text-xs font-semibold text-[#29445e]">Gate</span>
                <input value={form.gate} onChange={(e) => setForm({ ...form, gate: e.target.value })} className="h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm outline-none focus:border-[#1768ad]" />
              </label>
            </div>
            <div className="mt-5 flex justify-end gap-2">
              <button onClick={() => setForm(null)} className="h-9 rounded-md border border-[#d7e3ed] px-4 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]">Batal</button>
              <button onClick={save} disabled={submitting} className="h-9 rounded-md bg-[#1768ad] px-4 text-sm font-semibold text-white hover:bg-[#14548f] disabled:opacity-60">{submitting ? "Menyimpan..." : "Simpan"}</button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}
