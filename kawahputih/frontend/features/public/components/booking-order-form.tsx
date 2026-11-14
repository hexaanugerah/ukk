"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { bookingService, BookingError } from "@/services/booking.api.service";
import { ticketService } from "@/services/ticket.api.service";
import { useBookingStore } from "@/store/booking.store";
import { calculateBookingTotals, rupiah } from "@/lib/business";
import { useAuthStore } from "@/store/auth.store";
import { PhoneInput } from "@/components/ui/phone-input";
import { validatePhoneFull } from "@/lib/phone";
import type { Ticket } from "@/types/api";

// Booking order form — langkah kedua alur pemesanan. Pemilihan tanggal & tiket dibawa lewat query string (?date=...&items=id:qty). Halaman ini hanya melengkapi data pemesan.

const visitorSchemaZ = {
  name: (v: string) => (v.trim().length >= 3 ? null : "Nama lengkap minimal 3 karakter"),
  email: (v: string) => (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v) ? null : "Format email tidak valid"),
  phone: (v: string) => {
    const r = validatePhoneFull(v, true);
    return r.valid ? null : r.message ?? "Nomor HP tidak valid";
  },
};

interface FieldErrors {
  name?: string;
  email?: string;
  phone?: string;
  items?: string;
}

/** Parsing ringan `id:qty,id:qty` dari query string. */
function parseItems(raw: string | null): { id: string; quantity: number }[] {
  if (!raw) return [];
  return raw
    .split(",")
    .map((part) => {
      const [id, q] = part.split(":");
      return { id: (id ?? "").trim(), quantity: Number(q) };
    })
    .filter((i) => i.id.length > 0 && Number.isFinite(i.quantity) && i.quantity > 0);
}

