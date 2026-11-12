import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight, Leaf, MapPin, Ticket, Clock, Quote } from "lucide-react";
import { FaqAccordion } from "@/features/public/components/faq-accordion";
import {
  aboutHistory,
  aboutSections,
  homeTicketOptions,
  visitorStories,
} from "@/lib/content";

export const metadata: Metadata = {
  title: "Beranda",
  description: "Temukan pesona Kawah Putih — pesan tiket wisata online dengan mudah.",
};

const TICKET_ICONS = [Ticket, MapPin, Leaf];

export default function HomePage() {
  return (
    <div>
      {/* Hero */}
      <section className="border-b border-[#d7e3ed] bg-gradient-to-b from-[#eef6fc] to-white">
        <div className="page-shell grid items-center gap-10 py-14 lg:grid-cols-2 lg:py-20">
          <div>
            <p className="text-xs font-bold uppercase tracking-wide text-[#1768ad]">Website Wisata</p>
            <h1 className="mt-3 font-heading text-4xl font-bold leading-tight text-[#17324d] sm:text-5xl">
              Temukan Pesona Kawah Putih
            </h1>
            <p className="mt-4 max-w-lg text-[15px] leading-relaxed text-[#698097]">
              Keindahan alam, pengalaman tak terlupakan. Jelajahi destinasi, temukan informasi, dan pesan
              tiketmu dengan mudah.
            </p>
            <div className="mt-7 flex flex-wrap gap-3">
              <Link
                href="/packages"
                className="inline-flex h-11 items-center gap-2 rounded-md bg-[#1768ad] px-6 text-sm font-semibold text-white hover:bg-[#14548f]"
              >
                Pesan Tiket <ArrowRight className="h-4 w-4" />
              </Link>
              <Link
                href="/about"
                className="inline-flex h-11 items-center rounded-md border border-[#d7e3ed] bg-white px-6 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
              >
                Kenali Kawah Putih
              </Link>
            </div>
            <div className="mt-6 flex flex-wrap gap-x-6 gap-y-2 text-xs text-[#698097]">
              <span className="inline-flex items-center gap-1.5"><Clock className="h-3.5 w-3.5" /> Buka 07.00 – 17.00 WIB</span>
              <span className="inline-flex items-center gap-1.5"><MapPin className="h-3.5 w-3.5" /> Ciwidey, Bandung</span>
            </div>
          </div>

          <div className="relative h-64 overflow-hidden rounded-lg border border-[#d7e3ed] bg-gradient-to-br from-[#1768ad] via-[#2f86c9] to-[#8fc7e9] lg:h-80">
            <div className="absolute inset-0 opacity-30 [background:radial-gradient(circle_at_30%_20%,white,transparent_55%)]" />
            <div className="absolute bottom-5 left-5 rounded-md bg-white/95 px-4 py-3">
              <p className="text-[11px] font-semibold text-[#698097]">Danau kawah vulkanik</p>
              <p className="text-sm font-bold text-[#17324d]">Gunung Patuha, Jawa Barat</p>
            </div>
          </div>
        </div>
      </section>

      {/* Pilih Tiketmu */}
      <section className="page-shell py-14">
        <div className="grid gap-8 lg:grid-cols-[1fr_1.3fr]">
          <div>
            <h2 className="font-heading text-2xl font-bold text-[#17324d]">Pilih Tiketmu</h2>
            <p className="mt-3 text-[15px] leading-relaxed text-[#698097]">
              Temukan pilihan tiket yang sesuai dengan rencana kunjunganmu dan nikmati perjalanan ke Kawah
              Putih dengan lebih mudah.
            </p>
            <Link
              href="/packages"
              className="mt-5 inline-flex items-center gap-2 text-sm font-semibold text-[#1768ad] hover:underline"
            >
              Lihat semua tiket <ArrowRight className="h-4 w-4" />
            </Link>
          </div>

          <div className="grid gap-3 sm:grid-cols-3">
            {homeTicketOptions.map((opt, i) => {
              const Icon = TICKET_ICONS[i] ?? Ticket;
              return (
                <div key={opt.key} className="flex flex-col rounded-md border border-[#d7e3ed] bg-white p-4">
                  <span className="grid h-9 w-9 place-items-center rounded-md bg-[#e7f2fb] text-[#1768ad]">
                    <Icon className="h-5 w-5" />
                  </span>
                  <p className="mt-3 text-sm font-bold text-[#17324d]">{opt.title}</p>
                  <p className="mt-1 flex-1 text-xs leading-relaxed text-[#698097]">{opt.desc}</p>
                  <p className="mt-3 text-sm font-bold text-[#1768ad]">{opt.price}</p>
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* Kenali Kawah Putih */}
      <section className="border-y border-[#d7e3ed] bg-[#f7fbff]">
        <div className="page-shell grid items-center gap-8 py-14 lg:grid-cols-2">
          <div className="order-2 h-56 rounded-lg border border-[#d7e3ed] bg-gradient-to-br from-[#7fb8dd] to-[#1768ad] lg:order-1" />
          <div className="order-1 lg:order-2">
            <h2 className="font-heading text-2xl font-bold text-[#17324d]">Kenali Kawah Putih Lebih Dekat</h2>
            <p className="mt-3 text-[15px] leading-relaxed text-[#698097]">
              Temukan cerita, sejarah, fasilitas, dan berbagai informasi menarik tentang Kawah Putih.
            </p>
            <div className="mt-5 grid gap-3 sm:grid-cols-3">
              {aboutSections.map((s) => (
                <div key={s.key} className="rounded-md border border-[#d7e3ed] bg-white p-3">
                  <p className="text-xs font-bold text-[#17324d]">{s.title}</p>
                </div>
              ))}
            </div>
            <Link
              href="/about"
              className="mt-5 inline-flex items-center gap-2 text-sm font-semibold text-[#1768ad] hover:underline"
            >
              Selengkapnya <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>

      {/* Testimonials */}
      <section className="page-shell py-14">
        <div className="mb-6 max-w-xl">
          <h2 className="font-heading text-2xl font-bold text-[#17324d]">Cerita dari Mereka yang Sudah Berkunjung</h2>
          <p className="mt-2 text-[15px] text-[#698097]">
            Lihat pengalaman pengunjung dan temukan kesan mereka setelah menikmati pesona Kawah Putih.
          </p>
        </div>
        <div className="grid gap-3 md:grid-cols-3">
          {visitorStories.map((s) => (
            <figure key={s.name} className="rounded-md border border-[#d7e3ed] bg-white p-5">
              <Quote className="h-6 w-6 text-[#bcd8ee]" />
              <blockquote className="mt-3 text-sm leading-relaxed text-[#29445e]">“{s.quote}”</blockquote>
              <figcaption className="mt-4 flex items-center gap-2">
                <span className="grid h-8 w-8 place-items-center rounded-full bg-[#1768ad] text-xs font-bold text-white">
                  {s.name[0]}
                </span>
                <span className="text-xs">
                  <span className="block font-semibold text-[#17324d]">{s.name}</span>
                  <span className="block text-[#698097]">{s.origin}</span>
                </span>
              </figcaption>
            </figure>
          ))}
        </div>
      </section>

      {/* Sejarah / About */}
      <section className="border-y border-[#d7e3ed] bg-[#f7fbff]">
        <div className="page-shell grid gap-8 py-14 lg:grid-cols-[1fr_1.4fr]">
          <div>
            <h2 className="font-heading text-2xl font-bold text-[#17324d]">Tentang Kawah Putih</h2>
            <p className="mt-3 text-[15px] text-[#698097]">
              Mengenal lebih dekat pesona alam dan cerita di balik Kawah Putih.
            </p>
            <Link
              href="/about"
              className="mt-5 inline-flex items-center gap-2 text-sm font-semibold text-[#1768ad] hover:underline"
            >
              Baca selengkapnya <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
          <div className="rounded-md border border-[#d7e3ed] bg-white p-6">
            <p className="text-xs font-bold uppercase tracking-wide text-[#1768ad]">Sejarah</p>
            <p className="mt-3 text-sm leading-relaxed text-[#29445e]">{aboutHistory}</p>
            <p className="mt-4 text-xs text-[#94A3B8]">Informasi destinasi diperbarui secara berkala.</p>
          </div>
        </div>
      </section>

      {/* FAQ */}
      <section className="page-shell py-14">
        <div className="mb-6 max-w-xl">
          <h2 className="font-heading text-2xl font-bold text-[#17324d]">FAQ</h2>
          <p className="mt-2 text-[15px] text-[#698097]">
            Pertanyaan yang sering diajukan pengunjung Kawah Putih
          </p>
        </div>
        <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
          <FaqAccordion />
          <div className="rounded-md border border-[#1768ad] bg-[#1768ad] p-6 text-white">
            <h3 className="text-lg font-bold">Siap berkunjung?</h3>
            <p className="mt-2 text-sm text-blue-100">
              Pesan tiketmu sekarang dan dapatkan e-tiket QR langsung setelah pembayaran.
            </p>
            <Link
              href="/packages"
              className="mt-5 inline-flex h-10 items-center gap-2 rounded-md bg-white px-5 text-sm font-semibold text-[#1768ad] hover:bg-blue-50"
            >
              Pesan Tiket <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
