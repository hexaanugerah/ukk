"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { ChevronDown, Minus, Plus } from "lucide-react";
import { DatePicker } from "@/components/inputs/date-picker";
import { ticketService } from "@/services/ticket.api.service";
import { todayISO, addDaysISO, rupiah, calculateBookingTotals } from "@/lib/business";
import { BOOKING_CONFIG } from "@/config/app.config";
import type { Ticket, TicketCategory } from "@/types/api";

// Pilih tiket — hanya memilih tanggal & jumlah tiket. Daftar tiket dikelompokkan per kategori dan bisa dilipat. Semua jumlah dimulai dari 0.

const CATEGORY_ORDER: TicketCategory[] = [
  "ENTRANCE",
  "SUNRISE",
  "PACKAGE",
  "TRANSPORT",
  "ACTIVITY",
  "SERVICE",
];

const CATEGORY_LABELS: Record<TicketCategory, string> = {
  ENTRANCE: "Tiket Masuk",
  SUNRISE: "Sunrise",
  PACKAGE: "Paket Keluarga",
  TRANSPORT: "Transportasi",
  ACTIVITY: "Aktivitas & Wahana",
  SERVICE: "Fasilitas & Layanan",
};

function categoryLabel(category: string): string {
  return CATEGORY_LABELS[category as TicketCategory] ?? category;
}

