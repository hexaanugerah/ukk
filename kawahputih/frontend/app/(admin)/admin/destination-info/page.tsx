import type { Metadata } from "next";
import { DestinationInfoAdmin } from "@/features/destination-info/components/destination-info-admin";

export const metadata: Metadata = { title: "Kelola Informasi Kawah Putih" };

export default function AdminDestinationInfoPage() {
  return (
    <div>
      <h1 className="mb-6 text-xl font-semibold">Kelola Informasi Kawah Putih</h1>
      <DestinationInfoAdmin />
    </div>
  );
}