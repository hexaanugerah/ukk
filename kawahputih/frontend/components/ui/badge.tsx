import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

const STATUS_COLORS: Record<string, string> = {
  pending_payment: "bg-amber-100 text-amber-800",
  paid: "bg-green-100 text-green-800",
  checked_in: "bg-blue-100 text-blue-800",
  cancelled: "bg-slate-100 text-slate-600",
  expired: "bg-red-100 text-red-800",
  draft: "bg-slate-100 text-slate-600",
  published: "bg-green-100 text-green-800",
  archived: "bg-slate-100 text-slate-500",
  // Warna semantik yang dipakai tabel konten admin.
  success: "bg-green-100 text-green-800",
  warning: "bg-amber-100 text-amber-800",
  default: "bg-slate-100 text-slate-700",
  active: "bg-green-100 text-green-800",
  maintenance: "bg-amber-100 text-amber-800",
  inactive: "bg-slate-100 text-slate-600",
  low: "bg-blue-100 text-blue-800",
  normal: "bg-green-100 text-green-800",
  high: "bg-amber-100 text-amber-800",
  critical: "bg-red-100 text-red-800",
};

export function Badge({
  status,
  label,
  variant,
  className,
  children,
}: {
  status?: string;
  label?: string;
  variant?: string;
  className?: string;
  children?: ReactNode;
}) {
  const tone =
    (status && STATUS_COLORS[status]) ??
    (variant === "outline"
      ? "border border-slate-300 bg-transparent text-slate-600"
      : "bg-slate-100 text-slate-700");

  return (
    <span className={cn("inline-block rounded-full px-2.5 py-0.5 text-xs font-medium", tone, className)}>
      {children ?? label ?? (status ? status.replace(/_/g, " ") : "")}
    </span>
  );
}
