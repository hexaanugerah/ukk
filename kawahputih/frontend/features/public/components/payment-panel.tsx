"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Check, CheckCircle2, Copy, CreditCard, Loader2, Store, Timer, Wallet, XCircle } from "lucide-react";
import { toast } from "sonner";
import { bookingService, paymentService, BookingError } from "@/services/api";
import { rupiah } from "@/lib/business";
import {
  PAYMENT_METHOD_GROUPS,
  findPaymentMethod,
  paymentMethodLabel,
  type PaymentMethodCode,
  type PaymentMethodOption,
} from "@/config/payment.config";
import type { Booking, Payment } from "@/types/api";

type PaymentOutcome = "SUCCESS" | "FAILED" | "EXPIRED";

type Phase = "LOADING" | "WAITING" | "PROCESSING" | "SUCCESS" | "FAILED" | "EXPIRED" | "NOT_FOUND";

/** Prefix Virtual Account per bank — angka di belakangnya diturunkan dari kode booking. */
const VA_PREFIX: Record<string, string> = {
  BCA_VA: "8808",
  BNI_VA: "8810",
  MANDIRI_VA: "8888",
  BRI_VA: "8889",
};

/** Angka deterministik dari sebuah seed (kode booking) — untuk nomor VA/kode bayar. */
function deriveDigits(seed: string, len: number): string {
  let h = 2166136261;
  for (let i = 0; i < seed.length; i++) {
    h = Math.imul(h ^ seed.charCodeAt(i), 16777619) >>> 0;
  }
  let out = "";
  let n = h || 1;
  while (out.length < len) {
    n = (Math.imul(n, 1103515245) + 12345) >>> 0;
    out += String(n % 10);
  }
  return out.slice(0, len);
}

function virtualAccountNumber(method: string, bookingID: string): string {
  const prefix = VA_PREFIX[method] ?? "8808";
  return prefix + deriveDigits(`${method}:${bookingID}`, 12);
}

function retailPaymentCode(bookingID: string): string {
  return "KP" + deriveDigits(`retail:${bookingID}`, 10);
}

