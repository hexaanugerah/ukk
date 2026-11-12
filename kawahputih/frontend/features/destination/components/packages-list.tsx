"use client";

import { usePackages } from "@/hooks/use-destinations";
import { CardGridSkeleton } from "@/components/loading/skeleton";
import { EmptyState } from "@/components/empty/empty-state";
import { PackageCard } from "./package-card";

export function PackagesList() {
  const { data, isLoading, isError } = usePackages({ active_only: true, limit: 9 });
  const packages = data ?? [];

  if (isLoading) return <CardGridSkeleton count={9} />;
  if (isError) return <EmptyState title="Gagal memuat paket" description="Coba muat ulang halaman." />;
  if (packages.length === 0) return <EmptyState title="Belum ada paket wisata tersedia" />;

  return (
    <div className="space-y-8">
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {packages.map((pkg) => (
          <PackageCard key={pkg.id} pkg={pkg} />
        ))}
      </div>
    </div>
  );
}
