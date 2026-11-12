"use client";

// ConfirmDialog — dialog konfirmasi sederhana untuk aksi destruktif (hapus).
// Tanpa dependency tambahan: overlay fixed + kartu terpusat, tertutup saat
// klik overlay atau tekan Escape. Dipakai di semua halaman kelola konten
// agar penghapusan tidak langsung terjadi saat tombol ditekan.

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Trash2 } from "lucide-react";

interface ConfirmDialogProps {
  open: boolean;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  /** Tombol konfirmasi berputar saat mutasi sedang berjalan. */
  loading?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel = "Hapus",
  cancelLabel = "Batal",
  loading = false,
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  // Tutup dengan Escape + kunci scroll di belakang overlay.
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape" && !loading) onCancel();
    };
    document.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = "";
    };
  }, [open, loading, onCancel]);

  if (!open) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={title}
      className="fixed inset-0 z-50 grid place-items-center bg-slate-950/50 p-4"
      onClick={() => {
        if (!loading) onCancel();
      }}
    >
      <div
        className="w-full max-w-md rounded-xl bg-white p-5 shadow-xl dark:bg-slate-900"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start gap-3">
          <div className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-red-100 text-red-600 dark:bg-red-900/40">
            <Trash2 className="h-5 w-5" />
          </div>
          <div className="min-w-0">
            <h3 className="text-base font-bold text-slate-900 dark:text-white">{title}</h3>
            {description && (
              <p className="mt-1 text-sm leading-relaxed text-slate-600 dark:text-slate-400">{description}</p>
            )}
          </div>
        </div>
        <div className="mt-5 flex justify-end gap-2">
          <Button variant="outline" onClick={onCancel} disabled={loading}>
            {cancelLabel}
          </Button>
          <Button variant="destructive" onClick={onConfirm} loading={loading}>
            {confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  );
}
