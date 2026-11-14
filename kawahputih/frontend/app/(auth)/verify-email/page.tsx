import type { Metadata } from "next";
import { VerifyEmailForm } from "@/features/authentication/components/verify-email-form";

export const metadata: Metadata = { title: "Verifikasi Email" };

export default function VerifyEmailPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-brand-50 px-4 dark:bg-slate-950">
      <div className="w-full max-w-sm rounded-2xl border border-slate-200 bg-white p-8 shadow-sm dark:border-slate-800 dark:bg-slate-900">
        <h1 className="mb-1 text-xl font-semibold">Verifikasi Email</h1>
        <p className="mb-6 text-sm text-slate-500">Konfirmasi alamat email akun kamu.</p>
        <VerifyEmailForm />
      </div>
    </main>
  );
}
