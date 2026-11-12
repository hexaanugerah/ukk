"use client";

import type { LucideIcon } from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Home, LogOut, Menu, X } from "lucide-react";
import { useEffect, useState } from "react";
import { useLogout } from "@/hooks/use-auth";
import { useAuthStore } from "@/store/auth.store";
import { LogoMark } from "@/components/brand/logo-mark";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

export interface PanelLink {
  href: string;
  label: string;
  icon: LucideIcon;
  /** Sub-item opsional yang tampil sebagai label kecil di dalam grup. */
  section?: string;
  /** Angka kecil (mis. notifikasi belum dibaca) — disembunyikan bila 0. */
  badge?: number;
}

interface NavSection {
  section?: string;
  items: PanelLink[];
}

function Brand({ role, variant = "header" }: { role: string; variant?: "header" | "drawer" }) {
  return (
    <span className="flex min-w-0 items-center gap-2.5">
      <span className="shrink-0">
        <LogoMark />
      </span>
      <span className="min-w-0 leading-tight">
        <span
          className={cn(
            "block truncate font-extrabold",
            variant === "drawer" ? "text-[15px] text-white" : "text-[14px] text-[#14548f] sm:text-[15px]"
          )}
        >
          Kawah Putih
        </span>
        <span
          className={cn(
            "block truncate text-[10px] font-medium",
            variant === "drawer" ? "text-white/70" : "text-[#14548f]"
          )}
        >
          {role} Panel
        </span>
      </span>
    </span>
  );
}

