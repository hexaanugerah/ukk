import type { Metadata } from "next";
import { GalleryAdmin } from "@/features/gallery/components/gallery-admin";

export const metadata: Metadata = { title: "Kelola Galeri" };

export default function AdminGalleryPage() {
  return (
    <div>
      <h1 className="mb-6 text-xl font-semibold">Kelola Galeri</h1>
      <GalleryAdmin />
    </div>
  );
}
