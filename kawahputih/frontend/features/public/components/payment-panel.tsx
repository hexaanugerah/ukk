"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import QRCode from "react-qr-code";
import {
  ArrowRight,
  Check,
  CheckCircle2,
  Copy,
  CreditCard,
  ImageUp,
  Loader2,
  Store,
  Timer,
  Wallet,
  XCircle,
} from "lucide-react";
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
import type { Booking, Payment, PaymentDestination } from "@/types/api";

type PaymentOutcome = "SUCCESS" | "FAILED" | "EXPIRED";

type Phase = "LOADING" | "WAITING" | "PROCESSING" | "SUCCESS" | "FAILED" | "EXPIRED" | "NOT_FOUND";

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
  payment,
  destination,
}: {
  option: PaymentMethodOption;
  bookingID: string;
  total: number;
  payment: Payment | null;
  destination?: PaymentDestination;
}) {
  // QRIS — QR asli dari provider (qris_string Midtrans). Bila provider belum
  // mengembalikan QR, tampilkan pesan dan sarankan metode lain (tanpa QR palsu).
  if (option.kind === "qris") {
    if (payment?.qris_string) {
      return (
        <div className="mt-3 text-center">
          <div className="mx-auto grid h-52 w-52 place-items-center rounded-md border border-[#e5edf3] bg-white p-3">
            <QRCode value={payment.qris_string} size={176} aria-label="Kode QRIS pembayaran" />
          </div>
          <p className="mt-2 text-xs text-[#698097]">Pindai kode ini dengan aplikasi pembayaran Anda.</p>
        </div>
      );
    }
    return (
      <div className="mt-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-4 text-center">
        <p className="text-sm font-semibold text-[#b7791f]">Kode QRIS belum tersedia.</p>
        <p className="mt-1 text-xs text-[#698097]">
          Pembayaran QRIS belum dapat diproses saat ini. Silakan pilih metode pembayaran lain.
        </p>
        <p className="mt-2 text-xs text-[#94A3B8]">
          Nomor referensi: <span className="font-mono">{payment?.reference ?? bookingID}</span>
        </p>
      </div>
    );
  }

  // Transfer bank / Virtual Account — nomor tujuan diatur oleh admin.
  if (option.kind === "va") {
    return (
      <div className="mt-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-4">
        <p className="text-[11px] font-bold uppercase tracking-wide text-[#698097]">Tujuan Transfer</p>
        {destination ? (
          <>
            <div className="mt-2 flex items-center justify-between gap-3">
              <span className="font-mono text-lg font-bold tracking-wider text-[#17324d]">{destination.account_number}</span>
              <CopyButton value={destination.account_number} label="Salin nomor rekening" />
            </div>
            <p className="mt-1 text-xs text-[#698097]">
              a.n. <span className="font-semibold text-[#17324d]">{destination.account_name || destination.provider_name}</span>
            </p>
            <p className="mt-2 text-xs text-[#698097]">
              Transfer tepat sebesar <span className="font-semibold">{rupiah(total)}</span>, lalu pembayaran akan
              dikonfirmasi oleh petugas/admin.
            </p>
          </>
        ) : (
          <p className="mt-2 text-xs text-[#b7791f]">
            Tujuan transfer untuk metode ini belum diatur oleh admin. Silakan pilih metode lain.
          </p>
        )}
        <p className="mt-2 text-xs text-[#94A3B8]">
          Nomor referensi: <span className="font-mono">{payment?.reference ?? bookingID}</span>
        </p>
      </div>
    );
  }

  // E-Wallet — bila tersedia, lanjut ke halaman pembayaran Midtrans;
  // nomor tujuan manual dari admin ditampilkan sebagai alternatif.
  if (option.kind === "ewallet") {
    return (
      <div className="mt-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-4">
        <div className="flex items-center gap-2">
          <Wallet className="h-4 w-4 text-[#1768ad]" />
          <p className="text-sm font-bold text-[#17324d]">Bayar dengan {option.label}</p>
        </div>
        {payment?.midtrans_redirect_url && (
          <a
            href={payment.midtrans_redirect_url}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-3 inline-flex h-9 items-center rounded-md bg-[#1768ad] px-4 text-xs font-semibold text-white hover:bg-[#14548f]"
          >
            Lanjut ke Halaman Pembayaran
          </a>
        )}
        {destination ? (
          <div className="mt-3">
            <p className="text-[11px] font-bold uppercase tracking-wide text-[#698097]">
              Atau transfer ke {destination.provider_name}
            </p>
            <div className="mt-1 flex items-center justify-between gap-3">
              <span className="font-mono text-base font-bold tracking-wider text-[#17324d]">{destination.account_number}</span>
              <CopyButton value={destination.account_number} label="Salin nomor tujuan" />
            </div>
            <p className="mt-1 text-xs text-[#698097]">
              a.n. <span className="font-semibold text-[#17324d]">{destination.account_name || destination.provider_name}</span> • nominal {rupiah(total)}
            </p>
          </div>
        ) : (
          !payment?.midtrans_redirect_url && (
            <p className="mt-2 text-xs text-[#b7791f]">
              Tujuan pembayaran {option.label} belum diatur oleh admin. Silakan pilih metode lain.
            </p>
          )
        )}
        <p className="mt-2 text-xs text-[#94A3B8]">
          Nomor referensi: <span className="font-mono">{payment?.reference ?? bookingID}</span>
        </p>
      </div>
    );
  }

  // Gerai retail — kode/aturan pembayaran ditentukan admin.
  if (option.kind === "retail") {
    return (
      <div className="mt-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-4">
        <div className="flex items-center gap-2">
          <Store className="h-4 w-4 text-[#1768ad]" />
          <p className="text-sm font-bold text-[#17324d]">Bayar tunai di {option.label}</p>
        </div>
        {destination ? (
          <>
            <p className="mt-2 text-[11px] font-bold uppercase tracking-wide text-[#698097]">Nomor tujuan</p>
            <div className="mt-1 flex items-center justify-between gap-3">
              <span className="font-mono text-lg font-bold tracking-wider text-[#17324d]">{destination.account_number}</span>
              <CopyButton value={destination.account_number} label="Salin nomor tujuan" />
            </div>
            <p className="mt-1 text-xs text-[#698097]">
              a.n. <span className="font-semibold text-[#17324d]">{destination.account_name || destination.provider_name}</span>
            </p>
            <p className="mt-2 text-xs text-[#698097]">
              Sebutkan referensi berikut ke kasir {option.label} dan bayar sebesar{" "}
              <span className="font-semibold">{rupiah(total)}</span>.
            </p>
          </>
        ) : (
          <p className="mt-2 text-xs text-[#b7791f]">
            Pembayaran {option.label} belum diaktifkan oleh admin. Silakan pilih metode lain.
          </p>
        )}
        <p className="mt-2 text-xs text-[#94A3B8]">
          Nomor referensi: <span className="font-mono">{payment?.reference ?? bookingID}</span>
        </p>
      </div>
    );
  }

  // Kartu kredit / debit — hanya tersedia melalui halaman Midtrans.
  return (
    <div className="mt-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-4">
      <div className="flex items-center gap-2">
        <CreditCard className="h-4 w-4 text-[#1768ad]" />
        <p className="text-sm font-bold text-[#17324d]">Kartu Kredit / Debit</p>
      </div>
      {payment?.midtrans_redirect_url ? (
        <>
          <p className="mt-2 text-xs text-[#698097]">
            Pembayaran kartu diproses melalui halaman pembayaran Midtrans yang aman.
          </p>
          <a
            href={payment.midtrans_redirect_url}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-3 inline-flex h-9 items-center rounded-md bg-[#1768ad] px-4 text-xs font-semibold text-white hover:bg-[#14548f]"
          >
            Lanjut ke Halaman Pembayaran
          </a>
        </>
      ) : (
        <p className="mt-2 text-xs text-[#b7791f]">
          Pembayaran kartu belum tersedia. Silakan pilih metode lain.
        </p>
      )}
    </div>
  );
}

export function PaymentPanel({ code }: { code?: string }) {
  const router = useRouter();
  const [phase, setPhase] = useState<Phase>("LOADING");
  const [booking, setBooking] = useState<Booking | null>(null);
  const [payment, setPayment] = useState<Payment | null>(null);
  const [destinations, setDestinations] = useState<PaymentDestination[]>([]);
  const [countdown, setCountdown] = useState<number>(0);
  const [method, setMethod] = useState<PaymentMethodCode>("QRIS");
  const [proofUploading, setProofUploading] = useState(false);
  const [proofError, setProofError] = useState<string | null>(null);
  // null = tidak ada pembatasan (mis. endpoint belum tersedia).
  const [activeMethods, setActiveMethods] = useState<string[] | null>(null);
  const proofInputRef = useRef<HTMLInputElement | null>(null);

  // Muat booking + payment + daftar tujuan pembayaran dari backend API.
  useEffect(() => {
    let alive = true;
    (async () => {
      if (!code) {
        setPhase("NOT_FOUND");
        return;
      }
      try {
        const [b, p, dest, methods] = await Promise.all([
          bookingService.getById(code),
          paymentService.getByBookingId(code),
          paymentService.listDestinations(),
          paymentService.listActiveMethods(),
        ]);
        if (!alive) return;
        setBooking(b);
        setPayment(p);
        setDestinations(dest);
        // Metode yang dinonaktifkan admin tidak ditampilkan ke pengunjung.
        if (methods.length > 0) {
          setActiveMethods(methods.map((m) => m.method_code));
        }
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

  // Polling status pembayaran: webhook Midtrans bisa menandai PAID kapan pun
  // (QRIS/e-wallet), sehingga status diperbarui tanpa menekan tombol apa pun.
  useEffect(() => {
    if (phase !== "WAITING" || !code) return;
    const interval = setInterval(async () => {
      try {
        const p = await paymentService.getByBookingId(code);
        if (!p) return;
        setPayment(p);
        if (p.status === "PAID") {
          const b = await bookingService.getById(code);
          setBooking(b);
          setPhase("SUCCESS");
          toast.success("Pembayaran berhasil!", {
            description: `E-tiket untuk ${b?.booking_code ?? code} telah diterbitkan.`,
          });
        }
      } catch {
        // Biarkan polling berlanjut; kegagalan jaringan tidak mengubah fase.
      }
    }, 5000);
    return () => clearInterval(interval);
  }, [phase, code]);

  const selected = useMemo(() => findPaymentMethod(method), [method]);
  const destination = useMemo(
    () => destinations.find((d) => d.method_code === method),
    [destinations, method]
  );

  // Grup metode yang masih aktif menurut pengaturan admin.
  const methodGroups = useMemo(
    () =>
      PAYMENT_METHOD_GROUPS.map((group) => ({
        ...group,
        options: group.options.filter((o) => !activeMethods || activeMethods.includes(o.code)),
      })).filter((group) => group.options.length > 0),
    [activeMethods]
  );

  // Bila metode terpilih ternyata dinonaktifkan, pindah ke metode aktif pertama.
  useEffect(() => {
    if (!activeMethods || activeMethods.includes(method)) return;
    const first = methodGroups[0]?.options[0];
    if (first) setMethod(first.code);
  }, [activeMethods, method, methodGroups]);

  async function pay(outcome?: PaymentOutcome) {
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
      } else if (res.outcome === "PENDING") {
        setPhase("WAITING");
        setPayment(res.payment);
        toast.info("Pembayaran masih diproses", {
          description: "Status pembayaran belum dikonfirmasi penyedia. Kami akan terus memeriksa.",
        });
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

  async function uploadProof(file: File) {
    if (!booking) return;
    setProofError(null);
    if (file.size > 5 * 1024 * 1024) {
      setProofError("Ukuran berkas maksimal 5 MB.");
      return;
    }
    setProofUploading(true);
    try {
      const updated = await paymentService.uploadProof(booking.id, file);
      setPayment(updated);
      toast.success("Bukti pembayaran terunggah", {
        description: "Pembayaran akan diverifikasi oleh admin.",
      });
    } catch (err: any) {
      const message = err?.message || "Gagal mengunggah bukti pembayaran.";
      setProofError(message);
      toast.error("Gagal mengunggah", { description: message });
    } finally {
      setProofUploading(false);
    }
  }

  // Langkah berikutnya hanya terbuka setelah bukti bayar terunggah.
  function goNext() {
    if (!booking) return;
    if (!payment?.proof_url) {
      toast.error("Unggah bukti pembayaran terlebih dahulu", {
        description: "Bukti bayar wajib diunggah sebelum melanjutkan.",
      });
      return;
    }
    if (payment.status === "PAID") {
      router.push(`/e-ticket?code=${booking.id}&new=1`);
      return;
    }
    router.push("/dashboard/booking");
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
              {methodGroups.map((group) => (
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
              {methodGroups.length === 0 && (
                <p className="rounded-md bg-[#fdf3e3] px-3 py-2 text-xs text-[#8a6420]">
                  Belum ada metode pembayaran aktif. Silakan hubungi admin.
                </p>
              )}
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
              <Instruction
                option={selected}
                bookingID={booking.id}
                total={booking.total}
                payment={payment}
                destination={destination}
              />

              <p className="mt-3 flex items-center justify-center gap-1.5 text-sm font-semibold text-[#b7791f]">
                <Timer className="h-4 w-4" /> Bayar dalam {mm}:{ss}
              </p>

              {/* Tombol bayar — membuat / memeriksa transaksi ke payment gateway. */}
              <button
                type="button"
                onClick={() => pay()}
                className="mt-4 inline-flex h-11 w-full items-center justify-center gap-2 rounded-md bg-[#1768ad] text-sm font-semibold text-white hover:bg-[#14548f]"
              >
                <Wallet className="h-4 w-4" /> Bayar Sekarang
              </button>

              {/* Unggah bukti bayar — wajib sebelum lanjut ke halaman berikutnya. */}
              <div className="mt-4 rounded-md border border-dashed border-[#c9deee] bg-[#f7fbff] p-4">
                <div className="flex items-center gap-2">
                  <ImageUp className="h-4 w-4 text-[#1768ad]" />
                  <p className="text-sm font-bold text-[#17324d]">Unggah Bukti Pembayaran</p>
                </div>
                <p className="mt-1 text-xs text-[#698097]">
                  Format JPG, PNG, WEBP, atau PDF (maks. 5 MB).
                </p>
                <input
                  ref={proofInputRef}
                  type="file"
                  accept="image/jpeg,image/png,image/webp,application/pdf"
                  className="hidden"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file) void uploadProof(file);
                    e.target.value = "";
                  }}
                />

                {payment?.proof_url ? (
                  <div className="mt-3 rounded-md border border-[#16876a] bg-[#f0fbf7] px-3 py-2">
                    <p className="flex items-center gap-1.5 text-xs font-semibold text-[#16876a]">
                      <CheckCircle2 className="h-3.5 w-3.5" /> Bukti pembayaran sudah diunggah
                    </p>
                    <a
                      href={`${process.env.NEXT_PUBLIC_API_ORIGIN ?? "http://localhost:8080"}${payment.proof_url}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="mt-1 inline-block text-[11px] font-semibold text-[#1768ad] hover:underline"
                    >
                      Lihat bukti
                    </a>
                  </div>
                ) : (
                  <p className="mt-2 text-xs font-semibold text-[#b7791f]">
                    Belum ada bukti pembayaran diunggah.
                  </p>
                )}

                {proofError && <p className="mt-2 text-xs text-[#c53030]">{proofError}</p>}

                <button
                  type="button"
                  disabled={proofUploading}
                  onClick={() => proofInputRef.current?.click()}
                  className="mt-3 inline-flex h-9 items-center gap-1.5 rounded-md border border-[#1768ad] px-3 text-xs font-semibold text-[#1768ad] hover:bg-[#eef7fd] disabled:opacity-60"
                >
                  {proofUploading ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <ImageUp className="h-3.5 w-3.5" />}
                  {proofUploading ? "Mengunggah..." : payment?.proof_url ? "Ganti Bukti" : "Pilih Berkas"}
                </button>
              </div>

              {/* Tombol lanjut — terkunci bila bukti bayar belum ada. */}
              <button
                type="button"
                onClick={goNext}
                disabled={!payment?.proof_url}
                aria-disabled={!payment?.proof_url}
                title={!payment?.proof_url ? "Unggah bukti pembayaran terlebih dahulu" : undefined}
                className="mt-3 inline-flex h-11 w-full items-center justify-center gap-2 rounded-md border border-[#1768ad] bg-white text-sm font-semibold text-[#1768ad] hover:bg-[#eef7fd] disabled:cursor-not-allowed disabled:border-[#d7e3ed] disabled:text-[#94A3B8]"
              >
                Lanjut ke Halaman Berikutnya <ArrowRight className="h-4 w-4" />
              </button>
              {!payment?.proof_url && (
                <p className="mt-1.5 text-center text-[11px] text-[#94A3B8]">
                  Unggah bukti pembayaran untuk melanjutkan.
                </p>
              )}

              <p className="mt-4 text-center text-xs text-[#698097]">
                Status pembayaran diperiksa otomatis setiap beberapa detik.
              </p>
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
                onClick={() => pay()}
                className="mt-2 rounded-md bg-[#1768ad] px-4 py-2 text-sm font-semibold text-white hover:bg-[#14548f]"
              >
                Cek Ulang Status
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
