import type { Metadata } from "next";
import { ResetPasswordForm } from "@/features/authentication/components/reset-password-form";

export const metadata: Metadata = { title: "Reset Password" };

export default function ResetPasswordPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-brand-50 px-4 dark:bg-slate-950">
      <div className="w-full max-w-sm rounded-2xl border border-slate-200 bg-white p-8 shadow-sm dark:border-slate-800 dark:bg-slate-900">
        <h1 className="mb-1 text-xl font-semibold">Atur Ulang Password</h1>
        <p className="mb-6 text-sm text-slate-500">Masukkan password baru untuk akun kamu.</p>
        <ResetPasswordForm />
      </div>
    </main>
  );
}
