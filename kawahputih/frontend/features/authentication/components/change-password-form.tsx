"use client";

import { useState } from "react";
import { Loader2, ShieldCheck } from "lucide-react";
import { toast } from "sonner";
import { changePasswordSchema } from "../schemas/auth.schema";
import { useChangePassword } from "@/hooks/use-auth";

// Form ganti password untuk SEMUA role (pengunjung, petugas, manager, admin).
// Backend memverifikasi password lama dan menyimpan hash bcrypt baru; form hanya
// mengirim teks password lewat HTTPS/API dan tidak pernah menyimpannya.

type Field = "currentPassword" | "newPassword" | "confirmPassword";
const EMPTY: Record<Field, string> = { currentPassword: "", newPassword: "", confirmPassword: "" };

const LABELS: Record<Field, string> = {
  currentPassword: "Password Saat Ini",
  newPassword: "Password Baru",
  confirmPassword: "Konfirmasi Password Baru",
};

export function ChangePasswordForm({ className }: { className?: string }) {
  const changePassword = useChangePassword();
  const [values, setValues] = useState<Record<Field, string>>(EMPTY);
  const [errors, setErrors] = useState<Partial<Record<Field, string>>>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  function setField(field: Field, value: string) {
    setValues((prev) => ({ ...prev, [field]: value }));
    setErrors((prev) => {
      if (!prev[field]) return prev;
      const next = { ...prev };
      delete next[field];
      return next;
    });
    setFormError(null);
  }

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault();
    setFormError(null);

    const parsed = changePasswordSchema.safeParse(values);
    if (!parsed.success) {
      const next: Partial<Record<Field, string>> = {};
      for (const issue of parsed.error.issues) {
        const key = issue.path[0];
        if ((key === "currentPassword" || key === "newPassword" || key === "confirmPassword") && !next[key]) {
          next[key] = issue.message;
        }
      }
      setErrors(next);
      return;
    }

    setSubmitting(true);
    const result = await changePassword(values.currentPassword, values.newPassword);
    setSubmitting(false);

    if (!result.ok) {
      setFormError(result.error);
      return;
    }

    // Sesi sudah dicabut backend + store dibersihkan hook; pengguna diarahkan ke /login.
    setValues(EMPTY);
    toast.success("Password berhasil diubah. Silakan masuk kembali.");
  }

  return (
    <form onSubmit={onSubmit} className={className} noValidate>
      <div className="space-y-3">
        {(Object.keys(EMPTY) as Field[]).map((field) => {
          const error = errors[field];
          return (
            <div key={field}>
              <label htmlFor={`change-${field}`} className="mb-1 block text-sm font-medium">
                {LABELS[field]}
              </label>
              <input
                id={`change-${field}`}
                type="password"
                value={values[field]}
                onChange={(event) => setField(field, event.target.value)}
                autoComplete={field === "currentPassword" ? "current-password" : "new-password"}
                className={`h-10 w-full rounded-md border px-3 text-sm outline-none focus:border-brand-600 ${
                  error ? "border-red-500" : "border-slate-300"
                }`}
                aria-invalid={!!error}
                aria-describedby={error ? `change-${field}-error` : undefined}
              />
              {error && (
                <p id={`change-${field}-error`} className="mt-1 text-xs text-red-600" role="alert">
                  {error}
                </p>
              )}
            </div>
          );
        })}
      </div>

      <p className="mt-3 flex items-start gap-1.5 text-[11px] text-slate-500">
        <ShieldCheck className="mt-px h-3.5 w-3.5 shrink-0 text-brand-600" aria-hidden="true" />
        Minimal 8 karakter dengan kombinasi huruf besar, huruf kecil, dan angka.
      </p>

      {formError && (
        <p className="mt-3 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700" role="alert">
          {formError}
        </p>
      )}

      <button
        type="submit"
        disabled={submitting}
        className="mt-4 inline-flex h-10 w-full items-center justify-center gap-1.5 rounded-md bg-brand text-sm font-semibold text-white hover:bg-brand-dark disabled:opacity-60"
      >
        {submitting && <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />}
        {submitting ? "Menyimpan..." : "Simpan Password Baru"}
      </button>
    </form>
  );
}
