import { create } from "zustand";
import { persist } from "zustand/middleware";

// BRD mensyaratkan dukungan multi-bahasa (Fase 5). Belum ada library i18n, jadi store ini hanya menyimpan preferensi. Default id karena pengguna utama pengunjung Indonesia.
interface LanguageState {
  language: "id" | "en";
  setLanguage: (lang: "id" | "en") => void;
}

export const useLanguageStore = create<LanguageState>()(
  persist(
    (set) => ({
      language: "id",
      setLanguage: (language) => set({ language }),
    }),
    { name: "kpr-language" }
  )
);
