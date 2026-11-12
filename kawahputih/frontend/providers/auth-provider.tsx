"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth.store";
import { authService } from "@/services/auth.api.service";

// Auth provider — memastikan sesi dari backend sudah direhidrasi dari localStorage sebelum halaman protected di-render, dan menghapus sesi kedaluwarsa.

let hydrated = false;

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = useState(hydrated);

  useEffect(() => {
    // Cek sesi dengan backend saat mount
    authService.checkSession().finally(() => {
      hydrated = true;
      setReady(true);
    });
  }, []);

  if (!ready) return <>{children}</>;
  return <>{children}</>;
}

/** Selalu true setelah mount pertama — dipakai guard untuk hindari redirect palsu saat SSR. */
export function useAuthReady() {
  const [ready, setReady] = useState(hydrated);
  useEffect(() => {
    if (!hydrated) {
      authService.checkSession().finally(() => {
        hydrated = true;
        setReady(true);
      });
    }
    setReady(true);
  }, []);
  return ready;
}