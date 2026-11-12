# docker/unfixed/ — arsip snapshot kode Go SEBELUM FIX (Sprint 3)

Folder ini memuat snapshot kode sumber Go **sebelum diperbaiki/diedit** untuk
dokumentasi Sprint 3.

Isi:
- auth_service_before_fix.go    → snapshot Authenticate & CreateUser (bug TC06, TC07)
- jwtutil_before_fix.go        → snapshot generate & RefreshTokenPair (bug TC25)

Catatan:
- File ini **bukan bagian dari build** dan tidak akan dijalankan.
- Fungsi & tipe yang digunakan adalah replika (snapshot) yang disertai komentar
  yang merujuk pada baris asli sebelum perubahan.
- Jika ingin membandingkan dengan kode sesudah fix, lihat file asli:
    - backend/internal/domain/auth/service.go
    - backend/pkg/jwtutil/jwtutil.go

Status dokumentasi Sprint 3:
- TC06: login tidak cek IsVerified → belum diedit; snapshot ada di auth_service_before_fix.go
- TC07: register tidak kirim email verifikasi otomatis → belum diedit; snapshot ada di auth_service_before_fix.go
- TC25: refresh token identik & revoke tidak memutus sesi → belum diedit; snapshot ada di jwtutil_before_fix.go
- TC21: tidak ada endpoint update gate/shift → belum diedit; tidak ada file .go snapshot khusus di sini (tergantung implementasi staff)
- TC27: GOOGLE_CLIENT_ID kosong → konfigurasi, bukan kode fungsi (catatan: GOOGLE_CLIENT_ID perlu diisi)
