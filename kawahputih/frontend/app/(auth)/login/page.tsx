import type { Metadata } from "next";
import Link from "next/link";
import { LogoMark } from "@/components/brand/logo-mark";
import { LoginForm } from "@/features/authentication/components/login-form";

export const metadata: Metadata = { title: "Masuk" };

// Halaman login — panel logo Kawah Putih di kiri, kartu form di kanan.
export default function LoginPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-[#f3f8fc] px-4 py-10 sm:px-6 lg:px-8">
      <div className="w-full max-w-5xl">
        <div className="grid items-stretch gap-8 md:grid-cols-[1.15fr_1fr] lg:gap-12">
          {/* Panel logo Kawah Putih */}
          <div className="relative flex flex-col items-center justify-center gap-7 overflow-hidden rounded-2xl border border-[#d7e3ed] bg-gradient-to-br from-[#e7f2fb] via-[#f3f8fc] to-[#dceaf8] px-10 py-14 text-center">
            <LogoMark className="h-28 w-44" />
            <div>
              <p className="font-heading text-2xl font-extrabold tracking-wide text-[#145b98]">
                KAWAH PUTIH
              </p>
              <p className="mt-1 text-sm text-[#698097]">Rancabali, Bandung Selatan</p>
            </div>
          </div>

          {/* Kartu login */}
          <div className="flex flex-col justify-center rounded-2xl border border-[#d7e3ed] bg-white px-8 py-10 shadow-sm sm:px-10 sm:py-12">
            <h1 className="font-heading text-3xl font-bold tracking-wide text-[#145b98]">
              Masuk
            </h1>
            <p className="mt-2 text-sm leading-relaxed text-[#698097]">
              Masuk untuk melanjutkan perjalananmu di Kawah Putih.
            </p>

            <div className="mt-8">
              <LoginForm />
            </div>
          </div>
        </div>

        <p className="mt-6 text-center text-sm text-[#94A3B8]">
          <Link href="/" className="hover:underline">
            ← Kembali ke beranda
          </Link>
        </p>
      </div>
    </main>
  );
}
