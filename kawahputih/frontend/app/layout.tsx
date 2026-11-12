import type { Metadata } from "next";
import type { ReactNode } from "react";
import "./globals.css";
import { QueryProvider } from "@/providers/query-provider";
import { ThemeProvider } from "@/providers/theme-provider";
import { ToastProvider } from "@/providers/toast-provider";
import { AuthProvider } from "@/providers/auth-provider";
import { Poppins, Inter } from "next/font/google";
const poppins = Poppins({ subsets: ["latin"], weight: ["600", "700"], variable: "--font-poppins", display: "swap" });
const inter = Inter({ subsets: ["latin"], weight: ["400", "500", "600"], variable: "--font-inter", display: "swap" });

export const metadata: Metadata = {
  title: {
    default: "Kawah Putih Rancabali — Wisata Alam Bandung",
    template: "%s | Kawah Putih Rancabali",
  },
  description:
    "Pesan tiket wisata Kawah Putih Rancabali secara online. Informasi lengkap, harga tiket, galeri, dan artikel wisata terbaru.",
  keywords: ["kawah putih", "rancabali", "wisata bandung", "tiket wisata online"],
  openGraph: {
    type: "website",
    locale: "id_ID",
    siteName: "Kawah Putih Rancabali",
  },
  twitter: { card: "summary_large_image" },
  robots: { index: true, follow: true },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="id" suppressHydrationWarning className={`${poppins.variable} ${inter.variable}`}>
      <body>
        <QueryProvider>
          <ThemeProvider>
            <AuthProvider>
              {children}
              <ToastProvider />
            </AuthProvider>
          </ThemeProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
