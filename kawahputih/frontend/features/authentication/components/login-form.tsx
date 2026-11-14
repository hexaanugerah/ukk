"use client";

import { useState, useEffect } from "react";
import { Eye, EyeOff, Loader2, Facebook, Chrome } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useLogin } from "@/hooks/use-auth";
import { toast } from "sonner";
import { authService } from "@/services/auth.api.service";
import { loginSchema, sanitizeEmail } from "@/features/authentication/schemas/auth.schema";

// Login form — validasi email/password, show/hide password, remember me.

export function LoginForm() {
  const login = useLogin();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [remember, setRemember] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  // Aktif saat backend menolak login karena email belum diverifikasi (AUTH-06).
  // Menampilkan tombol kirim-ulang email verifikasi di bawah form.
  const [needsVerification, setNeedsVerification] = useState(false);
  const [resending, setResending] = useState(false);

  // CKEditor

  // Validasi email real-time
  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setEmail(value);
    
    if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
      setErrors((prev) => ({ ...prev, email: "Format email tidak valid" }));
    } else {
      setErrors((prev) => {
        const next = { ...prev };
        delete next.email;
        return next;
      });
    }
  };

  // Validasi password real-time
  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setPassword(value);
    
    if (value && value.length < 8) {
      setErrors((prev) => ({ ...prev, password: "Password minimal 8 karakter" }));
    } else {
      setErrors((prev) => {
        const next = { ...prev };
        delete next.password;
        return next;
      });
    }
  };

  // Isi ulang form dari sesi yang tersimpan kalau ada, supaya tidak perlu
  // mendaftar/masuk ulang tiap buka halaman.
  useEffect(() => {
    if (typeof window === "undefined") return;
    const saved = localStorage.getItem("kpr-auth");
    if (!saved) return;
    try {
      const parsed = JSON.parse(saved);
      const state = parsed?.state;
      if (state?.user) {
        setEmail(state.user.email ?? "");
      }
    } catch {
      // yauguh
    }
  }, []);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErrors({});

    // Validasi dengan skema Zod
    const formData = { email: sanitizeEmail(email), password };
    const result = loginSchema.safeParse(formData);
    
    if (!result.success) {
      const errs: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const path = issue.path.join(".");
        errs[path] = issue.message;
      }
      setErrors(errs);
      return;
    }

    setSubmitting(true);
    // Bawa tujuan asal (mis. /booking?... dari halaman tiket) supaya setelah
    // login pengunjung kembali ke halaman itu, bukan ke dasbor pengunjung.
    const redirectTo = new URLSearchParams(window.location.search).get("redirect") ?? undefined;
    const res = await login({ email: sanitizeEmail(email), password, remember, redirectTo });
    setSubmitting(false);

    if (res.ok) {
      setNeedsVerification(false);
      toast.success("Login berhasil", { description: "Selamat datang kembali!" });
    } else {
      setErrors({ form: res.error });
      setNeedsVerification(res.error.includes("belum diverifikasi"));
      toast.error("Login gagal", { description: res.error });
    }
  }

  async function resendVerification() {
    const value = sanitizeEmail(email);
    if (!value || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
      toast.error("Masukkan email terlebih dahulu untuk kirim ulang verifikasi");
      return;
    }
    setResending(true);
    try {
      await authService.requestEmailVerification(value);
      toast.success("Email verifikasi dikirim", {
        description: "Cek kotak masuk Anda (juga folder spam) untuk link verifikasi.",
      });
    } catch {
      toast.error("Gagal mengirim email verifikasi");
    } finally {
      setResending(false);
    }
  }

  return (
    <div className="space-y-5">
      <form onSubmit={onSubmit} className="space-y-5" noValidate>
        <div>
          <label htmlFor="login-email" className="sr-only">
            Email
          </label>
          <Input
            id="login-email"
            type="email"
            placeholder="Email"
            autoComplete="email"
            value={email}
            onChange={handleEmailChange}
            className={errors.email ? "h-11 border-red-500" : "h-11"}
            aria-invalid={!!errors.email}
            aria-describedby={errors.email ? "email-error" : undefined}
          />
          {errors.email && (
            <p id="email-error" role="alert" className="mt-1 text-xs text-red-600">
              {errors.email}
            </p>
          )}
        </div>

        <div>
          <label htmlFor="login-password" className="sr-only">
            Password
          </label>
          <div className="relative">
            <Input
              id="login-password"
              type={showPassword ? "text" : "password"}
              placeholder="Password"
              autoComplete="current-password"
              value={password}
              onChange={handlePasswordChange}
              className="h-11 pr-10"
              aria-invalid={!!errors.password}
              aria-describedby={errors.password ? "password-error" : undefined}
            />
            <button
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 p-1 text-[#698097] hover:text-[#29445e]"
              aria-label={showPassword ? "Sembunyikan password" : "Tampilkan password"}
            >
              {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </button>
          </div>
          {errors.password && (
            <p id="password-error" role="alert" className="mt-1 text-xs text-red-600">
              {errors.password}
            </p>
          )}
          <div className="mt-2 flex items-center justify-between">
            <label className="flex items-center gap-2 text-[13px] text-[#698097]">
              <input
                type="checkbox"
                checked={remember}
                onChange={(e) => setRemember(e.target.checked)}
                className="h-4 w-4 accent-[#1768ad]"
              />
              Ingat saya
            </label>
            <Link href="/forgot-password" className="text-[13px] font-semibold text-[#1768ad] hover:underline">
              Lupa password?
            </Link>
          </div>
        </div>

        {errors.form && (
          <p role="alert" className="rounded-md bg-[#fdeaea] px-3 py-2 text-sm text-[#c53030]">
            {errors.form}
          </p>
        )}

        {needsVerification && (
          <div className="rounded-md bg-[#fff8e6] px-3 py-2 text-sm text-[#8a6d1a]">
            <p>Email Anda belum diverifikasi. Buka link verifikasi di kotak masuk, atau minta kirim ulang di bawah.</p>
            <button
              type="button"
              onClick={resendVerification}
              disabled={resending}
              className="mt-2 rounded-md border border-[#e5d18c] px-3 py-1.5 text-xs font-semibold text-[#8a6d1a] hover:bg-[#fdf3d1] disabled:opacity-60"
            >
              {resending ? "Mengirim..." : "Kirim ulang email verifikasi"}
            </button>
          </div>
        )}

        <Button
          type="submit"
          size="lg"
          className="h-12 w-full bg-[#1768ad] text-base hover:bg-[#14548f]"
          disabled={submitting}
        >
          {submitting ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" /> Memproses...
            </>
          ) : (
            "Masuk"
          )}
        </Button>
      </form>

      <p className="text-center text-sm text-[#698097]">
        Belum punya akun?{" "}
        <Link href="/register" className="font-semibold text-[#1768ad] hover:underline">
          Register
        </Link>
      </p>

      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <span className="w-full border-t border-[#d4dce5]" />
        </div>
        <div className="relative flex justify-center text-xs uppercase font-medium text-[#698097]">
          <span className="bg-[#f7fafc] px-2">atau login dengan</span>
        </div>
      </div>

      <div className="grid gap-3">
        <Button
          type="button"
          size="lg"
          className="h-12 w-full bg-white border border-[#d4dce5] text-[#1768ad] hover:bg-[#eef4f9]"
          onClick={() => window.open("https://accounts.google.com", "_self")}
        >
          <Chrome className="mr-2 h-5 w-5" />
          Masuk dengan Google
        </Button>
        <Button
          type="button"
          size="lg"
          className="h-12 w-full bg-white border border-[#d4dce5] text-[#1768ad] hover:bg-[#eef4f9]"
          onClick={() => window.open("https://www.facebook.com", "_self")}
        >
          <Facebook className="mr-2 h-5 w-5" />
          Masuk dengan Facebook
        </Button>
      </div>

      {/* Google & Facebook tersedia bila backend login sosial sudah diaktifkan */}
    </div>
  );
}