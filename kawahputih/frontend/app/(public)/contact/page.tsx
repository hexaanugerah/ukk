import type { Metadata } from "next";
import { Mail, MapPin, Phone, Clock } from "lucide-react";
import Link from "next/link";
import { APP_CONFIG } from "@/config/app.config";
import { destinationInfoService } from "@/services/destination-info.service";

export const metadata: Metadata = {
  title: "Kontak",
  description: "Hubungi pengelola Kawah Putih Rancabali.",
};

// Konten destination info dikelola admin — render saat request.
export const dynamic = "force-dynamic";

// Halaman kontak — alamat, telepon, dan email dari konfigurasi aplikasi; jam
// operasional & kontak tambahan dari konten destination info yang dikelola admin.
export default async function ContactPage() {
  const destinationInfo = await destinationInfoService.getPublished();
  const contact = APP_CONFIG.contact;
  const hours = destinationInfo?.operating_hours || `${APP_CONFIG.operationalHour.open} - ${APP_CONFIG.operationalHour.close} WIB`;
  const phone = destinationInfo?.contact || contact.phone;

  const items = [
    { icon: MapPin, title: "Alamat", body: contact.address },
    { icon: Phone, title: "Telepon", body: phone },
    { icon: Mail, title: "Email", body: contact.email },
    { icon: Clock, title: "Jam Operasional", body: hours },
  ];

  return (
    <div>
      <section className="border-b border-[#d7e3ed] bg-gradient-to-b from-[#eef6fc] to-white">
        <div className="page-shell py-12">
          <h1 className="font-heading text-3xl font-bold text-[#17324d] sm:text-4xl">Kontak</h1>
          <p className="mt-3 max-w-2xl text-[15px] leading-relaxed text-[#698097]">
            Hubungi pengelola Kawah Putih untuk pertanyaan seputar kunjungan, tiket, atau layanan lainnya.
          </p>
        </div>
      </section>

      <section className="page-shell py-12">
        <div className="grid gap-4 md:grid-cols-2">
          {items.map(({ icon: Icon, title, body }) => (
            <article key={title} className="rounded-md border border-[#d7e3ed] bg-white p-5">
              <span className="grid h-10 w-10 place-items-center rounded-md bg-[#e7f2fb] text-[#1768ad]">
                <Icon className="h-5 w-5" />
              </span>
              <h2 className="mt-4 text-sm font-bold text-[#17324d]">{title}</h2>
              <p className="mt-1 text-sm leading-relaxed text-[#698097]">{body}</p>
            </article>
          ))}
        </div>

        <div className="mt-8 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-5 text-sm text-[#29445e]">
          <p className="font-semibold text-[#17324d]">Butuh bantuan cepat?</p>
          <p className="mt-1 text-[#698097]">
            Pertanyaan umum sudah dijawab di halaman{" "}
            <Link href="/faq" className="font-semibold text-[#1768ad] hover:underline">FAQ</Link>, atau lihat{" "}
            <Link href="/visitor-rules" className="font-semibold text-[#1768ad] hover:underline">aturan kunjungan</Link>{" "}
            sebelum berangkat.
          </p>
        </div>
      </section>
    </div>
  );
}
