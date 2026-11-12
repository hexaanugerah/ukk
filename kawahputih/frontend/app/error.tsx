"use client";

import { Button } from "@/components/ui/button";
import { AlertTriangle } from "lucide-react";

export default function Error({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 px-4 text-center">
      <AlertTriangle className="h-10 w-10 text-red-500" />
      <h1 className="text-lg font-semibold">Terjadi kesalahan</h1>
      <p className="max-w-sm text-sm text-slate-500">
        {error.message || "Sesuatu tidak berjalan sebagaimana mestinya. Coba lagi."}
      </p>
      <Button onClick={reset}>Coba lagi</Button>
    </div>
  );
}
