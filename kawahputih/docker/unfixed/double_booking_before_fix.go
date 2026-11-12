// KODE ASLI (kondisi SEBELUM FIX).
// Snapshot dari backend/internal/domain/booking/... : versi sebelum perbaikan
// terkait bug double booking: dua booking dilakukan pada waktu yang hampir bersamaan
// sehingga kuota tiket habis namun tidak terdeteksi bersaah hingga sesudahnya.
// File ini hanya dokumen/arsip, bukan bagian dari build.
package booking_snapshot

// CreateBooking — kondisi SEBELUM FIX: tidak ada pemeriksaan double booking yang memadai.
// Double booking bisa terjadi karena:
//   1) Pengecekan kuota & ketersediaan tiket dilakukan pada saat yang berbeda dengan insert.
//   2) Tanpa locking, transaksi berjalan paralel tanpa menjamin ketersedian kuota.
func (s *Service) CreateBooking(ctx context.Context, req CreateBookingRequest) (*Booking, error) {
    // (contoh pseudocode kondisi sebelum fix)
    // 1) Hitung kuota tiket & cek ketersediaan.
    // 2) Jika kuoting cukup, lanjut pembuatan booking.
    // 3) Simpan booking_items bersama booking.
    // Masalah: tanpa lock/transaction isolation yang tepat, double booking
    // memungkinkan dua sesi mengurangi kuota tiket sehingga melebihi kapasitas.
    // ...
    return booking, nil
}

// Permintaan double booking sebelum fix:
// - TCxx-DoubleBooking-1: dua pengunjung berbeda memesan kuota tiket yang sama hampir bersamaan
//   sehingga kuota terlanjur terpakai melebihi kapasitas yang tersedia.
// - TCxx-DoubleBooking-2: booking ganda oleh user yang sama dengan refresh cepat tidak terkekang.
