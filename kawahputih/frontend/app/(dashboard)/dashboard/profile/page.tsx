"use client";

import { useState } from "react";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth.store";
import { authService } from "@/services/auth.api.service";
import { useLogout } from "@/hooks/use-auth";
import { PhoneInput } from "@/components/ui/phone-input";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/panel/panel-ui";

// Profil — lihat & ubah info profil.

export default function ProfilePage() {
  const user = useAuthStore((s) => s.user);
  const logout = useLogout();
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(user?.name ?? "");
  const [phone, setPhone] = useState(user?.phone ?? "");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!user) return null;

  async function save() {
    if (name.trim().length < 3) {
      setError("Nama minimal 3 karakter");
      return;
    }
    setError(null);
    setSaving(true);
    try {
      await authService.updateProfile({ name: name.trim(), phone });
      toast.success("Profil diperbarui");
      setEditing(false);
    } catch {
      toast.error("Gagal memperbarui profil");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div>
      <PageHeader title="Profil" subtitle="Kelola informasi akun kamu" />
      <Card className="max-w-md p-6">
        <dl className="space-y-3 text-sm">
          <div>
            <dt className="text-slate-500">Nama</dt>
            <dd className="font-medium">
              {editing ? (
                <input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="mt-1 h-9 w-full rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-brand-600"
                  aria-label="Nama"
                />
              ) : (
                user.name
              )}
            </dd>
          </div>
          <div>
            <dt className="text-slate-500">Email</dt>
            <dd className="font-medium">{user.email}</dd>
          </div>
          <div>
            <dt className="text-slate-500">Telepon</dt>
            <dd className="font-medium">
              {editing ? (
                <PhoneInput
                  value={phone}
                  onChange={setPhone}
                  inputClassName="mt-1 h-9 w-full rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-brand-600"
                  selectClassName="mt-1 h-9 rounded-md border border-slate-300 bg-slate-50 px-2 text-sm text-slate-600 outline-none focus:border-brand-600"
                  ariaLabel="Telepon"
                />
              ) : (
                user.phone || "-"
              )}
            </dd>
          </div>
          <div>
            <dt className="text-slate-500">Role</dt>
            <dd className="font-medium">{user.role}</dd>
          </div>
          {user.gate && (
            <div>
              <dt className="text-slate-500">Gate / Shift</dt>
              <dd className="font-medium">
                {user.gate} • {user.shift}
              </dd>
            </div>
          )}
        </dl>

        {error && <p className="mt-3 text-xs text-red-600">{error}</p>}

        <div className="mt-5 flex gap-2">
          {editing ? (
            <>
              <button
                onClick={save}
                disabled={saving}
                className="h-9 rounded-md bg-brand px-4 text-sm font-semibold text-white hover:bg-brand-dark disabled:opacity-60"
              >
                {saving ? "Menyimpan..." : "Simpan"}
              </button>
              <button
                onClick={() => {
                  setEditing(false);
                  setName(user.name);
                  setPhone(user.phone);
                }}
                className="h-9 rounded-md border border-slate-300 px-4 text-sm font-semibold hover:bg-slate-50"
              >
                Batal
              </button>
            </>
          ) : (
            <button
              onClick={() => setEditing(true)}
              className="h-9 rounded-md bg-brand px-4 text-sm font-semibold text-white hover:bg-brand-dark"
            >
              Edit Profil
            </button>
          )}
        </div>
      </Card>

      <button
        onClick={logout}
        className="mt-6 inline-flex h-10 items-center gap-2 rounded-md border border-red-200 px-4 text-sm font-semibold text-red-600 hover:bg-red-50"
      >
        Logout
      </button>
    </div>
  );
}
