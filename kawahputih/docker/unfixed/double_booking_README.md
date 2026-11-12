# docker/unfixed/ — double booking (arsip snapshot kode SEBELUM & SETELAH FIX)

Isi:
- double_booking_before_fix.go    → kondisi sebelum fix (double booking memungkinkan karena tanpa lock/transaction)
- double_booking_after_fix.go    → kondisi sesudah fix (transaksi & lock mencegah double booking)

Catatan:
- File ini **bukan bagian dari build** dan tidak akan dijalankan.
- Fungsi & tipe adalah replika (snapshot) yang disertai komentar perbandingan kondisi sebelum & sesudah fix.
- Snapshot ini dibuat untuk dokumentasi Sprint 3, bukan sebagai kode yang dijalankan.

Status perkembangan dokumentasi Sprint 3:
- Double booking → snapshot sebelum & sesudah fix ada di double_booking_before_fix.go dan double_booking_after_fix.go
