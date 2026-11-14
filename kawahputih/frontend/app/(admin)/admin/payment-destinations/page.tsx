"use client";

import { useEffect, useState } from "react";
import {
  CheckCircle2,
  Link2,
  Pencil,
  Plus,
  ShieldAlert,
  Trash2,
  Wallet,
  XCircle,
} from "lucide-react";
import { toast } from "sonner";
import { PageHeader, SectionCard } from "@/components/panel/panel-ui";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { paymentService, PaymentError } from "@/services/payment.api.service";
import { paymentMethodLabel } from "@/config/payment.config";
import type {
  PaymentDestination,
  PaymentDestinationPayload,
  PaymentMethodConfig,
  PaymentSettings,
} from "@/types/api";

// Tujuan Pembayaran — admin mengatur merchant, rekening pencairan, kanal metode
// pembayaran, serta daftar tujuan pembayaran yang tampil ke pengunjung.
// Kredensial gateway (server key) tetap hanya di environment backend.

const METHOD_OPTIONS = [
  "QRIS",
  "GOPAY",
  "OVO",
  "DANA",
  "SHOPEEPAY",
  "BCA_VA",
  "BNI_VA",
  "MANDIRI_VA",
  "BRI_VA",
  "ALFAMART",
  "INDOMARET",
  "CARD",
];

const emptyForm: PaymentDestinationPayload = {
  method_code: "BCA_VA",
  provider_name: "",
  account_number: "",
  account_name: "",
  active: true,
  sort_order: 0,
};

