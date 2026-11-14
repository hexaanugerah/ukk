"use client";

import { useEffect, useMemo, useState } from "react";
import { Search, ShieldCheck } from "lucide-react";
import { toast } from "sonner";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { roleService, RoleError, type AdminUserItem, type RoleItem } from "@/services/role.api.service";
import { searchIn, sortData } from "@/lib/business";

// Kelola Role — hanya bisa diakses super_admin (backend /super-admin/*).
// Super admin memilih pengguna, lalu mengganti seluruh role-nya (replace).

export default function AdminRolesPage() {
  const [roles, setRoles] = useState<RoleItem[]>([]);
  const [users, setUsers] = useState<AdminUserItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [forbidden, setForbidden] = useState(false);

  const [q, setQ] = useState("");
  const [saving, setSaving] = useState(false);
  // { userId, selected } — form penetapan role untuk satu pengguna.
  const [editing, setEditing] = useState<{ user: AdminUserItem; selected: string[] } | null>(null);

  useEffect(() => {
    load();
  }, []);

  async function load() {
    setLoading(true);
    setForbidden(false);
    try {
      const [roleRows, userRows] = await Promise.all([roleService.listRoles(), roleService.listUsers()]);
      setRoles(roleRows);
      setUsers(userRows);
    } catch (err: any) {
      if (err?.status === 403) {
        setForbidden(true);
      } else {
        toast.error("Gagal memuat data role", { description: err?.message });
      }
    } finally {
      setLoading(false);
    }
  }

  const rows = useMemo(() => sortData(searchIn(users, q, ["name", "email"]), "name", "asc"), [users, q]);

  function openEdit(u: AdminUserItem) {
    setEditing({ user: u, selected: [...u.roles] });
  }

  function toggleRole(name: string) {
    setEditing((prev) => {
      if (!prev) return prev;
      const selected = prev.selected.includes(name)
        ? prev.selected.filter((r) => r !== name)
        : [...prev.selected, name];
      return { ...prev, selected };
    });
  }

  async function save() {
    if (!editing) return;
    if (editing.selected.length === 0) {
      toast.error("Pilih minimal satu role", { description: "Pengguna tanpa role tidak bisa login." });
      return;
    }
    setSaving(true);
    try {
      await roleService.assignRoles(editing.user.id, editing.selected);
      toast.success(`Role "${editing.user.name}" diperbarui`);
      setEditing(null);
      const userRows = await roleService.listUsers();
      setUsers(userRows);
    } catch (err: any) {
      toast.error(err instanceof RoleError ? err.message : err?.message || "Gagal menyimpan role");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div>
        <PageHeader title="Kelola Role" subtitle="Penetapan role pengguna (super admin)" />
        <SectionCard title="Daftar Pengguna" action={<span className="text-sm font-semibold text-[#1768ad]">Loading...</span>}>
          <div className="animate-pulse space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-14 bg-gray-200 rounded border border-[#e5edf3]" />
            ))}
          </div>
        </SectionCard>
      </div>
    );
  }

  if (forbidden) {
    return (
      <div>
        <PageHeader title="Kelola Role" subtitle="Penetapan role pengguna (super admin)" />
        <SectionCard title="Akses Ditolak">
          <p className="text-sm text-[#698097]">
            Hanya <strong>super_admin</strong> yang bisa mengelola penetapan role. Masuk dengan akun super admin
            (superadmin@kawahputih.id) untuk menggunakan halaman ini.
          </p>
        </SectionCard>
      </div>
    );
  }

  return (
    <div>
      <PageHeader title="Kelola Role" subtitle="Penetapan role pengguna (super admin)" />

      <SectionCard title="Daftar Role" action={<span className="text-sm font-semibold text-[#1768ad]">{roles.length} role</span>}>
        <div className="grid gap-2 sm:grid-cols-2">
          {roles.map((r) => (
            <div key={r.id} className="flex items-start gap-2 rounded-md border border-[#e5edf3] bg-[#f7fbff] px-3 py-2">
              <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0 text-[#1768ad]" />
              <div>
                <p className="text-sm font-semibold text-[#17324d]">{r.name}</p>
                <p className="text-xs text-[#698097]">{r.description || "-"}</p>
              </div>
            </div>
          ))}
        </div>
      </SectionCard>

      <SectionCard title="Penetapan Role Pengguna" action={<span className="text-sm font-semibold text-[#1768ad]">{rows.length} pengguna</span>}>
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <div className="relative w-full sm:max-w-xs">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#94A3B8]" />
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="Cari nama / email…"
              className="h-9 w-full rounded-md border border-[#d7e3ed] bg-white pl-9 pr-3 text-sm outline-none focus:border-[#1768ad]"
            />
          </div>
        </div>

        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>NO</th>
                <th>Nama</th>
                <th>Email</th>
                <th>Role Saat Ini</th>
                <th>Status</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {rows.length === 0 ? (
                <tr><td colSpan={6} className="py-10 text-center text-sm text-[#94A3B8]">Belum ada pengguna.</td></tr>
              ) : (
                rows.map((u, index) => (
                  <tr key={u.id}>
                    <td className="tabular-nums text-[#698097]">{index + 1}.</td>
                    <td><p className="font-medium text-[#17324d]">{u.name}</p></td>
                    <td>{u.email}</td>
                    <td>
                      <div className="flex flex-wrap gap-1">
                        {u.roles.length === 0 ? (
                          <span className="text-xs text-red-600">tanpa role</span>
                        ) : (
                          u.roles.map((r) => (
                            <span key={r} className="rounded-full bg-[#eef4f9] px-2 py-0.5 text-[11px] font-semibold text-[#29445e]">{r}</span>
                          ))
                        )}
                      </div>
                    </td>
                    <td><StatusPill status={u.is_active ? "Aktif" : "Nonaktif"} /></td>
                    <td>
                      <button
                        onClick={() => openEdit(u)}
                        className="rounded-md border border-[#d7e3ed] px-2 py-1 text-xs font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
                      >
                        Atur Role
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </SectionCard>

      {editing && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" onClick={() => setEditing(null)} role="dialog" aria-modal="true">
          <div className="w-full max-w-sm rounded-md bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-base font-bold text-[#17324d]">Atur Role — {editing.user.name}</h3>
            <p className="mt-1 text-xs text-[#698097]">{editing.user.email}</p>
            <div className="mt-4 space-y-2">
              {roles.map((r) => (
                <label key={r.id} className="flex items-center gap-2 rounded-md border border-[#e5edf3] px-3 py-2 text-sm">
                  <input
                    type="checkbox"
                    checked={editing.selected.includes(r.name)}
                    onChange={() => toggleRole(r.name)}
                    className="h-4 w-4 accent-[#1768ad]"
                  />
                  <span>
                    <span className="font-semibold text-[#17324d]">{r.name}</span>
                    {r.description && <span className="block text-xs text-[#698097]">{r.description}</span>}
                  </span>
                </label>
              ))}
            </div>
            <p className="mt-2 text-[11px] text-[#94A3B8]">
              Menyimpan menggantikan seluruh role pengguna ini dengan pilihan di atas.
            </p>
            <div className="mt-5 flex justify-end gap-2">
              <button onClick={() => setEditing(null)} className="h-9 rounded-md border border-[#d7e3ed] px-4 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]">Batal</button>
              <button onClick={save} disabled={saving} className="h-9 rounded-md bg-[#1768ad] px-4 text-sm font-semibold text-white hover:bg-[#14548f] disabled:opacity-60">
                {saving ? "Menyimpan..." : "Simpan"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
