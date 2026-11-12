import type { Metadata } from "next";
import { RegisterForm } from "@/features/authentication/components/register-form";

export const metadata: Metadata = { title: "Daftar" };

export default function RegisterPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-brand-50 px-4 py-10 dark:bg-slate-950">
      <div className="w-full max-w-sm rounded-2xl border border-slate-200 bg-white p-8 shadow-sm dark:border-slate-800 dark:bg-slate-900">
        <h1 className="mb-1 text-xl font-semibold">Buat Akun Baru</h1>
        <p className="mb-6 text-sm text-slate-500">Kawah Putih Rancabali Tourism</p>
        <RegisterForm />
      </div>
    </main>
  );
}