function CopyButton({ value, label }: { value: string; label: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      type="button"
      onClick={() => {
        navigator.clipboard?.writeText(value).then(
          () => {
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
          },
          () => toast.error("Gagal menyalin")
        );
      }}
      className="inline-flex items-center gap-1 rounded-md border border-[#d7e3ed] px-2 py-1 text-[11px] font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
      aria-label={label}
    >
      {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
      {copied ? "Tersalin" : "Salin"}
    </button>
  );
}

/** Instruksi pembayaran sesuai jenis metode yang dipilih. */
function Instruction({
  option,
  bookingID,
  total,
}: {
  option: PaymentMethodOption;
  bookingID: string;
  total: number;
}) {
  if (option.kind === "qris") {
    return (
      <div className="mt-3 text-center">
        <div className="mx-auto grid h-44 w-44 place-items-center rounded-md border border-[#e5edf3] bg-[#f7fbff] p-2">
          <div
            className="h-full w-full opacity-90"
            style={{
              backgroundImage: "repeating-conic-gradient(#17324d 0% 25%, #ffffff 0% 50%)",
              backgroundSize: "12px 12px",
            }}
            role="img"
            aria-label="Kode QRIS pembayaran"
          />
        </div>
        <p className="mt-2 text-xs text-[#698097]">Pindai kode ini dengan aplikasi pembayaran Anda.</p>
      </div>
    );
  }

  if (option.kind === "va") {
    const va = virtualAccountNumber(option.code, bookingID);
    return (
      <div className="mt-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-4">
        <p className="text-[11px] font-bold uppercase tracking-wide text-[#698097]">Nomor Virtual Account</p>
        <div className="mt-2 flex items-center justify-between gap-3">
          <span className="font-mono text-lg font-bold tracking-wider text-[#17324d]">{va}</span>
          <CopyButton value={va} label="Salin nomor virtual account" />
        </div>
        <p className="mt-2 text-xs text-[#698097]">
          Masukkan nomor di atas pada menu <span className="font-semibold">Transfer → Virtual Account</span> di ATM,
          mobile banking, atau internet banking. Nominal: <span className="font-semibold">{rupiah(total)}</span>.
        </p>
      </div>
    );
  }

  if (option.kind === "ewallet") {
    return (
      <div className="mt-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-4">
        <div className="flex items-center gap-2">
          <Wallet className="h-4 w-4 text-[#1768ad]" />
          <p className="text-sm font-bold text-[#17324d]">Bayar dengan {option.label}</p>
        </div>
        <p className="mt-2 text-xs text-[#698097]">
          Buka aplikasi <span className="font-semibold">{option.label}</span>, pilih menu{" "}
          <span className="font-semibold">Bayar / Scan</span>, lalu konfirmasi pembayaran sebesar{" "}
          <span className="font-semibold">{rupiah(total)}</span>.
        </p>
        <p className="mt-2 text-xs text-[#94A3B8]">
          Nomor referensi: <span className="font-mono">{bookingID}</span>
        </p>
      </div>
    );
  }

  if (option.kind === "retail") {
    const code = retailPaymentCode(bookingID);
    return (
      <div className="mt-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-4">
        <div className="flex items-center gap-2">
          <Store className="h-4 w-4 text-[#1768ad]" />
          <p className="text-sm font-bold text-[#17324d]">Bayar tunai di {option.label}</p>
        </div>
        <p className="mt-2 text-[11px] font-bold uppercase tracking-wide text-[#698097]">Kode pembayaran</p>
        <div className="mt-1 flex items-center justify-between gap-3">
          <span className="font-mono text-lg font-bold tracking-wider text-[#17324d]">{code}</span>
          <CopyButton value={code} label="Salin kode pembayaran" />
        </div>
        <p className="mt-2 text-xs text-[#698097]">
          Tunjukkan kode ini ke kasir {option.label} dan bayar sebesar{" "}
          <span className="font-semibold">{rupiah(total)}</span>. Simpan struk sebagai bukti.
        </p>
      </div>
    );
  }

  // Kartu kredit / debit
  return (
    <div className="mt-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-4">
      <div className="flex items-center gap-2">
        <CreditCard className="h-4 w-4 text-[#1768ad]" />
        <p className="text-sm font-bold text-[#17324d]">Detail Kartu</p>
      </div>
      <div className="mt-3 space-y-2">
        <input
          disabled
          placeholder="1234 5678 9012 3456"
          className="h-10 w-full rounded-md border border-[#d7e3ed] bg-white px-3 font-mono text-sm text-[#17324d] disabled:opacity-70"
          aria-label="Nomor kartu"
        />
        <div className="grid grid-cols-2 gap-2">
          <input
            disabled
            placeholder="MM/YY"
            className="h-10 w-full rounded-md border border-[#d7e3ed] bg-white px-3 font-mono text-sm text-[#17324d] disabled:opacity-70"
            aria-label="Masa berlaku kartu"
          />
          <input
            disabled
            placeholder="CVV"
            className="h-10 w-full rounded-md border border-[#d7e3ed] bg-white px-3 font-mono text-sm text-[#17324d] disabled:opacity-70"
            aria-label="CVV kartu"
          />
        </div>
      </div>
      <p className="mt-2 text-xs text-[#94A3B8]">
        Simulasi — data kartu tidak dikirim ke mana pun.
      </p>
    </div>
  );
}

export function PaymentPanel({ code }: { code?: string }) {
  const router = useRouter();
  const [phase, setPhase] = useState<Phase>("LOADING");
  const [booking, setBooking] = useState<Booking | null>(null);
  const [payment, setPayment] = useState<Payment | null>(null);
  const [countdown, setCountdown] = useState<number>(0);
  const [method, setMethod] = useState<PaymentMethodCode>("QRIS");

  // Muat booking + payment dari backend API.
  useEffect(() => {
    let alive = true;
    (async () => {
      if (!code) {
        setPhase("NOT_FOUND");
        return;
      }
      try {
        const b = await bookingService.getById(code);
        const p = await paymentService.getByBookingId(code);
        if (!alive) return;
        setBooking(b);
        setPayment(p);
        if (p?.method && findPaymentMethod(p.method)) {
          setMethod(p.method as PaymentMethodCode);
        }
        setPhase(
          b?.status === "PENDING" ? "WAITING" :
          b?.status === "PAID" || b?.status === "CONFIRMED" || b?.status === "USED" ? "SUCCESS" : "FAILED"
        );
      } catch {
        if (alive) setPhase("NOT_FOUND");
      }
    })();
    return () => {
      alive = false;
    };
  }, [code]);

  // Countdown kadaluarsa pembayaran.
  useEffect(() => {
    if (!payment || phase !== "WAITING") return;
    const tick = () => {
      const left = Math.max(0, Math.floor((new Date(payment.expired_at).getTime() - Date.now()) / 1000));
      setCountdown(left);
      if (left === 0) {
        setPhase("EXPIRED");
        paymentService.expirePayment(payment.booking_id);
      }
    };
    tick();
    const timer = setInterval(tick, 1000);
    return () => clearInterval(timer);
  }, [payment, phase]);

  const selected = useMemo(() => findPaymentMethod(method), [method]);

  async function pay(outcome: PaymentOutcome) {
    if (!booking) return;
    setPhase("PROCESSING");
    try {
      const res = await paymentService.processPayment({ booking_id: booking.id, method, forced: outcome });
      if (res.outcome === "SUCCESS") {
        setPhase("SUCCESS");
        setBooking(res.booking);
        setPayment(res.payment);
        toast.success("Pembayaran berhasil!", {
          description: `E-tiket untuk ${res.booking.booking_code} telah diterbitkan.`,
        });
        setTimeout(() => router.push(`/e-ticket?code=${res.booking.id}&new=1`), 1200);
      } else if (res.outcome === "FAILED") {
        setPhase("FAILED");
        setPayment(res.payment);
        toast.error("Pembayaran gagal", { description: "Pembayaran gagal. Silakan coba kembali." });
      } else {
        setPhase("EXPIRED");
        toast.warning("Pembayaran kadaluarsa", { description: "Waktu pembayaran telah berakhir." });
      }
    } catch (err: any) {
      const message = err instanceof BookingError ? err.message : err.message || "Terjadi kesalahan saat memproses pembayaran.";
      toast.error("Pembayaran gagal", { description: message });
      setPhase("WAITING");
    }
  }

  const mm = String(Math.floor(countdown / 60)).padStart(2, "0");
  const ss = String(countdown % 60).padStart(2, "0");

  if (phase === "LOADING") {
    return <div className="h-80 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />;
  }

  if (phase === "NOT_FOUND" || !booking) {
    return (
      <div className="rounded-md border border-[#d7e3ed] bg-white p-8 text-center">
        <XCircle className="mx-auto h-10 w-10 text-[#c53030]" />
        <h2 className="mt-3 text-base font-bold text-[#17324d]">Pesanan tidak ditemukan</h2>
        <p className="mt-1 text-sm text-[#698097]">Kode booking tidak valid atau pesanan sudah diproses.</p>
        <Link href="/packages" className="mt-4 inline-flex h-10 items-center rounded-md bg-[#1768ad] px-5 text-sm font-semibold text-white hover:bg-[#14548f]">
          Buat Pesanan Baru
        </Link>
      </div>
    );
  }

  if (phase === "SUCCESS") {
    return (
      <div className="rounded-md border border-[#16876a] bg-[#f0fbf7] p-8 text-center">
        <CheckCircle2 className="mx-auto h-12 w-12 text-[#16876a]" />
        <h2 className="mt-3 text-lg font-bold text-[#17324d]">Pembayaran Berhasil!</h2>
        <p className="mt-1 text-sm text-[#698097]">
          Kode booking <span className="font-bold text-[#17324d]">{booking.booking_code}</span> telah dikonfirmasi
          {payment?.method ? ` via ${paymentMethodLabel(payment.method)}` : ""}. E-tiket Anda sedang dibuka...
        </p>
        <button
          onClick={() => router.push(`/e-ticket?code=${booking.id}&new=1`)}
          className="mt-4 inline-flex h-10 items-center rounded-md bg-[#1768ad] px-5 text-sm font-semibold text-white hover:bg-[#14548f]"
        >
          Lihat E-Tiket
        </button>
      </div>
    );
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
      <div className="space-y-4">
        <section className="rounded-md border border-[#d7e3ed] bg-white p-5">
          <h2 className="text-base font-bold text-[#17324d]">Metode Pembayaran</h2>
          <p className="mt-1 text-sm text-[#698097]">
            {phase === "WAITING"
              ? "Pilih salah satu metode pembayaran yang tersedia."
              : `Metode: ${paymentMethodLabel(method)}`}
          </p>

          {phase === "WAITING" && (
            <div className="mt-4 space-y-4">
              {PAYMENT_METHOD_GROUPS.map((group) => (
                <div key={group.key}>
                  <p className="text-[11px] font-bold uppercase tracking-wide text-[#698097]">{group.label}</p>
                  <div className="mt-2 grid gap-2 sm:grid-cols-2">
                    {group.options.map((opt) => {
                      const active = opt.code === method;
                      return (
                        <button
                          key={opt.code}
                          type="button"
                          onClick={() => setMethod(opt.code)}
                          aria-pressed={active}
                          className={`flex items-start gap-2 rounded-md border px-3 py-2.5 text-left transition-colors ${
                            active
                              ? "border-[#1768ad] bg-[#f2f8fe]"
                              : "border-[#d7e3ed] bg-white hover:bg-[#f7fbff]"
                          }`}
                        >
                          <span
                            className={`mt-0.5 grid h-4 w-4 shrink-0 place-items-center rounded-full border ${
                              active ? "border-[#1768ad] bg-[#1768ad]" : "border-[#c9deee]"
                            }`}
                          >
                            {active && <Check className="h-2.5 w-2.5 text-white" strokeWidth={3} />}
                          </span>
                          <span className="min-w-0">
                            <span className="block truncate text-sm font-semibold text-[#17324d]">{opt.label}</span>
                            <span className="block text-[11px] text-[#698097]">{opt.hint}</span>
                          </span>
                        </button>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="rounded-md border border-[#d7e3ed] bg-white p-5">
          <h2 className="text-base font-bold text-[#17324d]">Instruksi Pembayaran</h2>
          <p className="mt-1 text-sm text-[#698097]">
            {selected ? selected.label : paymentMethodLabel(method)}
          </p>

          {phase === "WAITING" && selected && (
            <>
              <Instruction option={selected} bookingID={booking.id} total={booking.total} />

              <p className="mt-3 flex items-center justify-center gap-1.5 text-sm font-semibold text-[#b7791f]">
                <Timer className="h-4 w-4" /> Bayar dalam {mm}:{ss}
              </p>

              <div className="mt-4 rounded-md border border-dashed border-[#bcd8ee] bg-[#f7fbff] p-3">
                <p className="text-[11px] font-bold uppercase tracking-wide text-[#698097]">Simulasi hasil pembayaran</p>
                <div className="mt-2 flex flex-wrap gap-2">
                  <button
                    onClick={() => pay("SUCCESS")}
                    className="rounded-md bg-[#16876a] px-3 py-1.5 text-xs font-semibold text-white hover:brightness-95"
                  >
                    Saya Sudah Bayar
                  </button>
                  <button
                    onClick={() => pay("FAILED")}
                    className="rounded-md bg-[#c53030] px-3 py-1.5 text-xs font-semibold text-white hover:brightness-95"
                  >
                    Simulasi Gagal
                  </button>
                </div>
                <p className="mt-1.5 text-[11px] text-[#94A3B8]">
                  Tombol testing — pada produksi ini digantikan status webhook gateway.
                </p>
              </div>
            </>
          )}

          {phase === "PROCESSING" && (
            <div className="grid h-40 w-full place-items-center">
              <div className="text-center">
                <Loader2 className="mx-auto h-8 w-8 animate-spin text-[#1768ad]" />
                <p className="mt-2 text-sm font-semibold text-[#29445e]">Memproses pembayaran...</p>
                <p className="text-xs text-[#698097]">Mohon tunggu, jangan tutup halaman ini.</p>
              </div>
            </div>
          )}

          {phase === "FAILED" && (
            <div className="mt-4 rounded-md bg-[#fdeaea] p-4 text-center">
              <p className="text-sm font-semibold text-[#c53030]">Pembayaran gagal diproses.</p>
              <button
                onClick={() => pay("SUCCESS")}
                className="mt-2 rounded-md bg-[#1768ad] px-4 py-2 text-sm font-semibold text-white hover:bg-[#14548f]"
              >
                Coba Bayar Lagi
              </button>
            </div>
          )}

          {phase === "EXPIRED" && (
            <div className="mt-4 rounded-md bg-[#fdeaea] p-4 text-center">
              <p className="text-sm font-semibold text-[#c53030]">Pembayaran kadaluarsa.</p>
              <button
                onClick={() => router.push("/packages")}
                className="mt-2 rounded-md bg-[#1768ad] px-4 py-2 text-sm font-semibold text-white hover:bg-[#14548f]"
              >
                Buat Pesanan Baru
              </button>
            </div>
          )}
        </section>
      </div>

      <aside className="h-fit rounded-md border border-[#d7e3ed] bg-white p-5 lg:sticky lg:top-24">
        <h3 className="text-sm font-bold text-[#17324d]">Ringkasan Pesanan</h3>
        <dl className="mt-3 space-y-2 text-sm">
          <div className="flex justify-between">
            <dt className="text-[#698097]">Kode Booking</dt>
            <dd className="font-medium text-[#17324d]">{booking.id}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-[#698097]">Nama</dt>
            <dd className="font-medium text-[#17324d]">{booking.visitor_name}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-[#698097]">Tanggal Kunjungan</dt>
            <dd className="font-medium text-[#17324d]">{booking.visit_date}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-[#698097]">Metode</dt>
            <dd className="font-medium text-[#17324d]">{paymentMethodLabel(method)}</dd>
          </div>
        </dl>
        <ul className="mt-3 space-y-2 border-t border-[#e5edf3] pt-3 text-sm">
          {booking.items.map((v) => (
            <li key={v.ticket_id} className="flex justify-between">
              <span className="text-[#698097]">
                {v.ticket_name} × {v.quantity}
              </span>
              <span className="font-medium tabular-nums text-[#17324d]">{rupiah(v.price * v.quantity)}</span>
            </li>
          ))}
        </ul>
        <div className="mt-3 space-y-1.5 border-t border-[#e5edf3] pt-3 text-sm">
          <div className="flex justify-between">
            <span className="text-[#698097]">Subtotal</span>
            <span className="tabular-nums text-[#17324d]">{rupiah(booking.subtotal)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-[#698097]">Biaya Layanan</span>
            <span className="tabular-nums text-[#17324d]">{rupiah(booking.service_fee)}</span>
          </div>
          <div className="flex justify-between font-bold">
            <span className="text-[#17324d]">Total</span>
            <span className="tabular-nums text-[#17324d]">{rupiah(booking.total)}</span>
          </div>
        </div>
        <Link href="/dashboard/booking" className="mt-3 block text-center text-xs text-[#698097] hover:underline">
          Lihat semua booking saya
        </Link>
      </aside>
    </div>
  );
}
