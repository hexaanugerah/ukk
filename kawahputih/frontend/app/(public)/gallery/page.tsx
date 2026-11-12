import type { Metadata } from "next";
import { galleryService } from "@/services/gallery.service";
import Image from "next/image";

export const metadata: Metadata = { title: "Galeri", description: "Galeri foto keindahan Kawah Putih Rancabali." };

export default async function GalleryPage() {
  const items = await galleryService.list({ limit: 30 });

  return (
    <div className="page-shell py-10">
      <h1 className="mb-6 text-2xl font-semibold">Galeri</h1>
      {items.length === 0 ? (
        <p className="text-sm text-slate-500">Belum ada foto di galeri.</p>
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
          {items.map((item) => (
            <div key={item.id} className="relative aspect-square overflow-hidden rounded-xl bg-slate-100 dark:bg-slate-800">
              <Image src={item.image_url} alt={item.title} fill className="object-cover transition hover:scale-105" />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