function NavGroups({
  sections,
  isActive,
  onNavigate,
  ariaLabel,
  variant,
}: {
  sections: NavSection[];
  isActive: (href: string) => boolean;
  onNavigate: () => void;
  ariaLabel: string;
  /** "full" = ikon + label, "rail" = ikon saja (tooltip lewat title/aria-label). */
  variant: "full" | "rail";
}) {
  if (variant === "rail") {
    return (
      <nav
        aria-label={ariaLabel}
        className="panel-nav min-h-0 flex-1 space-y-1 overflow-y-auto overscroll-contain px-2 py-3"
      >
        {sections.map((group, gi) => (
          <div
            key={group.section ?? gi}
            className={cn("space-y-1", gi > 0 && "border-t border-white/20 pt-2")}
          >
            {group.items.map(({ href, label, icon: Icon, badge }) => {
              const active = isActive(href);
              return (
                <Link
                  key={href}
                  href={href}
                  title={label}
                  aria-label={label}
                  aria-current={active ? "page" : undefined}
                  className={cn(
                    "relative grid h-11 w-full place-items-center rounded-md transition-colors",
                    active
                      ? "bg-[#114a82] text-white"
                      : "text-white/85 hover:bg-white/10 hover:text-white"
                  )}
                >
                  <Icon className="h-[18px] w-[18px]" aria-hidden="true" />
                  {typeof badge === "number" && badge > 0 && (
                    <span className="absolute right-1 top-1 grid h-3.5 min-w-3.5 place-items-center rounded-full bg-white px-1 text-[8px] font-bold text-[#1768ad]">
                      {badge > 99 ? "99+" : badge}
                    </span>
                  )}
                </Link>
              );
            })}
          </div>
        ))}
      </nav>
    );
  }

  return (
    <nav
      aria-label={ariaLabel}
      className="panel-nav min-h-0 flex-1 overflow-y-auto overscroll-contain px-2.5 pb-2 pt-3"
    >
      <div className="space-y-1">
        {sections.map((group, gi) => (
          <div key={group.section ?? gi} className={cn(gi > 0 && "pt-1")}>
            {group.section && (
              <p className="px-2 pb-1 pt-1 text-[10px] font-bold uppercase tracking-wide text-white/70 lg:text-[8px]">
                {group.section}
              </p>
            )}
            <div className="space-y-1">
              {group.items.map(({ href, label, icon: Icon, badge }) => {
                const active = isActive(href);
                return (
                  <Link
                    key={href}
                    href={href}
                    onClick={onNavigate}
                    title={label}
                    aria-current={active ? "page" : undefined}
                    className={cn(
                      "flex h-11 items-center gap-3 rounded-md px-2.5 text-[13px] transition-colors lg:h-8 lg:gap-2 lg:px-2 lg:text-[11.5px]",
                      active
                        ? "bg-[#114a82] font-bold text-white"
                        : "text-white/90 hover:bg-white/10 hover:text-white"
                    )}
                  >
                    <Icon className="h-4 w-4 shrink-0 lg:h-3 lg:w-3" aria-hidden="true" />
                    <span className="min-w-0 whitespace-nowrap">{label}</span>
                    {typeof badge === "number" && badge > 0 && (
                      <span className="ml-auto grid h-4 min-w-4 place-items-center rounded-full bg-white px-1 text-[9px] font-bold text-[#1768ad] lg:h-3.5 lg:min-w-3.5 lg:text-[8px]">
                        {badge > 99 ? "99+" : badge}
                      </span>
                    )}
                  </Link>
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </nav>
  );
}

/* Kerangka panel mengikuti referensi PDF: header putih, sidebar biru,
   kanvas biru muda, dan isi panel yang padat.
   Tiga mode tata letak:
   - < 1024px (layar sentuh) : drawer penuh lewat tombol menu, target sentuh besar
   - 768–1023 (mouse/trackpad): rail ikon 68px, label lewat tooltip
   - ≥ 1024                   : sidebar penuh 200px
   Aturan pointer ada di globals.css (.panel-rail / .panel-burger) supaya tablet
   sentuh tidak terjebak pada rail ikon tanpa label. */
export function PanelLayout({
  children,
  role,
  links,
}: {
  children: React.ReactNode;
  role: string;
  links: PanelLink[];
}) {
  const pathname = usePathname();
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const logout = useLogout();
  const [open, setOpen] = useState(false);

  const roleInitial = role.charAt(0).toUpperCase();
  const isDashboardHome = links[0]?.href;

  function signOut() {
    logout();
    toast.success("Logout berhasil");
    router.push("/login");
  }

  function isActive(href: string) {
    if (pathname === href) return true;
    if (href === isDashboardHome) return false;
    return pathname.startsWith(`${href}/`);
  }

  // Tutup drawer setiap kali rute berubah.
  useEffect(() => {
    setOpen(false);
  }, [pathname]);

  // Drawer hanya untuk < 1024px: tutup saat layar melebar agar scroll body
  // tidak tertinggal terkunci.
  useEffect(() => {
    const media = window.matchMedia("(min-width: 1024px)");
    const onChange = () => {
      if (media.matches) setOpen(false);
    };
    onChange();
    media.addEventListener("change", onChange);
    return () => media.removeEventListener("change", onChange);
  }, []);

  // Kunci scroll body + tutup dengan Escape selama drawer terbuka.
  useEffect(() => {
    if (!open) return;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  // Urutkan ulang link berdasarkan `section` agar sub-item berkelompok.
  const sections = links.reduce<NavSection[]>((acc, link) => {
    const last = acc[acc.length - 1];
    if (last && last.section === link.section) last.items.push(link);
    else acc.push({ section: link.section, items: [link] });
    return acc;
  }, []);

  const ariaLabel = `Navigasi ${role}`;

  return (
    <div className="min-h-screen bg-[#eef7fd]">
      <header className="sticky top-0 z-40 flex h-14 items-center justify-between gap-2 border-b border-[#c9deee] bg-white px-3 sm:h-16 sm:gap-3 sm:px-5 lg:px-6">
        <div className="flex min-w-0 items-center gap-1.5 sm:gap-2">
          <button
            className="panel-burger -ml-1 h-10 w-10 shrink-0 place-items-center rounded-md text-[#29445e] hover:bg-[#f2f7fb]"
            onClick={() => setOpen(true)}
            aria-label="Buka menu navigasi"
            aria-expanded={open}
            aria-controls="panel-mobile-nav"
          >
            <Menu className="h-5 w-5" />
          </button>
          <Link
            href="/"
            className="flex min-w-0 items-center gap-2.5"
            aria-label="Kembali ke beranda"
          >
            <Brand role={role} />
          </Link>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <Link
            href="/"
            className="hidden h-8 items-center gap-1.5 rounded-md border border-[#c9deee] px-3 text-[11px] font-bold text-[#14548f] hover:bg-[#f2f7fb] sm:inline-flex"
          >
            <Home className="h-3.5 w-3.5" /> Beranda
          </Link>
          <span className="grid h-7 w-7 place-items-center rounded-full bg-[#1768ad] text-[11px] font-bold text-white">
            {roleInitial}
          </span>
          <span className="hidden max-w-[10rem] truncate text-[13px] font-bold text-[#17324d] sm:block">
            {user?.name ?? role}
          </span>
        </div>
      </header>

      <div className="flex min-h-[calc(100dvh-56px)] sm:min-h-[calc(100dvh-64px)]">
        {/* Rail ikon — 768–1023px dengan mouse/trackpad */}
        <aside className="panel-rail sticky top-16 h-[calc(100dvh-64px)] w-[68px] shrink-0 flex-col self-start bg-[#1760a7]">
          <NavGroups
            sections={sections}
            isActive={isActive}
            onNavigate={() => setOpen(false)}
            ariaLabel={ariaLabel}
            variant="rail"
          />
          <button
            onClick={signOut}
            title="Logout"
            aria-label="Logout"
            className="mx-2 mb-3 grid h-11 shrink-0 place-items-center rounded-md border border-white/60 text-white hover:bg-white/10"
          >
            <LogOut className="h-4 w-4" />
          </button>
        </aside>

        {/* Sidebar penuh — desktop (≥ 1024px) */}
        <aside className="sticky top-16 hidden h-[calc(100dvh-64px)] w-[200px] shrink-0 flex-col self-start bg-[#1760a7] lg:flex">
          <NavGroups
            sections={sections}
            isActive={isActive}
            onNavigate={() => setOpen(false)}
            ariaLabel={ariaLabel}
            variant="full"
          />
          <button
            onClick={signOut}
            className="mx-2.5 mb-3 flex h-9 shrink-0 items-center gap-2 rounded-md border border-white/60 px-2 text-left text-[11.5px] font-bold text-white hover:bg-white/10"
          >
            <LogOut className="h-3 w-3" /> Logout
          </button>
        </aside>

        {/* Drawer — ponsel & tablet sentuh (< 1024px) */}
        {open && (
          <div
            className="fixed inset-0 z-50 lg:hidden"
            role="dialog"
            aria-modal="true"
            aria-label={`Menu ${role}`}
          >
            <button
              type="button"
              className="panel-scrim absolute inset-0 bg-[#0b2942]/45"
              aria-label="Tutup menu"
              onClick={() => setOpen(false)}
            />
            <aside
              id="panel-mobile-nav"
              className="panel-drawer absolute inset-y-0 left-0 flex h-full w-[min(84vw,288px)] flex-col bg-[#1760a7] shadow-2xl sm:w-[320px]"
            >
              <div className="flex h-16 shrink-0 items-center justify-between gap-2 border-b border-white/15 px-4">
                <Brand role={role} variant="drawer" />
                <button
                  onClick={() => setOpen(false)}
                  aria-label="Tutup menu"
                  className="-mr-1 grid h-10 w-10 shrink-0 place-items-center rounded-md text-white hover:bg-white/10"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>
              <NavGroups
                sections={sections}
                isActive={isActive}
                onNavigate={() => setOpen(false)}
                ariaLabel={ariaLabel}
                variant="full"
              />
              <div className="shrink-0 border-t border-white/15 px-2.5 py-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]">
                <button
                  onClick={signOut}
                  className="flex h-11 w-full items-center gap-3 rounded-md border border-white/60 px-3 text-left text-[13px] font-bold text-white hover:bg-white/10"
                >
                  <LogOut className="h-4 w-4" /> Logout
                </button>
              </div>
            </aside>
          </div>
        )}

        <main className="min-w-0 flex-1 px-3 pb-[max(1rem,env(safe-area-inset-bottom))] pt-4 sm:px-5 sm:py-5 lg:px-6">
          {children}
        </main>
      </div>
    </div>
  );
}
