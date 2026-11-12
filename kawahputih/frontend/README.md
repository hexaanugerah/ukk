# Frontend — Kawah Putih Rancabali (Part 2.3)

## Status jujur

Struktur folder mengikuti Part 2.3 secara lengkap. **Halaman fungsional
cuma dibangun untuk 7 domain yang backend-nya benar-benar ada** (auth,
user/staff, article, gallery, package, booking, dan sedikit role lewat
admin users). Sisa ~20 route yang didaftar dokumen (destinasi, kategori,
fasilitas, event, FAQ, banner, promo, voucher, kontak, laporan, analytics,
notifikasi, audit log, dst) dibuat sebagai **halaman "Coming Soon"** yang
jelas menyatakan dirinya belum tersambung ke backend — bukan halaman
kosong atau data palsu.

## Belum pernah di-`npm install` / `npm run build`

Sama seperti backend, saya tidak punya akses ke registry npm (`registry.npmjs.org`
sebenarnya ada di allowlist jaringan saya, jadi ini kemungkinan besar
bisa jalan — tapi saya belum sempat coba langsung). Jalankan:

```bash
cd frontend
npm install
npm run dev
```

dan laporkan kalau ada error TypeScript — kemungkinan ada 1-2 tipe yang
meleset di tempat yang tidak saya periksa ulang.

## Keputusan desain yang perlu kamu tahu

1. **Routing paket & artikel pakai ID, bukan slug** — backend hasil
   rebuild Part 2.2 cuma punya `FindByID` (bukan `FindBySlug` seperti versi
   sebelumnya). URL jadi `/packages/<uuid>` bukan `/packages/kawah-putih-basic`.
   Kurang SEO-friendly, tapi jujur sesuai kapasitas backend saat ini. Kalau mau
   URL slug, backend perlu ditambah endpoint `FindBySlug` dulu.

2. **QR Ticket di-generate di frontend** (`react-qr-code`), backend cuma
   kirim `ticket_code` — sesuai keputusan desain sejak Fase 2 backend.

3. **`store/notification.store.ts`, `wishlist.store.ts`, `language.store.ts`
   sengaja inert** (cuma nyimpen state lokal) — ditandai jelas di komentar
   kode karena tidak ada backend module yang mendukungnya (Notification,
   Wishlist server-side, i18n).

4. **Halaman profil read-only** — backend tidak punya endpoint update profil
   (FR-006 di BRD belum diimplementasi backend-nya).

5. **Check-in tiket manual (ketik kode)**, bukan scan kamera langsung —
   scanner QR fisik (kamera browser) belum diintegrasikan, cuma input teks.

6. **Role check di `middleware.ts` cuma cek keberadaan cookie**, bukan
   decode isi JWT — dijelaskan di komentar kenapa (menghindari duplikasi
   logic verifikasi signature di dua tempat). Guard role sesungguhnya
   tetap di backend (`RequireRole` middleware) + client-side check di
   `lib/permission.ts` untuk UX saja.

## Environment variable

Buat `.env.local` di folder `frontend/`:
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
```
