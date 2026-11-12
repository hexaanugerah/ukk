import type { Metadata } from "next";
import { destinationService } from "@/services/destination.service";
import { BookingForm } from "@/features/booking/components/booking-form";
import { formatCurrency } from "@/lib/currency";
import { Clock, Users } from "lucide-react";
import { notFound } from "next/navigation";

interface Props {
  params: Promise<{ id: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const pkg = await destinationService.get(id);
  if (!pkg) return { title: "Paket tidak ditemukan" };
  return {
    title: pkg.name,
    description: pkg.description.slice(0, 160),
    openGraph: { images: pkg.cover_image ? [pkg.cover_image] : [] },
  };
}

export default async function PackageDetailPage({ params }: Props) {
  const { id } = await params;
  const pkg = await destinationService.get(id);
  if (!pkg) notFound();

  return (
    <div className="page-shell py-10">
      <div className="grid grid-cols-1 gap-10 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <div className="mb-4 h-72 w-full overflow-hidden rounded-2xl bg-slate-100 dark:bg-slate-800">
            {pkg.cover_image && (
              // eslint-disable-next-line @next/next/no-img-element -- server component, dynamic remote source, next/image would need remotePatterns confirmed at runtime
              <img src={pkg.cover_image} alt={pkg.name} className="h-full w-full object-cover" />
            )}
          </div>
          <h1 className="text-2xl font-semibold">{pkg.name}</h1>
          <div className="mt-2 flex items-center gap-4 text-sm text-slate-500">
            <span className="flex items-center gap-1">
              <Clock className="h-4 w-4" /> {pkg.duration_hours} jam
            </span>
            {pkg.max_capacity > 0 && (
              <span className="flex items-center gap-1">
                <Users className="h-4 w-4" /> Maks {pkg.max_capacity} orang/hari
              </span>
            )}
          </div>
          <p className="mt-4 whitespace-pre-line text-slate-600 dark:text-slate-400">{pkg.description}</p>
        </div>

        <div>
          <div className="sticky top-24 rounded-2xl border border-slate-200 p-6 dark:border-slate-800">
            <p className="text-2xl font-bold text-brand-600">{formatCurrency(pkg.price_cents, pkg.currency)}</p>
            <p className="mb-4 text-sm text-slate-500">per orang</p>
            <BookingForm pkg={pkg} />
          </div>
        </div>
      </div>
    </div>
  );
}
