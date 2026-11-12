"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { resetPasswordSchema, type ResetPasswordFormValues } from "../schemas/auth.schema";

// Mode mock: reset password hanya simulasi — tidak ada backend email.
export function ResetPasswordForm() {
  const router = useRouter();
  const [token, setToken] = useState<string | null>(null);
  const [tokenChecked, setTokenChecked] = useState(false);

  const [values, setValues] = useState<ResetPasswordFormValues>({
    token: "",
    password: "",
    confirmPassword: "",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const t = new URLSearchParams(window.location.search).get("token");
    setToken(t);
    if (t) setValues((v) => ({ ...v, token: t }));
    setTokenChecked(true);
  }, []);

  if (!tokenChecked) {
    return <p className="text-sm text-slate-500">Memeriksa link reset...</p>;
  }

  if (!token) {
    return <p className="text-sm text-red-600">Link reset password tidak valid atau sudah kedaluwarsa.</p>;
  }

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setValues((v) => ({ ...v, password: value }));
    
    if (value && value.length < 8) {
      setErrors((prev) => ({ ...prev, password: "Password minimal 8 karakter" }));
    } else if (value && value.length > 72) {
      setErrors((prev) => ({ ...prev, password: "Password maksimal 72 karakter" }));
    } else if (value && !/[A-Z]/.test(value)) {
      setErrors((prev) => ({ ...prev, password: "Password harus mengandung huruf besar" }));
    } else if (value && !/[a-z]/.test(value)) {
      setErrors((prev) => ({ ...prev, password: "Password harus mengandung huruf kecil" }));
    } else if (value && !/[0-9]/.test(value)) {
      setErrors((prev) => ({ ...prev, password: "Password harus mengandung angka" }));
    } else {
      setErrors((prev) => {
        const next = { ...prev };
        delete next.password;
        return next;
      });
    }
  };

  const handleConfirmPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setValues((v) => ({ ...v, confirmPassword: value }));
    
    if (value && value !== values.password) {
      setErrors((prev) => ({ ...prev, confirmPassword: "Konfirmasi password tidak cocok" }));
    } else {
      setErrors((prev) => {
        const next = { ...prev };
        delete next.confirmPassword;
        return next;
      });
    }
  };

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const parsed = resetPasswordSchema.safeParse(values);
    if (!parsed.success) {
      const errs: Record<string, string> = {};
      for (const issue of parsed.error.issues) {
        const key = issue.path[0] as string;
        if (key && !errs[key]) errs[key] = issue.message;
      }
      setErrors(errs);
      return;
    }
    setErrors({});
    setSubmitting(true);
    setTimeout(() => {
      setSubmitting(false);
      toast.success("Password berhasil direset (simulasi), silakan masuk kembali");
      router.push("/login");
    }, 600);
  }

  return (
    <form onSubmit={onSubmit} className="space-y-4" noValidate>
      <div>
        <label className="mb-1 block text-sm font-medium">Password Baru</label>
        <input
          type="password"
          value={values.password}
          onChange={handlePasswordChange}
          className={`h-10 w-full rounded-md border px-3 text-sm outline-none focus:border-brand-600 ${
            errors.password ? "border-red-500" : "border-slate-300"
          }`}
          aria-invalid={!!errors.password}
          aria-describedby={errors.password ? "password-error" : undefined}
        />
        {errors.password && <p id="password-error" className="mt-1 text-xs text-red-600" role="alert">{errors.password}</p>}
      </div>
      <div>
        <label className="mb-1 block text-sm font-medium">Konfirmasi Password</label>
        <input
          type="password"
          value={values.confirmPassword}
          onChange={handleConfirmPasswordChange}
          className={`h-10 w-full rounded-md border px-3 text-sm outline-none focus:border-brand-600 ${
            errors.confirmPassword ? "border-red-500" : "border-slate-300"
          }`}
          aria-invalid={!!errors.confirmPassword}
          aria-describedby={errors.confirmPassword ? "confirm-error" : undefined}
        />
        {errors.confirmPassword && <p id="confirm-error" className="mt-1 text-xs text-red-600" role="alert">{errors.confirmPassword}</p>}
      </div>
      <button
        type="submit"
        disabled={submitting}
        className="h-10 w-full rounded-md bg-brand text-sm font-semibold text-white hover:bg-brand-dark disabled:opacity-60"
      >
        {submitting ? "Memproses..." : "Reset Password"}
      </button>
    </form>
  );
}