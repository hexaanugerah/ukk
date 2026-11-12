"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth.store";
import { useAuthReady } from "@/providers/auth-provider";
import type { AppRole } from "@/config/auth.config";

// Route guard — proteksi rute berbasis role. Belum login -> redirect /login?redirect=<path>; role tidak sesuai -> /unauthorized. Render null sampai sesi terverifikasi agar panel tidak berkedip.

export function useRouteGuard(allowed: AppRole[]): "checking" | "allowed" | "denied" {
  const router = useRouter();
  const ready = useAuthReady();
  const user = useAuthStore((s) => s.user);

  useEffect(() => {
    if (!ready) return;
    if (!user) {
      const current = window.location.pathname + window.location.search;
      router.replace(`/login?redirect=${encodeURIComponent(current)}`);
    } else if (!allowed.includes(user.role)) {
      router.replace("/unauthorized");
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, user, router]);

  if (!ready) return "checking";
  if (!user) return "checking";
  if (!allowed.includes(user.role)) return "denied";
  return "allowed";
}
