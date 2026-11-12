"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  Users,
  FileText,
  Image as ImageIcon,
  Package,
  Ticket,
} from "lucide-react";

const LINKS = [
  { href: "/admin/dashboard", label: "Ringkasan", icon: LayoutDashboard },
  { href: "/admin/users", label: "Pengguna", icon: Users },
  { href: "/admin/articles", label: "Artikel", icon: FileText },
  { href: "/admin/gallery", label: "Galeri", icon: ImageIcon },
  { href: "/admin/packages", label: "Paket Wisata", icon: Package },
  { href: "/admin/bookings", label: "Booking", icon: Ticket },
];

export function AdminSidebar() {
  const pathname = usePathname();
  return (
    <aside className="w-full border-b border-slate-200 p-4 dark:border-slate-800 md:h-screen md:w-56 md:border-b-0 md:border-r">
      <p className="mb-3 px-3 text-xs font-semibold uppercase tracking-wider text-slate-400">
        Admin Panel
      </p>
      <nav className="flex gap-1 overflow-x-auto md:flex-col md:overflow-visible">
        {LINKS.map(({ href, label, icon: Icon }) => (
          <Link
            key={href}
            href={href}
            className={cn(
              "flex items-center gap-2 whitespace-nowrap rounded-lg px-3 py-2 text-sm",
              pathname === href
                ? "bg-brand-50 font-medium text-brand-700 dark:bg-brand-950"
                : "hover:bg-slate-100 dark:hover:bg-slate-800",
            )}
          >
            <Icon className="h-4 w-4" /> {label}
          </Link>
        ))}
      </nav>
    </aside>
  );
}
