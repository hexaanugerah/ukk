import type { Metadata } from "next";
import { Suspense } from "react";
import Link from "next/link";
import { BookingOrderForm } from "@/features/public/components/booking-order-form";

export const metadata: Metadata = {
  title: "Pemesanan",
  description: "Lengkapi data pemesanan tiket Kawah Putih.",
};

export default function BookingPage() {
  return (
    <div className="page-shell py-10">
      <div className="mb-7 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="font-heading text-2xl font-bold text-[#17324d]">Pemesanan Anda</h1>
          <p className="mt-1 text-sm text-[#698097]">
            Lengkapi data pemesanan untuk melanjutkan ke pembayaran.
          </p>
        </div>
      </div>

      <Suspense
        fallback={
          <div className="grid gap-4 lg:grid-cols-[1.6fr_1fr]">
            <div className="h-72 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />
            <div className="h-72 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />
          </div>
        }
      >
        <BookingOrderForm />
      </Suspense>
    </div>
  );
}