export default function PaymentDestinationsPage() {
  // ---- Tujuan pembayaran (rekening / nomor tujuan per metode) ----
  const [items, setItems] = useState<PaymentDestination[]>([]);
  const [loading, setLoading] = useState(true);
  const [form, setForm] = useState<PaymentDestinationPayload>(emptyForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);

  // ---- Merchant & pencairan dana ----
  const [settings, setSettings] = useState<PaymentSettings | null>(null);
  const [merchantName, setMerchantName] = useState("");
  const [settlementBank, setSettlementBank] = useState("");
  const [settlementMasked, setSettlementMasked] = useState("");
  const [gatewayEnabled, setGatewayEnabled] = useState(true);
  const [savingSettings, setSavingSettings] = useState(false);

  // ---- Metode pembayaran ----
  const [methods, setMethods] = useState<PaymentMethodConfig[]>([]);
  const [savingMethod, setSavingMethod] = useState<string | null>(null);

  async function loadDestinations() {
    setLoading(true);
    try {
      setItems(await paymentService.listDestinationsAdmin());
    } catch (error: any) {
      toast.error("Gagal memuat tujuan pembayaran", { description: error.message });
    } finally {
      setLoading(false);
    }
  }

  async function loadMerchant() {
    try {
      const [s, m] = await Promise.all([paymentService.getSettings(), paymentService.listMethods()]);
      setSettings(s);
      setMethods(m);
      setMerchantName(s.merchant_name);
      setSettlementBank(s.settlement_bank);
      setSettlementMasked(s.settlement_account_masked);
      setGatewayEnabled(s.enabled);
    } catch (error: any) {
      toast.error("Gagal memuat pengaturan merchant", { description: error.message });
    }
  }

  useEffect(() => {
    void loadDestinations();
    void loadMerchant();
  }, []);

  function startCreate() {
    setEditingId(null);
    setForm(emptyForm);
    setShowForm(true);
  }

  function startEdit(d: PaymentDestination) {
    setEditingId(d.id);
    setForm({
      method_code: d.method_code,
      provider_name: d.provider_name,
      account_number: d.account_number,
      account_name: d.account_name,
      active: d.active,
      sort_order: d.sort_order,
    });
    setShowForm(true);
  }

  function handleCancel() {
    setShowForm(false);
    setEditingId(null);
    setForm(emptyForm);
  }

  async function submit() {
    if (!form.method_code || !form.provider_name.trim() || !form.account_number.trim()) {
      toast.error("Metode, penyedia, dan nomor tujuan wajib diisi");
      return;
    }
    setSaving(true);
    try {
      const payload: PaymentDestinationPayload = {
        ...form,
        provider_name: form.provider_name.trim(),
        account_number: form.account_number.trim(),
        account_name: form.account_name?.trim(),
      };
      if (editingId) {
        await paymentService.updateDestination(editingId, payload);
        toast.success("Tujuan pembayaran diperbarui");
      } else {
        await paymentService.createDestination(payload);
        toast.success("Tujuan pembayaran ditambahkan");
      }
      handleCancel();
      await loadDestinations();
    } catch (error: any) {
      toast.error("Gagal menyimpan", {
        description: error instanceof PaymentError ? error.message : error.message,
      });
    } finally {
      setSaving(false);
    }
  }

  async function remove(d: PaymentDestination) {
    if (!confirm(`Hapus tujuan pembayaran ${d.provider_name} (${paymentMethodLabel(d.method_code)})?`)) return;
    try {
      await paymentService.deleteDestination(d.id);
      toast.success("Tujuan pembayaran dihapus");
      await loadDestinations();
    } catch (error: any) {
      toast.error("Gagal menghapus", { description: error.message });
    }
  }

  async function saveMerchant() {
    setSavingSettings(true);
    try {
      const updated = await paymentService.updateSettings({
        enabled: gatewayEnabled,
        merchant_name: merchantName.trim() || undefined,
        settlement_bank: settlementBank.trim() || undefined,
        settlement_account_masked: settlementMasked.trim() || undefined,
      });
      setSettings(updated);
      setSettlementMasked(updated.settlement_account_masked);
      toast.success("Data merchant disimpan");
    } catch (error: any) {
      toast.error("Gagal menyimpan merchant", { description: error.message });
    } finally {
      setSavingSettings(false);
    }
  }

  async function toggleMethod(m: PaymentMethodConfig) {
    setSavingMethod(m.method_code);
    try {
      const updated = await paymentService.updateMethod(m.method_code, { enabled: !m.enabled });
      setMethods((prev) => prev.map((x) => (x.method_code === updated.method_code ? updated : x)));
      toast.success(`${m.method_name} ${updated.enabled ? "diaktifkan" : "dinonaktifkan"}`);
    } catch (error: any) {
      toast.error("Gagal mengubah metode", { description: error.message });
    } finally {
      setSavingMethod(null);
    }
  }

  return (
    <div>
      <PageHeader
        title="Tujuan Pembayaran"
        subtitle="Atur merchant, rekening pencairan, dan kanal pembayaran yang tampil ke pengunjung"
        action={
          <Button onClick={startCreate}>
            <Plus className="mr-1 h-4 w-4" /> Tambah Tujuan
          </Button>
        }
      />

      {/* Merchant + pencairan dana */}
      <SectionCard
        title="Merchant & Pencairan Dana"
        subtitle="Perubahan rekening pencairan tetap harus lewat mekanisme resmi provider"
        className="mb-4"
      >
        {settings ? (
          <dl className="grid gap-2 text-sm sm:grid-cols-2">
            <Row label="Provider" value={settings.provider} />
            <Row label="Mode" value={settings.environment === "production" ? "Production" : "Sandbox"} />
            <Row
              label="Koneksi gateway"
              value={settings.connected ? "Terhubung" : settings.configured ? "Terkonfigurasi, nonaktif" : "Belum terkonfigurasi"}
              tone={settings.connected ? "ok" : "warn"}
            />
            <Row label="Server key" value={settings.configured ? "Tersedia" : "Belum diisi"} tone={settings.configured ? "ok" : "warn"} />
            <Row label="Status rekening" value={settings.settlement_status} />
            <Row
              label="Sinkronisasi terakhir"
              value={settings.last_synced_at ? new Date(settings.last_synced_at).toLocaleString("id-ID") : "—"}
            />
            <Row label="Webhook URL" value={settings.webhook_url} mono />
          </dl>
        ) : (
          <p className="py-3 text-sm text-[#698097]">Memuat pengaturan merchant…</p>
        )}

        {settings && !settings.configured && (
          <div className="mt-3 flex items-start gap-2 rounded-md border border-[#fdf3e3] bg-[#fdf3e3] px-3 py-2">
            <ShieldAlert className="mt-0.5 h-4 w-4 shrink-0 text-[#b7791f]" />
            <p className="text-[11px] leading-relaxed text-[#8a6420]">
              MIDTRANS_SERVER_KEY belum diisi di environment backend. Isi (mode Sandbox untuk pengujian) lalu restart
              server. Server key tidak boleh ditaruh di frontend.
            </p>
          </div>
        )}

        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          <label className="block">
            <span className="mb-1 block text-xs font-semibold text-[#29445e]">Nama Pemilik Merchant</span>
            <Input
              value={merchantName}
              onChange={(e) => setMerchantName(e.target.value)}
              placeholder="mis. PT Kawah Putih Wisata"
            />
          </label>
          <label className="block">
            <span className="mb-1 block text-xs font-semibold text-[#29445e]">Nama Bank Pencairan</span>
            <Input value={settlementBank} onChange={(e) => setSettlementBank(e.target.value)} placeholder="mis. BCA" />
          </label>
          <label className="block">
            <span className="mb-1 block text-xs font-semibold text-[#29445e]">Nomor Rekening (disamarkan)</span>
            <Input
              value={settlementMasked}
              onChange={(e) => setSettlementMasked(e.target.value)}
              placeholder="****1234"
            />
            <span className="mt-1 block text-[11px] text-[#94A3B8]">Hanya 4 digit terakhir disimpan & ditampilkan.</span>
          </label>
          <div className="flex items-end">
            <div className="flex w-full items-center justify-between rounded-md border border-[#e5edf3] bg-[#f7fbff] px-3 py-2">
              <div>
                <p className="text-sm font-semibold text-[#17324d]">Gateway aktif</p>
                <p className="text-[11px] text-[#698097]">Nonaktifkan untuk menghentikan transaksi baru.</p>
              </div>
              <button
                type="button"
                role="switch"
                aria-checked={gatewayEnabled}
                aria-label="Aktifkan gateway"
                onClick={() => setGatewayEnabled((v) => !v)}
                className={`relative h-6 w-11 shrink-0 rounded-full transition-colors ${
                  gatewayEnabled ? "bg-[#16876a]" : "bg-[#c9d6e2]"
                }`}
              >
                <span
                  className={`absolute top-0.5 h-5 w-5 rounded-full bg-white transition-all ${
                    gatewayEnabled ? "left-[22px]" : "left-0.5"
                  }`}
                />
              </button>
            </div>
          </div>
        </div>

        <Button className="mt-4" onClick={saveMerchant} loading={savingSettings}>
          Simpan Merchant
        </Button>
      </SectionCard>

      {/* Kanal metode pembayaran */}
      <SectionCard
        title="Metode Pembayaran"
        subtitle="Status di aplikasi bukan bukti metode aktif di akun merchant"
        className="mb-4"
      >
        <div className="divide-y divide-[#e5edf3]">
          {methods.map((m) => (
            <div key={m.method_code} className="flex flex-wrap items-center justify-between gap-3 py-3">
              <div className="flex items-center gap-2">
                <Wallet className="h-4 w-4 text-[#1768ad]" />
                <div>
                  <p className="text-sm font-semibold text-[#17324d]">
                    {m.method_name}
                    <span className="ml-2 rounded bg-[#f2f8fe] px-1.5 py-0.5 text-[10px] font-bold text-[#1768ad]">
                      {m.method_code}
                    </span>
                  </p>
                  <p className="text-[11px] text-[#698097]">Urutan tampil: {m.display_order}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <span
                  className={`inline-flex items-center gap-1 text-[11px] font-bold ${
                    m.enabled ? "text-[#16876a]" : "text-[#94A3B8]"
                  }`}
                >
                  {m.enabled ? <CheckCircle2 className="h-3.5 w-3.5" /> : <XCircle className="h-3.5 w-3.5" />}
                  {m.enabled ? "Aktif" : "Nonaktif"}
                </span>
                <button
                  type="button"
                  role="switch"
                  aria-checked={m.enabled}
                  aria-label={`${m.enabled ? "Nonaktifkan" : "Aktifkan"} ${paymentMethodLabel(m.method_code)}`}
                  disabled={savingMethod === m.method_code}
                  onClick={() => toggleMethod(m)}
                  className={`relative h-6 w-11 shrink-0 rounded-full transition-colors disabled:opacity-60 ${
                    m.enabled ? "bg-[#16876a]" : "bg-[#c9d6e2]"
                  }`}
                >
                  <span
                    className={`absolute top-0.5 h-5 w-5 rounded-full bg-white transition-all ${
                      m.enabled ? "left-[22px]" : "left-0.5"
                    }`}
                  />
                </button>
              </div>
            </div>
          ))}
          {methods.length === 0 && (
            <p className="py-6 text-center text-sm text-[#698097]">Belum ada metode pembayaran.</p>
          )}
        </div>
        <p className="mt-3 flex items-start gap-1.5 text-[11px] text-[#94A3B8]">
          <Link2 className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          Ketersediaan final ditentukan akun merchant di provider. Aktifkan hanya metode yang sudah disetujui di
          dashboard Midtrans.
        </p>
      </SectionCard>

      {/* Form tambah/ubah tujuan pembayaran */}
      {showForm && (
        <SectionCard title={editingId ? "Ubah Tujuan Pembayaran" : "Tambah Tujuan Pembayaran"} className="mb-4">
          <div className="grid gap-3 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium">Metode</label>
              <select
                value={form.method_code}
                onChange={(e) => setForm({ ...form, method_code: e.target.value })}
                className="h-10 w-full rounded-md border border-[#d7e3ed] bg-white px-3 text-sm outline-none focus:border-[#1768ad]"
                aria-label="Metode pembayaran"
              >
                {METHOD_OPTIONS.map((m) => (
                  <option key={m} value={m}>
                    {paymentMethodLabel(m)}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Nama Penyedia / Bank</label>
              <Input
                value={form.provider_name}
                onChange={(e) => setForm({ ...form, provider_name: e.target.value })}
                placeholder="mis. BCA, DANA"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Nomor Rekening / E-Wallet / Kode</label>
              <Input
                value={form.account_number}
                onChange={(e) => setForm({ ...form, account_number: e.target.value })}
                placeholder="mis. 1234567890"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Nama Pemilik (a.n.)</label>
              <Input
                value={form.account_name}
                onChange={(e) => setForm({ ...form, account_name: e.target.value })}
                placeholder="mis. Kawah Putih Ciwidey"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Urutan Tampil</label>
              <Input
                type="number"
                value={form.sort_order ?? 0}
                onChange={(e) => setForm({ ...form, sort_order: Number(e.target.value) })}
              />
            </div>
            <div className="flex items-end gap-2">
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={form.active}
                  onChange={(e) => setForm({ ...form, active: e.target.checked })}
                  className="h-4 w-4"
                />
                Aktif (tampil ke pengunjung)
              </label>
            </div>
          </div>
          <div className="mt-4 flex gap-2">
            <Button onClick={submit} loading={saving}>
              {editingId ? "Simpan Perubahan" : "Tambah"}
            </Button>
            <Button variant="ghost" onClick={handleCancel}>
              Batal
            </Button>
          </div>
        </SectionCard>
      )}

      <SectionCard title="Daftar Tujuan Pembayaran">
        {loading ? (
          <div className="animate-pulse space-y-3">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="h-12 rounded border border-[#e5edf3] bg-gray-200" />
            ))}
          </div>
        ) : items.length === 0 ? (
          <div className="py-8 text-center text-sm text-[#698097]">
            <Wallet className="mx-auto mb-2 h-8 w-8 text-[#c9deee]" />
            Belum ada tujuan pembayaran. Tambahkan rekening atau nomor e-wallet Anda.
          </div>
        ) : (
          <div className="divide-y divide-[#e5edf3]">
            {items.map((d) => (
              <div key={d.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                <div className="min-w-0">
                  <p className="text-sm font-semibold text-[#17324d]">
                    {d.provider_name}{" "}
                    <span className="ml-1 rounded bg-[#f2f8fe] px-1.5 py-0.5 text-[10px] font-bold text-[#1768ad]">
                      {paymentMethodLabel(d.method_code)}
                    </span>
                    {!d.active && (
                      <span className="ml-1 rounded bg-[#fdf3e3] px-1.5 py-0.5 text-[10px] font-bold text-[#8a6420]">
                        Nonaktif
                      </span>
                    )}
                    {!d.account_number && (
                      <span className="ml-1 rounded bg-[#fdeaea] px-1.5 py-0.5 text-[10px] font-bold text-[#c53030]">
                        Nomor kosong
                      </span>
                    )}
                  </p>
                  <p className="mt-0.5 text-xs text-[#698097]">
                    {d.account_number
                      ? `${d.account_number} • a.n. ${d.account_name || d.provider_name}`
                      : "Nomor tujuan belum diisi"}
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button size="sm" variant="ghost" onClick={() => startEdit(d)} aria-label={`Ubah ${d.provider_name}`}>
                    <Pencil className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="text-red-600 hover:text-red-700"
                    onClick={() => remove(d)}
                    aria-label={`Hapus ${d.provider_name}`}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </SectionCard>
    </div>
  );
}

function Row({
  label,
  value,
  tone,
  mono,
}: {
  label: string;
  value: string;
  tone?: "ok" | "warn";
  mono?: boolean;
}) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="shrink-0 text-[#698097]">{label}</dt>
      <dd
        className={`text-right font-medium ${
          tone === "ok" ? "text-[#16876a]" : tone === "warn" ? "text-[#b7791f]" : "text-[#17324d]"
        } ${mono ? "break-all font-mono text-[11px]" : ""}`}
      >
        {value}
      </dd>
    </div>
  );
}
