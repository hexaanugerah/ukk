import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight, Building2, Info, ShieldCheck } from "lucide-react";
import { destinationInfoService } from "@/services/destination-info.service";

export const metadata: Metadata = {
  title: "Tentang Kawah Putih",
  description: "Mengenal lebih dekat pesona alam dan cerita di balik Kawah Putih.",
};

// Konten destination info dikelola admin — render saat request.
export const dynamic = "force-dynamic";

const SECTION_ICONS = [Building2, Info, ShieldCheck];

export default async function AboutPage() {
  const destinationInfo = await destinationInfoService.getPublished();

  return (
    <div>
      <section className="border-b border-[#d7e3ed] bg-gradient-to-b from-[#eef6fc] to-white">
        <div className="page-shell py-12">
          <h1 className="font-heading text-3xl font-bold text-[#17324d] sm:text-4xl">Tentang Kawah Putih</h1>
          <p className="mt-3 max-w-2xl text-[15px] leading-relaxed text-[#698097]">
            Mengenal lebih dekat pesona alam dan cerita di balik Kawah Putih.
          </p>
        </div>
      </section>

      <section className="page-shell grid gap-8 py-12 lg:grid-cols-[1fr_1.3fr]">
        <div className="h-64 rounded-lg border border-[#d7e3ed] bg-gradient-to-br from-[#8fc7e9] via-[#2f86c9] to-[#1768ad] lg:h-auto" />
        <div>
          <p className="text-xs font-bold uppercase tracking-wide text-[#1768ad]">Sejarah</p>
          <h2 className="mt-2 font-heading text-2xl font-bold text-[#17324d]">Cerita di Balik Kawah Putih</h2>
          <p className="mt-4 text-sm leading-relaxed text-[#29445e]">
            {destinationInfo?.history || "Kawah Putih merupakan danau kawah vulkanik yang berada di kawasan Gunung Patuha, Ciwidey, Jawa Barat. Danau ini terbentuk dari aktivitas vulkanik dan memiliki air berwarna putih kehijauan yang dapat berubah mengikuti kondisi alam. Keunikan warna dan suasana berkabutnya menjadikan Kawah Putih salah satu destinasi wisata alam yang menarik di Jawa Barat."}
          </p>
          <p className="mt-4 text-xs text-[#94A3B8]">Informasi destinasi diperbarui secara berkala.</p>
        </div>
      </section>

      <section className="border-t border-[#d7e3ed] bg-[#f7fbff]">
        <div className="page-shell py-12">
          <div className="mb-6 max-w-xl">
            <h2 className="font-heading text-2xl font-bold text-[#17324d]">Fasilitas & Informasi</h2>
            <p className="mt-2 text-sm text-[#698097]">
              Hal-hal yang perlu kamu tahu sebelum berkunjung ke Kawah Putih.
            </p>
          </div>

          <div className="grid gap-4 md:grid-cols-3">
            {destinationInfo ? [
              { key: "fasilitas", title: "Fasilitas", body: destinationInfo.additional_info || destinationInfo.description },
              { key: "informasi", title: "Informasi Destinasi", body: destinationInfo.destination_desc || destinationInfo.description },
              { key: "aturan", title: "Aturan Kunjungan", body: "Jaga kebersihan kawasan, ikuti jalur yang tersedia, dilarang berenang, dan selalu patuhi arahan petugas di lapangan." },
            ].map((s, i) => {
              const Icon = SECTION_ICONS[i] ?? Info;
              return (
                <article key={s.key} className="rounded-md border border-[#d7e3ed] bg-white p-5">
                  <span className="grid h-10 w-10 place-items-center rounded-md bg-[#e7f2fb] text-[#1768ad]">
                    <Icon className="h-5 w-5" />
                  </span>
                  <h3 className="mt-4 text-sm font-bold text-[#17324d]">{s.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-[#698097]">{s.body}</p>
                </article>
              );
            }) : (
              ""
            )}
          </div>
        </div>
      </section>

      {destinationInfo && (
        <section className="page-shell py-12">
          <div className="rounded-lg border border-[#d7e3ed] bg-white p-6">
            <h2 className="font-heading text-xl font-bold text-[#17324d]">Lokasi & Kontak</h2>
            <div className="mt-4 grid gap-4 md:grid-cols-2">
              {destinationInfo.location && (
                <div>
                  <h3 className="text-sm font-bold text-[#17324d]">Lokasi</h3>
                  <p className="mt-1 text-sm text-[#698097]">{destinationInfo.location}</p>
                </div>
              )}
              {destinationInfo.operating_hours && (
                <div>
                  <h3 className="text-sm font-bold text-[#17324d]">Jam Operasional</h3>
                  <p className="mt-1 text-sm text-[#698097]">{destinationInfo.operating_hours}</p>
                </div>
              )}
              {destinationInfo.contact && (
                <div>
                  <h3 className="text-sm font-bold text-[#17324d]">Kontak</h3>
                  <p className="mt-1 text-sm text-[#698097]">{destinationInfo.contact}</p>
                </div>
              )}
            </div>
          </div>
        </section>
      )}

      <section className="page-shell py-12">
        <div className="flex flex-col items-start justify-between gap-4 rounded-lg border border-[#1768ad] bg-[#1768ad] p-6 text-white sm:flex-row sm:items-center">
          <div>
            <h2 className="text-lg font-bold">Siap menjelajahi Kawah Putih?</h2>
            <p className="mt-1 text-sm text-blue-100">
              Pesan tiketmu sekarang dan nikmati kemudahan berkunjung.
            </p>
          </div>
          <Link
            href="/packages"
            className="inline-flex h-10 items-center gap-2 rounded-md bg-white px-5 text-sm font-semibold text-[#1768ad] hover:bg-blue-50"
          >
            Pesan Tiket <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </section>
    </div>
  );
}