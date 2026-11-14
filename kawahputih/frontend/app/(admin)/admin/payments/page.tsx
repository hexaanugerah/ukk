"use client";

import { useEffect, useState } from "react";
import { CheckCircle2, Eye, ExternalLink, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { SearchInput } from "@/components/panel/panel-toolbar";
import { paymentService } from "@/services/payment.api.service";
import { rupiah } from "@/lib/business";
import { paymentMethodLabel, PAYMENT_METHOD_GROUPS } from "@/config/payment.config";
import type { Payment, PaymentStatus, SettlementStatus } from "@/types/api";

const STATUS_ID: Record<PaymentStatus, string> = {
  PENDING: "Menunggu",
  PROCESSING: "Diproses",
  PAID: "Berhasil",
  FAILED: "Gagal",
  EXPIRED: "Kadaluarsa",
  REFUNDED: "Refund",
};

const SETTLEMENT_ID: Record<SettlementStatus, string> = {
  UNSETTLED: "Belum cair",
  SETTLED: "Sudah cair",
  REFUNDED: "Refund",
};

const STATUS_OPTIONS = Object.keys(STATUS_ID) as PaymentStatus[];
const METHOD_OPTIONS = PAYMENT_METHOD_GROUPS.flatMap((g) => g.options);

export default function AdminPaymentsPage() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);
  const [detail, setDetail] = useState<Payment | null>(null);
  const [verifying, setVerifying] = useState(false);

  // Verifikasi bukti bayar manual → PAID. Hanya muncul bila bukti sudah diunggah;
  // backend menolak perubahan ke PAID tanpa bukti/verifikasi provider.
  async function verifyPayment(p: Payment) {
    if (!confirm(`Verifikasi bukti bayar transaksi ${p.reference || p.id} dan tandai LUNAS?`)) return;
    setVerifying(true);
    try {
      const updated = await paymentService.updateStatus(p.id, "PAID");
      setPayments((prev) => prev.map((x) => (x.id === updated.id ? updated : x)));
      setDetail(updated);
      toast.success("Pembayaran diverifikasi", { description: "E-tiket diterbitkan untuk booking ini." });
    } catch (error: any) {
      toast.error("Gagal verifikasi", { description: error.message });
    } finally {
      setVerifying(false);
    }
  }

  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");
  const [method, setMethod] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [page, setPage] = useState(1);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    (async () => {
      try {
        const data = await paymentService.list({
          page,
          limit: 10,
          status: status || undefined,
          method: method || undefined,
          from: from || undefined,
          to: to || undefined,
          q: q || undefined,
        });
        if (alive) setPayments(data);
      } catch (error: any) {
        toast.error("Gagal memuat pembayaran", { description: error.message });
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => {
      alive = false;
    };
  }, [page, status, method, from, to, q]);

  return (
    <div>
      <PageHeader title="Transaksi Pembayaran" subtitle="Seluruh transaksi terpusat lewat payment gateway" />

      <SectionCard
        title="Daftar Transaksi"
        action={<span className="text-sm font-semibold text-[#1768ad]">{payments.length} data</span>}
      >
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <SearchInput
            value={q}
            onChange={(v) => {
              setQ(v);
              setPage(1);
            }}
            placeholder="Cari ID / referensi transaksi…"
            className="w-full sm:max-w-xs"
          />
          <select
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter status pembayaran"
          >
            <option value="">Semua status</option>
            {STATUS_OPTIONS.map((s) => (
              <option key={s} value={s}>
                {STATUS_ID[s]}
              </option>
            ))}
          </select>
          <select
            value={method}
            onChange={(e) => {
              setMethod(e.target.value);
              setPage(1);
            }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter metode pembayaran"
          >
            <option value="">Semua metode</option>
            {METHOD_OPTIONS.map((m) => (
              <option key={m.code} value={m.code}>
                {m.label}
              </option>
            ))}
          </select>
          <input
            type="date"
            value={from}
            onChange={(e) => {
              setFrom(e.target.value);
              setPage(1);
            }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-2 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter tanggal mulai"
          />
          <input
            type="date"
            value={to}
            onChange={(e) => {
              setTo(e.target.value);
              setPage(1);
            }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-2 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter tanggal akhir"
          />
        </div>

        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>No</th>
                <th>ID Transaksi</th>
                <th>Metode</th>
                <th>Nominal</th>
                <th>Status</th>
                <th>Settlement</th>
                <th>Dibuat</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={8} className="py-10 text-center text-sm text-[#698097]">
                    Memuat…
                  </td>
                </tr>
              ) : payments.length === 0 ? (
                <tr>
                  <td colSpan={8} className="py-10 text-center text-sm text-[#94A3B8]">
                    Belum ada data pembayaran.
                  </td>
                </tr>
              ) : (
                payments.map((p, index) => (
                  <tr key={p.id}>
                    <td className="tabular-nums text-[#698097]">{(page - 1) * 10 + index + 1}.</td>
                    <td className="font-mono text-[11px] font-medium text-[#17324d]">
                      {p.reference || p.id.slice(0, 8)}
                    </td>
                    <td>{paymentMethodLabel(p.method)}</td>
                    <td className="tabular-nums font-semibold text-[#17324d]">{rupiah(p.amount)}</td>
                    <td>
                      <StatusPill status={STATUS_ID[p.status]} />
                    </td>
                    <td>
                      <span
                        className={`inline-block rounded-full px-2 py-0.5 text-[10px] font-bold ${
                          p.settlement_status === "SETTLED"
                            ? "bg-[#e6f7f0] text-[#16876a]"
                            : p.settlement_status === "REFUNDED"
                            ? "bg-[#fdeaea] text-[#c53030]"
                            : "bg-[#eef2f6] text-[#698097]"
                        }`}
                      >
                        {SETTLEMENT_ID[(p.settlement_status ?? "UNSETTLED") as SettlementStatus]}
                      </span>
                    </td>
                    <td className="text-[11px] text-[#698097]">
                      {new Date(p.created_at).toLocaleDateString("id-ID")}
                    </td>
                    <td>
                      <button
                        onClick={() => setDetail(p)}
                        className="grid h-7 w-7 place-items-center rounded-md border border-[#d7e3ed] text-[#29445e] hover:bg-[#f2f7fb]"
                        aria-label={`Detail ${p.reference || p.id}`}
                      >
                        <Eye className="h-3.5 w-3.5" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        <div className="mt-3 flex items-center justify-end gap-1 text-xs text-[#698097]">
          <button
            onClick={() => setPage(page - 1)}
            disabled={page <= 1}
            className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40"
          >
            ‹
          </button>
          <span className="px-2">Hal {page}</span>
          <button
            onClick={() => setPage(page + 1)}
            disabled={payments.length < 10}
            className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40"
          >
            ›
          </button>
        </div>
      </SectionCard>

      {detail && (
        <div
          className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4"
          onClick={() => setDetail(null)}
          role="dialog"
          aria-modal="true"
        >
          <div className="w-full max-w-md rounded-md bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between border-b border-[#e5edf3] pb-3">
              <h3 className="text-base font-bold text-[#17324d]">Detail Transaksi</h3>
              <button
                onClick={() => setDetail(null)}
                className="text-[#698097] hover:text-[#17324d]"
                aria-label="Tutup"
              >
                ✕
              </button>
            </div>
            <dl className="mt-3 space-y-2 text-sm">
              <Row label="Referensi" value={detail.reference || detail.id} />
              <Row label="ID Booking" value={detail.booking_id} />
              <Row label="Provider" value={detail.provider ?? "midtrans"} />
              <Row label="Metode" value={paymentMethodLabel(detail.method)} />
              <Row label="Nominal" value={rupiah(detail.amount)} />
              <Row label="Status" value={STATUS_ID[detail.status]} />
              <Row
                label="Settlement"
                value={SETTLEMENT_ID[(detail.settlement_status ?? "UNSETTLED") as SettlementStatus]}
              />
              <Row label="Dibuat" value={new Date(detail.created_at).toLocaleString("id-ID")} />
              {detail.paid_at && <Row label="Dibayar" value={new Date(detail.paid_at).toLocaleString("id-ID")} />}
              <Row label="Kedaluwarsa" value={new Date(detail.expired_at).toLocaleString("id-ID")} />
              {detail.proof_uploaded_at && (
                <Row label="Bukti diunggah" value={new Date(detail.proof_uploaded_at).toLocaleString("id-ID")} />
              )}
            </dl>

            {detail.proof_url ? (
              <div className="mt-3 flex flex-wrap items-center gap-2">
                <a
                  href={`${process.env.NEXT_PUBLIC_API_ORIGIN ?? "http://localhost:8080"}${detail.proof_url}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex h-9 items-center gap-1.5 rounded-md border border-[#1768ad] bg-white px-3 text-xs font-semibold text-[#1768ad] hover:bg-[#eef7fd]"
                >
                  <ExternalLink className="h-3.5 w-3.5" /> Lihat Bukti Pembayaran
                </a>
                {detail.status !== "PAID" && detail.status !== "REFUNDED" && (
                  <button
                    onClick={() => verifyPayment(detail)}
                    disabled={verifying}
                    className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#16876a] px-3 text-xs font-semibold text-white hover:bg-[#126e57] disabled:opacity-60"
                  >
                    {verifying ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : (
                      <CheckCircle2 className="h-3.5 w-3.5" />
                    )}
                    Verifikasi & Tandai Lunas
                  </button>
                )}
              </div>
            ) : (
              <p className="mt-3 rounded-md bg-[#fdf3e3] px-3 py-2 text-[11px] text-[#8a6420]">
                Belum ada bukti pembayaran. Status PAID tidak dapat diubah manual tanpa bukti atau verifikasi provider.
              </p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="shrink-0 text-[#698097]">{label}</dt>
      <dd className="break-all text-right font-medium text-[#17324d]">{value}</dd>
    </div>
  );
}
