import Link from "next/link";
import { Mountain } from "lucide-react";

export function Footer() {
  return (
    <footer className="mt-16 border-t border-[#d7e3ed] bg-white">
      <div className="page-shell grid gap-8 py-10 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <div className="flex items-center gap-2.5">
            <span className="grid h-9 w-9 place-items-center rounded-md bg-[#e7f2fb] text-[#1768ad]">
              <Mountain className="h-5 w-5" strokeWidth={2.4} />
            </span>
            <span className="text-[15px] font-bold text-[#145b98]">Kawah Putih</span>
          </div>
          <p className="mt-3 text-sm leading-relaxed text-[#698097]">
            Keindahan alam, pengalaman tak terlupakan. Jelajahi destinasi, temukan informasi, dan pesan tiketmu dengan mudah.
          </p>
        </div>

        <div>
          <p className="text-sm font-bold text-[#17324d]">Jelajahi</p>
          <ul className="mt-3 space-y-2 text-sm text-[#698097]">
            <li><Link href="/" className="hover:text-[#1768ad]">Beranda</Link></li>
            <li><Link href="/packages" className="hover:text-[#1768ad]">Pilih Tiket</Link></li>
            <li><Link href="/gallery" className="hover:text-[#1768ad]">Galeri</Link></li>
            <li><Link href="/articles" className="hover:text-[#1768ad]">Artikel</Link></li>
          </ul>
        </div>

        <div>
          <p className="text-sm font-bold text-[#17324d]">Informasi</p>
          <ul className="mt-3 space-y-2 text-sm text-[#698097]">
            <li><Link href="/about" className="hover:text-[#1768ad]">Tentang Kawah Putih</Link></li>
            <li><Link href="/faq" className="hover:text-[#1768ad]">FAQ</Link></li>
            <li><Link href="/contact" className="hover:text-[#1768ad]">Kontak</Link></li>
          </ul>
        </div>

        <div>
          <p className="text-sm font-bold text-[#17324d]">Kontak</p>
          <ul className="mt-3 space-y-2 text-sm text-[#698097]">
            <li>Ciwidey, Kab. Bandung, Jawa Barat</li>
            <li>info@kawahputih.id</li>
            <li>07.00 – 17.00 WIB</li>
          </ul>
        </div>
      </div>
      <div className="border-t border-[#e5edf3]">
        <div className="page-shell py-4 text-center text-xs text-[#94A3B8]">
          &copy; {new Date().getFullYear()} Kawah Putih Rancabali Tourism. Semua hak dilindungi.
        </div>
      </div>
    </footer>
  );
}
