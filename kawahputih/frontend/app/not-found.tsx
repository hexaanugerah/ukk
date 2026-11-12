import Link from "next/link";
import { cn } from "@/lib/utils";

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 px-4 text-center">
      <p className="text-6xl font-bold text-brand-600">404</p>
      <h1 className="text-lg font-semibold">Halaman tidak ditemukan</h1>
      <p className="max-w-sm text-sm text-slate-500">
        Halaman yang kamu cari tidak ada atau sudah dipindahkan.
      </p>
      <Link
        href="/"
        className={cn(
          "inline-flex h-10 items-center justify-center rounded-xl bg-brand-600 px-4 text-sm font-medium text-white hover:bg-brand-700",
        )}
      >
        Kembali ke Beranda
      </Link>
    </div>
  );
}
