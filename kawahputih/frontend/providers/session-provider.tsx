"use client";

// Proyek ini TIDAK memakai SessionProvider NextAuth. Backend menerbitkan
// pasangan token JWT access/refresh sendiri, dan frontend menyimpan sesi di
// store/auth.store.ts (Zustand, dipersist ke localStorage), bukan cookie
// sesi sisi server. Komponen ini hanya memenuhi struktur providers/ dan
// menjadi tempat rehidrasi sesi nyata bila nanti ditambahkan validasi sesi
// di sisi server.
export function SessionProvider({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
