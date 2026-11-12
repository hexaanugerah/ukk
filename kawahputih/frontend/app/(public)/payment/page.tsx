import type { Metadata } from "next";
import { PaymentPanel } from "@/features/public/components/payment-panel";

export const metadata: Metadata = {
  title: "Metode Pembayaran",
  description: "Pilih metode pembayaran dan unggah bukti pembayaran tiket Kawah Putih.",
};

export default async function PaymentPage({
  searchParams,
}: {
  searchParams: Promise<{ code?: string }>;
}) {
  const { code } = await searchParams;
  return (
    <div className="page-shell py-10">
      <div className="mb-7">
        <h1 className="font-heading text-2xl font-bold text-[#17324d]">Metode Pembayaran</h1>
        <p className="mt-1 text-sm text-[#698097]">Pilih metode pembayaran yang Anda inginkan</p>
      </div>
      <PaymentPanel code={code} />
    </div>
  );
}
