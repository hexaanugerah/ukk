import { create } from "zustand";
import { persist } from "zustand/middleware";

// Sama seperti notification.store.ts: BRD mencantumkan wishlist sebagai user story (FR-032), tapi belum ada modul backend. Dipersist ke localStorage agar tetap berguna.
interface WishlistState {
  packageIds: string[];
  toggle: (packageId: string) => void;
  isSaved: (packageId: string) => boolean;
}

export const useWishlistStore = create<WishlistState>()(
  persist(
    (set, get) => ({
      packageIds: [],
      toggle: (packageId) =>
        set((state) => ({
          packageIds: state.packageIds.includes(packageId)
            ? state.packageIds.filter((id) => id !== packageId)
            : [...state.packageIds, packageId],
        })),
      isSaved: (packageId) => get().packageIds.includes(packageId),
    }),
    { name: "kpr-wishlist" }
  )
);
