import { ApiMeta } from "@/lib/api";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

export function Pagination({ meta, onPageChange }: { meta: ApiMeta; onPageChange: (page: number) => void }) {
  if (meta.total_pages <= 1) return null;

  return (
    <div className="flex items-center justify-center gap-2">
      <button
        onClick={() => onPageChange(meta.page - 1)}
        disabled={meta.page <= 1}
        className="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-300 disabled:opacity-40 dark:border-slate-700"
        aria-label="Halaman sebelumnya"
      >
        <ChevronLeft className="h-4 w-4" />
      </button>
      <span className="text-sm text-slate-600 dark:text-slate-400">
        Halaman {meta.page} dari {meta.total_pages}
      </span>
      <button
        onClick={() => onPageChange(meta.page + 1)}
        disabled={meta.page >= meta.total_pages}
        className={cn("flex h-9 w-9 items-center justify-center rounded-lg border border-slate-300 disabled:opacity-40 dark:border-slate-700")}
        aria-label="Halaman berikutnya"
      >
        <ChevronRight className="h-4 w-4" />
      </button>
    </div>
  );
}
