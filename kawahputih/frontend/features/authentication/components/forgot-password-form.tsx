"use client";

import { useState } from "react";
import Link from "next/link";
import { CheckCircle2 } from "lucide-react";
import { forgotPasswordSchema, sanitizeEmail } from "../schemas/auth.schema";

export function ForgotPasswordForm() {
  const [sent, setSent] = useState(false);
  const [email, setEmail] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = sanitizeEmail(e.target.value);
    setEmail(value);
    
    if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
      setError("Format email tidak valid");
    } else {
      setError(null);
    }
  };

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const parsed = forgotPasswordSchema.safeParse({ email: sanitizeEmail(email) });
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? "Email tidak valid");
      return;
    }
    setError(null);
    setSending(true);
    setTimeout(() => {
      setSending(false);
      setSent(true);
    }, 600);
  }

  if (sent) {
    return (
      <div className="flex flex-col items-center gap-3 py-6 text-center">
        <CheckCircle2 className="h-10 w-10 text-green-600" />
        <p className="text-sm text-slate-600 dark:text-slate-400">
          Kalau email kamu terdaftar, link reset password sudah dikirim. Cek kotak masuk kamu.
        </p>
        <Link href="/login" className="text-sm font-medium text-brand-600 hover:underline">
          Kembali ke halaman Masuk
        </Link>
      </div>
    );
  }

  return (
    <form onSubmit={onSubmit} className="space-y-4" noValidate>
      <div>
        <label className="mb-1 block text-sm font-medium">Email</label>
        <input
          type="email"
          value={email}
          onChange={handleEmailChange}
          className={`h-10 w-full rounded-md border px-3 text-sm outline-none focus:border-brand-600 ${
            error ? "border-red-500" : "border-slate-300"
          }`}
          aria-invalid={!!error}
          aria-describedby={error ? "email-error" : undefined}
        />
        {error && <p id="email-error" className="mt-1 text-xs text-red-600" role="alert">{error}</p>}
      </div>
      <button
        type="submit"
        disabled={sending}
        className="h-10 w-full rounded-md bg-brand text-sm font-semibold text-white hover:bg-brand-dark disabled:opacity-60"
      >
        {sending ? "Mengirim..." : "Kirim Link Reset"}
      </button>
    </form>
  );
}