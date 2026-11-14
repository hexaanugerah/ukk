// App config — pengaturan umum aplikasi (brand, kapasitas, biaya, dsb). Semua nilai yang mungkin perlu diubah pemilik proyek dikumpulkan di sini.

export const APP_CONFIG = {
  name: "Kawah Putih Rancabali",
  shortName: "Kawah Putih",
  tagline: "Wisata Alam Bandung",
  currency: "IDR",
  /** Kapasitas pengunjung per hari — batas ketersediaan tanggal pemesanan. */
  dailyVisitorCapacity: 1000,
  operationalHour: { open: "07:00", close: "17:00" },
  contact: {
    address: "Jl. Raya Soreang - Ciwidey No.KM.25, Ciwidey, Bandung, Jawa Barat",
    phone: "022-1234567",
    email: "info@kawahputih.com",
  },
} as const;

export const BOOKING_CONFIG = {
  /** Tanggal minimum pemesanan (hari dari sekarang). */
  minDaysAhead: 0,
  /** Batas maksimum pemesanan ke depan (hari). 365 = boleh memesan hingga 1 tahun ke depan. */
  maxDaysAhead: 365,
  /** Maksimum tiket per jenis dalam satu pesanan. */
  maxQuantityPerTicket: 20,
  /**
   * Biaya layanan (% dari subtotal). HARUS sama dengan backend
   * (internal/domain/booking/service.go: serviceFeeRate = 2.0) karena backend
   * adalah satu-satunya sumber kebenaran saat transaksi disimpan.
   */
  serviceFeePercent: 2,
  /** Persentase diskon promo (0 = nonaktif). */
  discountPercent: 0,
  /** Pajak (0 = nonaktif). */
  taxPercent: 0,
  /** Kadaluarsa pembayaran QRIS (menit). */
  paymentExpiryMinutes: 15,
} as const;