import type { Metadata } from "next";
import { FacilityAdmin } from "@/features/facility/components/facility-admin";

export const metadata: Metadata = { title: "Kelola Fasilitas" };

export default function AdminFacilityPage() {
  return (
    <div>
      <h1 className="mb-6 text-xl font-semibold">Kelola Fasilitas</h1>
      <FacilityAdmin />
    </div>
  );
}