import type { Metadata } from "next";
import Link from "next/link";
import { Car, Coffee, Clock3, Info, Landmark, ShieldCheck, ShowerHead, TentTree, UtensilsCrossed, Waves } from "lucide-react";
import { facilityService } from "@/services/facility.service";

export const metadata: Metadata = {
  title: "Fasilitas",
  description: "Fasilitas dan layanan yang tersedia untuk kenyamanan pengunjung Kawah Putih Rancabali.",
};

// Konten fasilitas dikelola admin — render saat request.
export const dynamic = "force-dynamic";

const ICON_MAP: Record<string, typeof Car> = {
  car: Car,
  coffee: Coffee,
  clock3: Clock3,
  info: Info,
  landmark: Landmark,
  shieldcheck: ShieldCheck,
  showerhead: ShowerHead,
  tenttree: TentTree,
  utensilscrossed: UtensilsCrossed,
  waves: Waves,
};

export default async function FacilitiesPage() {
  const facilities = await facilityService.listPublished();

  return (
    <div className="page-shell py-12">
      <div className="mb-8 max-w-2xl">
        <span className="inline-flex items-center gap-2 rounded-full bg-[#e9f3fb] px-3 py-1 text-xs font-semibold text-[#1768ad]">
          <Info className="h-3.5 w-3.5" /> Informasi Wisata
        </span>
        <h1 className="mt-4 font-heading text-3xl font-bold text-[#17324d]">Fasilitas Kawah Putih</h1>
        <p className="mt-3 text-[15px] leading-relaxed text-[#698097]">
          Fasilitas dan layanan yang tersedia untuk kenyamanan pengunjung
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {facilities.length === 0 ? (
          <p className="col-span-full text-sm text-slate-500">Belum ada fasilitas yang dipublikasikan.</p>
        ) : (
          facilities.map((facility) => {
            const Icon = ICON_MAP[facility.icon.toLowerCase()] || Info;
            return (
              <div key={facility.id} className="rounded-md border border-[#d7e3ed] bg-white p-5 transition-colors hover:border-[#1768ad]">
                <span className="grid h-9 w-9 place-items-center rounded-md bg-[#e7f2fb] text-[#1768ad]">
                  <Icon className="h-4 w-4" aria-hidden="true" />
                </span>
                <h2 className="mt-3 text-[15px] font-bold text-[#17324d]">{facility.name}</h2>
                <p className="mt-1.5 text-sm leading-relaxed text-[#698097]">{facility.description || "Tidak ada deskripsi"}</p>
                {facility.location && (
                  <p className="mt-1.5 text-xs text-slate-500">
                    <span className="font-medium">Lokasi:</span> {facility.location}
                  </p>
                )}
                {facility.operating_hours && (
                  <p className="mt-1 text-xs text-slate-500">
                    <span className="font-medium">Jam:</span> {facility.operating_hours}
                  </p>
                )}
              </div>
            );
          })
        )}
      </div>

      <div className="mt-8 flex flex-col items-center justify-between gap-4 rounded-md bg-[#1768ad] p-6 text-white sm:flex-row">
        <div>
          <h2 className="text-lg font-bold">Siap berkunjung?</h2>
          <p className="mt-1 text-sm text-blue-100">Pesan tiket online sekarang dan hindari antre di lokasi.</p>
        </div>
        <Link
          href="/packages"
          className="inline-flex h-10 shrink-0 items-center justify-center rounded-md bg-white px-5 text-sm font-bold text-[#14548f] hover:bg-blue-50"
        >
          Pesan Tiket Sekarang
        </Link>
      </div>
    </div>
  );
}