export default function PilihTiketPage() {
  const router = useRouter();
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [loading, setLoading] = useState(true);
  const [date, setDate] = useState("");
  const [qty, setQty] = useState<Record<string, number>>({});
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({});

  useEffect(() => {
    ticketService
      .getActive()
      .then((ts) => {
        setTickets(ts);
        // Semua tiket mulai dari 0 — pengunjung yang menentukan jumlahnya.
        setQty(Object.fromEntries(ts.map((t) => [t.id, 0])));
        const firstCategory =
          CATEGORY_ORDER.find((c) => ts.some((t) => t.category === c)) ?? ts[0]?.category;
        if (firstCategory) setOpenGroups({ [firstCategory]: true });
        setLoading(false);
      })
      .catch(() => {
        setLoading(false);
      });
  }, []);

  function change(key: string, delta: number) {
    setQty((prev) => ({ ...prev, [key]: Math.max(0, Math.min(20, (prev[key] ?? 0) + delta)) }));
  }

  function toggleGroup(category: string) {
    setOpenGroups((prev) => ({ ...prev, [category]: !prev[category] }));
  }

  // Kelompokkan tiket per kategori, urut sesuai CATEGORY_ORDER.
  const groups = useMemo(() => {
    const byCategory = new Map<string, Ticket[]>();
    for (const t of tickets) {
      const list = byCategory.get(t.category) ?? [];
      list.push(t);
      byCategory.set(t.category, list);
    }
    const ordered = [
      ...CATEGORY_ORDER,
      ...[...byCategory.keys()].filter((c) => !CATEGORY_ORDER.includes(c as TicketCategory)),
    ];
    return ordered
      .filter((c) => byCategory.has(c))
      .map((c) => ({ category: c, items: byCategory.get(c)! }));
  }, [tickets]);

  const lines = useMemo(
    () => tickets.map((t) => ({ ...t, quantity: qty[t.id] ?? 0 })).filter((t) => t.quantity > 0),
    [tickets, qty]
  );

  // Total dihitung lewat satu utilitas terpusat.
  const totals = useMemo(
    () => calculateBookingTotals(lines.map((l) => ({ price: l.price, quantity: l.quantity }))),
    [lines]
  );

  const canOrder = totals.totalTickets > 0 && !!date;

  const minDate = addDaysISO(todayISO(), BOOKING_CONFIG.minDaysAhead);
  const maxDate = addDaysISO(todayISO(), BOOKING_CONFIG.maxDaysAhead);

  function startOrder() {
    if (!canOrder) return;
    const items = lines.map((l) => `${l.id}:${l.quantity}`).join(",");
    const params = new URLSearchParams({ date, items });
    router.push(`/booking?${params.toString()}`);
  }

  if (loading) {
    return (
      <div className="page-shell py-10">
        <div className="animate-pulse space-y-4 max-w-2xl">
          <div className="h-8 bg-gray-200 rounded w-3/4" />
          <div className="h-4 bg-gray-200 rounded w-1/2" />
          <div className="grid gap-4 lg:grid-cols-[1.6fr_1fr]">
            <div className="h-64 bg-gray-200 rounded" />
            <div className="h-64 bg-gray-200 rounded" />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="page-shell py-10">
      <div className="mb-8 max-w-2xl">
        <h1 className="font-heading text-3xl font-bold text-[#17324d]">Pilih Tiket</h1>
        <p className="mt-3 text-[15px] leading-relaxed text-[#698097]">
          “Temukan pilihan tiket untuk pengalaman berkunjung yang lebih mudah dan nyaman.”
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <div>
          <div className="rounded-md border border-[#d7e3ed] bg-white p-5">
            <h2 className="text-lg font-bold text-[#17324d]">Kawah Putih</h2>
            <p className="mt-1 text-sm text-[#698097]">
              Pilih tanggal kunjungan, lalu tambahkan tiket yang Anda butuhkan.
            </p>

            <div className="mt-5">
              <span className="mb-1.5 block text-xs font-semibold text-[#29445e]">Pilih tanggal kunjungan</span>
              <DatePicker
                id="visit-date"
                value={date}
                onChange={setDate}
                min={minDate}
                max={maxDate}
                placeholder="Pilih tanggal kunjungan"
              />
              <span className="mt-1 block text-[11px] text-[#94A3B8]">
                Pemesanan {BOOKING_CONFIG.minDaysAhead === 0 ? "hari ini" : `H-${BOOKING_CONFIG.minDaysAhead}`} hingga {Math.round(BOOKING_CONFIG.maxDaysAhead / 30)} bulan ke depan.
              </span>
            </div>
          </div>

          <div className="mt-4 space-y-3">
            {groups.map((group) => {
              const open = !!openGroups[group.category];
              const selectedInGroup = group.items.reduce((sum, t) => sum + (qty[t.id] ?? 0), 0);
              return (
                <div key={group.category} className="overflow-hidden rounded-md border border-[#d7e3ed] bg-white">
                  <button
                    type="button"
                    onClick={() => toggleGroup(group.category)}
                    aria-expanded={open}
                    className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left hover:bg-[#f7fbff]"
                  >
                    <span className="flex items-center gap-2">
                      <ChevronDown
                        className={`h-4 w-4 shrink-0 text-[#698097] transition-transform ${open ? "" : "-rotate-90"}`}
                        aria-hidden="true"
                      />
                      <span className="text-sm font-bold text-[#17324d]">{categoryLabel(group.category)}</span>
                      <span className="text-xs text-[#94A3B8]">({group.items.length})</span>
                    </span>
                    {selectedInGroup > 0 && (
                      <span className="rounded-full bg-[#1768ad] px-2 py-0.5 text-[11px] font-bold text-white">
                        {selectedInGroup} dipilih
                      </span>
                    )}
                  </button>

                  {open && (
                    <ul className="border-t border-[#e5edf3]">
                      {group.items.map((t) => (
                        <li
                          key={t.id}
                          className="flex items-center justify-between gap-3 border-b border-[#eef4f9] px-4 py-3 last:border-b-0"
                        >
                          <div className="min-w-0">
                            <p className="truncate text-sm font-semibold text-[#17324d]">{t.name}</p>
                            {t.description && (
                              <p className="truncate text-xs text-[#698097]">{t.description}</p>
                            )}
                            <p className="mt-0.5 text-xs font-semibold text-[#1768ad]">{rupiah(t.price)}</p>
                          </div>
                          <div className="flex shrink-0 items-center gap-2">
                            <button
                              onClick={() => change(t.id, -1)}
                              className="grid h-7 w-7 place-items-center rounded-md border border-[#d7e3ed] text-[#29445e] hover:bg-[#f2f7fb] disabled:opacity-40"
                              aria-label={`Kurangi ${t.name}`}
                              disabled={(qty[t.id] ?? 0) === 0}
                            >
                              <Minus className="h-3.5 w-3.5" />
                            </button>
                            <span className="w-7 text-center text-sm font-bold tabular-nums text-[#17324d]">
                              {qty[t.id] ?? 0}
                            </span>
                            <button
                              onClick={() => change(t.id, 1)}
                              className="grid h-7 w-7 place-items-center rounded-md border border-[#d7e3ed] text-[#29445e] hover:bg-[#f2f7fb]"
                              aria-label={`Tambah ${t.name}`}
                            >
                              <Plus className="h-3.5 w-3.5" />
                            </button>
                          </div>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        <aside className="h-fit rounded-md border border-[#d7e3ed] bg-white p-5 lg:sticky lg:top-24">
          <h3 className="text-sm font-bold text-[#17324d]">Informasi Harga</h3>

          <dl className="mt-3 space-y-2 border-b border-[#e5edf3] pb-3 text-sm">
            <div className="flex justify-between">
              <dt className="text-[#698097]">Tanggal</dt>
              <dd className="font-medium text-[#17324d]">{date || "Belum dipilih"}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-[#698097]">Jumlah tiket</dt>
              <dd className="font-medium tabular-nums text-[#17324d]">{totals.totalTickets}</dd>
            </div>
          </dl>

          <ul className="mt-3 space-y-2 text-sm">
            {lines.length === 0 && <li className="text-[#94A3B8]">Belum ada tiket dipilih.</li>}
            {lines.map((l) => (
              <li key={l.id} className="flex justify-between">
                <span className="text-[#698097]">
                  {l.name} × {l.quantity}
                </span>
                <span className="font-medium tabular-nums text-[#17324d]">{rupiah(l.price * l.quantity)}</span>
              </li>
            ))}
          </ul>

          <div className="mt-4 space-y-1.5 border-t border-[#e5edf3] pt-3 text-sm">
            <div className="flex justify-between">
              <span className="text-[#698097]">Subtotal</span>
              <span className="tabular-nums text-[#17324d]">{rupiah(totals.subtotal)}</span>
            </div>
            {totals.discount > 0 && (
              <div className="flex justify-between text-[#16876a]">
                <span>Diskon</span>
                <span className="tabular-nums">-{rupiah(totals.discount)}</span>
              </div>
            )}
            <div className="flex justify-between">
              <span className="text-[#698097]">Biaya Layanan</span>
              <span className="tabular-nums text-[#17324d]">{rupiah(totals.serviceFee)}</span>
            </div>
            {totals.tax > 0 && (
              <div className="flex justify-between">
                <span className="text-[#698097]">Pajak</span>
                <span className="tabular-nums text-[#17324d]">{rupiah(totals.tax)}</span>
              </div>
            )}
            <div className="flex justify-between border-t border-[#e5edf3] pt-2 font-bold">
              <span className="text-[#17324d]">Total</span>
              <span className="tabular-nums text-[#17324d]">{rupiah(totals.total)}</span>
            </div>
          </div>

          <button
            onClick={startOrder}
            disabled={!canOrder}
            className="mt-5 flex h-11 w-full items-center justify-center rounded-md bg-[#1768ad] text-sm font-semibold text-white hover:bg-[#14548f] disabled:cursor-not-allowed disabled:opacity-60"
          >
            Pesan
          </button>
          <p className="mt-2 text-center text-[11px] text-[#94A3B8]">
            {canOrder ? "Lanjut ke data pemesanan & pembayaran." : "Pilih tanggal & minimal 1 tiket untuk lanjut."}
          </p>
        </aside>
      </div>
    </div>
  );
}