export function BookingOrderForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setDraft = useBookingStore((s) => s.setDraft);
  const user = useAuthStore((s) => s.user);

  const visitDate = searchParams.get("date") ?? "";
  const selectionItems = useMemo(() => parseItems(searchParams.get("items")), [searchParams]);

  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [note, setNote] = useState("");
  const [errors, setErrors] = useState<FieldErrors>({});

  // Prefill dari profil user yang login.
  useEffect(() => {
    if (user) {
      setName(user.name);
      setEmail(user.email);
      setPhone(user.phone);
    }
  }, [user]);

  // Muat tiket aktif untuk menerjemahkan id pilihan → nama & harga.
  useEffect(() => {
    let alive = true;
    ticketService
      .getActive()
      .then((ts) => {
        if (!alive) return;
        setTickets(ts);
        setLoading(false);
      })
      .catch(() => {
        if (!alive) return;
        setLoadError(true);
        setLoading(false);
      });
    return () => {
      alive = false;
    };
  }, []);

  // Terjemahkan pilihan dari query ke baris tiket lengkap.
  const lines = useMemo(
    () =>
      selectionItems
        .map((si) => {
          const t = tickets.find((x) => x.id === si.id);
          return t ? { ...t, quantity: si.quantity } : null;
        })
        .filter((l): l is Ticket & { quantity: number } => l !== null),
    [selectionItems, tickets]
  );

  const totals = useMemo(
    () => calculateBookingTotals(lines.map((l) => ({ price: l.price, quantity: l.quantity }))),
    [lines]
  );

  const hasValidSelection =
    selectionItems.length > 0 && visitDate !== "" && lines.length === selectionItems.length;

  function validate(): boolean {
    const errs: FieldErrors = {};
    const nameErr = visitorSchemaZ.name(name);
    const emailErr = visitorSchemaZ.email(email);
    const phoneErr = visitorSchemaZ.phone(phone);
    if (nameErr) errs.name = nameErr;
    if (emailErr) errs.email = emailErr;
    if (phoneErr) errs.phone = phoneErr;
    setErrors(errs);
    return Object.keys(errs).length === 0;
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!hasValidSelection) return;
    // Backend hanya mengizinkan pemesanan oleh pengguna yang sudah masuk.
    if (!user) {
      toast.error("Silakan masuk terlebih dahulu", {
        description: "Login untuk menyelesaikan pemesanan tiket.",
      });
      const back = `/booking?${searchParams.toString()}`;
      router.push(`/login?redirect=${encodeURIComponent(back)}`);
      return;
    }
    if (!validate()) return;

    setSubmitting(true);
    try {
      const response = await bookingService.create({
        visit_date: visitDate,
        items: lines.map((l) => ({ ticket_id: l.id, quantity: l.quantity })),
        visitor: {
          name,
          email,
          phone,
        },
        note: note || undefined,
      });
      const { booking } = response;
      setDraft({
        visitDate,
        items: lines.map((l) => ({ ticketId: l.id, ticketName: l.name, price: l.price, quantity: l.quantity })),
        visitorName: name,
        email,
        phone,
        note: note || undefined,
        totals,
      });
      toast.success("Pesanan dibuat", { description: `Kode booking ${booking.booking_code}. Lanjutkan ke pembayaran.` });
      router.push(`/payment?code=${booking.id}`);
    } catch (err: any) {
      const message = err instanceof BookingError ? err.message : (err.message || "Terjadi kesalahan saat membuat pesanan.");
      toast.error("Gagal membuat pesanan", { description: message });
      setErrors({ items: message });
    } finally {
      setSubmitting(false);
    }
  }

  const field =
    "h-10 w-full rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]";
  const errText = "mt-1 text-xs text-red-600";

  if (loading) {
    return (
      <div className="grid gap-4 lg:grid-cols-[1.6fr_1fr]">
        <div className="h-72 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />
        <div className="h-72 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />
      </div>
    );
  }

  if (loadError) {
    return (
      <div className="rounded-md border border-[#fdeaea] bg-[#fdeaea] p-6 text-center">
        <p className="text-sm text-[#c53030]">Terjadi kesalahan saat memuat data tiket.</p>
        <button
          onClick={() => {
            setLoadError(false);
            setLoading(true);
            ticketService
              .getActive()
              .then((ts) => {
                setTickets(ts);
                setLoading(false);
              })
              .catch(() => {
                setLoadError(true);
                setLoading(false);
              });
          }}
          className="mt-3 rounded-md bg-[#1768ad] px-4 py-2 text-sm font-semibold text-white hover:bg-[#14548f]"
        >
          Coba Lagi
        </button>
      </div>
    );
  }

  // Tidak ada pilihan tiket (mis. halaman dibuka langsung / di-refresh) →
  // arahkan kembali ke halaman Pilih Tiket.
  if (!hasValidSelection) {
    return (
      <div className="rounded-md border border-[#d7e3ed] bg-white p-8 text-center">
        <h2 className="text-base font-bold text-[#17324d]">Belum ada tiket yang dipilih</h2>
        <p className="mt-1 text-sm text-[#698097]">
          Pilih tiket terlebih dahulu, lalu lengkapi data pemesanan di sini.
        </p>
        <Link
          href="/packages"
          className="mt-4 inline-flex h-10 items-center rounded-md bg-[#1768ad] px-5 text-sm font-semibold text-white hover:bg-[#14548f]"
        >
          Pilih Tiket
        </Link>
      </div>
    );
  }

  return (
    <form onSubmit={onSubmit} className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
      <div>
        <section className="rounded-md border border-[#d7e3ed] bg-white p-5">
          <h2 className="text-base font-bold text-[#17324d]">Data Pemesan</h2>
          <p className="mt-1 text-xs text-[#698097]">Lengkapi data pemesan untuk melanjutkan ke pembayaran.</p>

          <div className="mt-4 space-y-4">
            <label className="block">
              <span className="mb-1.5 block text-xs font-semibold text-[#29445e]">Nama Lengkap*</span>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Sesuai KTP/paspor/SIM"
                className={field}
                aria-invalid={!!errors.name}
              />
              {errors.name && <span className={errText}>{errors.name}</span>}
            </label>

            <label className="block">
              <span className="mb-1.5 block text-xs font-semibold text-[#29445e]">No. Handphone*</span>
              <PhoneInput
                value={phone}
                onChange={setPhone}
                invalid={!!errors.phone}
                placeholder="81234567890"
                inputClassName={`${field}`}
                selectClassName="h-10 rounded-md border border-[#d7e3ed] bg-[#f7fbff] px-2 text-sm text-[#698097] outline-none focus:border-brand-600"
                ariaLabel="Nomor handphone"
              />
              {errors.phone && <span className={errText}>{errors.phone}</span>}
            </label>

            <label className="block">
              <span className="mb-1.5 block text-xs font-semibold text-[#29445e]">Email*</span>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="email@contoh.com"
                className={field}
                aria-invalid={!!errors.email}
              />
              {errors.email && <span className={errText}>{errors.email}</span>}
            </label>

            <label className="block">
              <span className="mb-1.5 block text-xs font-semibold text-[#29445e]">Catatan (opsional)</span>
              <textarea
                value={note}
                onChange={(e) => setNote(e.target.value)}
                rows={2}
                className="w-full rounded-md border border-[#d7e3ed] bg-white px-3 py-2 text-sm text-[#17324d] outline-none focus:border-[#1768ad]"
              />
            </label>
          </div>
        </section>
      </div>

      <aside className="h-fit rounded-md border border-[#d7e3ed] bg-white p-5 lg:sticky lg:top-24">
        <h3 className="text-sm font-bold text-[#17324d]">Rincian Pesanan</h3>

        <dl className="mt-3 space-y-2 border-b border-[#e5edf3] pb-3 text-sm">
          <div className="flex justify-between">
            <dt className="text-[#698097]">Tanggal Kunjungan</dt>
            <dd className="font-medium text-[#17324d]">{visitDate}</dd>
          </div>
        </dl>

        <ul className="mt-3 space-y-2 text-sm">
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
            <span className="text-[#17324d]">Total ({totals.totalTickets} tiket)</span>
            <span className="tabular-nums text-[#17324d]">{rupiah(totals.total)}</span>
          </div>
        </div>

        {errors.items && <p className="mt-3 text-xs text-red-600">{errors.items}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="mt-5 flex h-11 w-full items-center justify-center gap-2 rounded-md bg-[#1768ad] text-sm font-semibold text-white hover:bg-[#14548f] disabled:opacity-60"
        >
          {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
          {submitting ? "Memproses..." : "Lanjut ke Pembayaran"}
        </button>
        <Link
          href="/packages"
          className="mt-2 block text-center text-[11px] font-semibold text-[#1768ad] hover:underline"
        >
          Ubah pilihan tiket
        </Link>
        <p className="mt-1 text-center text-[11px] text-[#94A3B8]">Pembayaran diproses via Midtrans (QRIS, VA, e-wallet).</p>
      </aside>
    </form>
  );
}
