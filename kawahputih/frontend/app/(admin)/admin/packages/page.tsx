import type { Metadata } from "next";
import { PackagesAdmin } from "@/features/destination/components/packages-admin";

export const metadata: Metadata = { title: "Kelola Paket Wisata" };

export default function AdminPackagesPage() {
  return (
    <div>
      <h1 className="mb-6 text-xl font-semibold">Kelola Paket Wisata</h1>
      <PackagesAdmin />
    </div>
  );
}
