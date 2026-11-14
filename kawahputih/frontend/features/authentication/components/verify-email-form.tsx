"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { CheckCircle2, XCircle } from "lucide-react";
import { toast } from "sonner";
import { authService, AuthError } from "@/services/auth.api.service";
import { sanitizeEmail } from "@/features/authentication/schemas/auth.schema";

// Verifikasi email — membuka halaman ini dari link email verifikasi backend
// (?token=...). Backend memvalidasi token, menandai akun IsVerified=true, dan
// token dianggap terpakai (tidak bisa dipakai ulang).
export function VerifyEmailForm() {
  const [token, setToken] = useState<string | null>(null);
  const [tokenChecked, setTokenChecked] = useState(false);
  const [state, setState] = useState<"idle" | "verifying" | "success" | "error">("idle");
  const [message, setMessage] = useState<string | null>(null);
  const [resendEmail, setResendEmail] = useState("");
  const [resending, setResending] = useState(false);

  useEffect(() => {
    const t = new URLSearchParams(window.location.search).get("token");
    setToken(t);
    setTokenChecked(true);
  }, []);

  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    setState("verifying");
    (async () => {
      try {
        await authService.verifyEmail(token);
        if (!cancelled) setState("success");
      } catch (err) {
        if (cancelled) return;
        setMessage(err instanceof AuthError ? err.message : "Verifikasi email gagal.");
        setState("error");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [token]);

  async function resend() {
    const email = sanitizeEmail(resendEmail);
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      toast.error("Masukkan email yang valid");
      return;
    }
    setResending(true);
    try {
      await authService.requestEmailVerification(email);
      toast.success("Email verifikasi dikirim", { description: "Cek kotak masuk Anda (juga folder spam)." });
    } catch (err) {
      toast.error(err instanceof AuthError ? err.message : "Gagal mengirim email verifikasi.");
    } finally {
      setResending(false);
    }
  }

  if (!tokenChecked) {
    return <p className="text-sm text-slate-500">Memeriksa link verifikasi...</p>;
  }

  if (!token) {
    return (
      <div className="space-y-3">
        <p className="text-sm text-red-600">Link verifikasi tidak ditemukan. Buka link dari email Anda.</p>
        <div className="rounded-md bg-slate-50 p-3">
          <p className="text-xs text-slate-500">Belum menerima email? Masukkan email akun Anda untuk kirim ulang:</p>
          <input
            type="email"
            value={resendEmail}
            onChange={(e) => setResendEmail(e.target.value)}
            placeholder="email@contoh.com"
            className="mt-2 h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-brand-600"
          />
          <button
            type="button"
            onClick={resend}
            disabled={resending}
            className="mt-2 h-9 w-full rounded-md bg-brand text-xs font-semibold text-white hover:bg-brand-dark disabled:opacity-60"
          >
            {resending ? "Mengirim..." : "Kirim ulang email verifikasi"}
          </button>
        </div>
        <Link href="/login" className="block text-center text-sm font-medium text-brand-600 hover:underline">
          Kembali ke halaman Masuk
        </Link>
      </div>
    );
  }

  if (state === "verifying") {
    return <p className="text-sm text-slate-500">Memverifikasi email...</p>;
  }

  if (state === "success") {
    return (
      <div className="flex flex-col items-center gap-3 py-4 text-center">
        <CheckCircle2 className="h-10 w-10 text-green-600" />
        <p className="text-sm text-slate-600">Email berhasil diverifikasi. Sekarang Anda bisa masuk ke akun.</p>
        <Link href="/login" className="text-sm font-medium text-brand-600 hover:underline">
          Masuk sekarang
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center gap-3 py-4 text-center">
      <XCircle className="h-10 w-10 text-red-600" />
      <p className="text-sm text-slate-600">{message ?? "Link verifikasi tidak valid atau sudah kedaluwarsa."}</p>
      <div className="w-full rounded-md bg-slate-50 p-3 text-left">
        <p className="text-xs text-slate-500">Minta link baru (berlaku 24 jam):</p>
        <input
          type="email"
          value={resendEmail}
          onChange={(e) => setResendEmail(e.target.value)}
          placeholder="email akun Anda"
          className="mt-2 h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-brand-600"
        />
        <button
          type="button"
          onClick={resend}
          disabled={resending}
          className="mt-2 h-9 w-full rounded-md bg-brand text-xs font-semibold text-white hover:bg-brand-dark disabled:opacity-60"
        >
          {resending ? "Mengirim..." : "Kirim ulang email verifikasi"}
        </button>
      </div>
      <Link href="/login" className="text-sm font-medium text-brand-600 hover:underline">
        Kembali ke halaman Masuk
      </Link>
    </div>
  );
}
