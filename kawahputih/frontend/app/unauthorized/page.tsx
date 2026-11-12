import type { Metadata } from "next";
import Link from "next/link";
import { ShieldAlert } from "lucide-react";

export const metadata: Metadata = { title: "Akses Ditolak" };

export default function UnauthorizedPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-[#f3f8fc] px-4 py-10">
      <div className="w-full max-w-md rounded-md border border-[#d7e3ed] bg-white p-8 text-center shadow-sm">
        <span className="mx-auto grid h-14 w-14 place-items-center rounded-full bg-[#fdeaea] text-[#c53030]">
          <ShieldAlert className="h-7 w-7" />
        </span>
        <h1 className="mt-4 text-xl font-bold text-[#17324d]">Akses Ditolak</h1>
        <p className="mt-2 text-sm text-[#698097]">
          Anda tidak memiliki izin untuk mengakses halaman ini. Halaman tersebut khusus untuk peran tertentu.
        </p>
        <div className="mt-6 flex flex-wrap justify-center gap-2">
          <Link
            href="/"
            className="inline-flex h-10 items-center rounded-md border border-[#d7e3ed] px-5 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
          >
            Ke Beranda
          </Link>
          <Link
            href="/login"
            className="inline-flex h-10 items-center rounded-md bg-[#1768ad] px-5 text-sm font-semibold text-white hover:bg-[#14548f]"
          >
            Ganti Akun
          </Link>
        </div>
      </div>
    </main>
  );
}
