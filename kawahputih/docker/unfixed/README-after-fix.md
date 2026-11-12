# docker/unfixed/ — snapshot kode Go SEBELUM & SETELAH FIX (Sprint 3)

Folder ini memuat snapshot kode sumber Go dalam dua versi untuk dokumentasi Sprint 3:

## Versi SEBELUM FIX
- auth_service_before_fix.go    → snapshot Authenticate & CreateUser (bug TC06, TC07)
- jwtutil_before_fix.go        → snapshot generate & RefreshTokenPair (bug TC25)

## Versi SETELAH FIX
- auth_service_after_fix.go    → versi fix Authenticate (cek IsVerified) & CreateUser (kirim email verifikasi otomatis)
- jwtutil_after_fix.go        → versi fix generate (tambahkan jti unik) & penanganan duplikat revoke

Catatan:
- File ini **bukan bagian dari build** dan tidak akan dijalankan.
- Fungsi & tipe yang digunakan adalah replika (snapshot) yang disertai komentar
  yang membandingkan kondisi sebelum fix dengan sesudah fix.
- Snapshot ini dibuat untuk dokumentasi Sprint 3, bukan sebagai kode yang dijalankan.

Status perkembangan dokumentasi Sprint 3:
- TC06: login tidak cek IsVerified → snapshot sebelum & sesudah fix ada di auth_service_before_fix.go dan auth_service_after_fix.go
- TC07: register tidak kirim email verifikasi otomatis → snapshot sebelum & sesudah fix ada di auth_service_before_fix.go dan auth_service_after_fix.go
- TC25: refresh token identik & revoke tidak memutus sesi → snapshot sebelum & sesudah fix ada di jwtutil_before_fix.go dan jwtutil_after_fix.go
- TC21: tidak ada endpoint update gate/shift → belum diedit; tidak ada file .go snapshot khusus di sini
- TC27: GOOGLE_CLIENT_ID kosong → konfigurasi, bukan kode fungsi (catatan: GOOGLE_CLIENT_ID perlu diisi)

