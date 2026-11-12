// KODE FIX (setelah diperbaiki).
// Snapshot dari backend/internal/domain/booking/... : versi SETELAH perbaikan
// terkait bug double booking.
// File ini hanya dokumen/arsip, bukan bagian dari build.
package booking_fix_snapshot

// CreateBooking — SETELAH FIX: menggunakan transaksi & kunci eksklusif untuk mencegah double booking.
// Dengan mutex / SELECT ... FOR UPDATE (atau row-level lock) dan pengecekan ulang kuota di dalam
// transaksi yang sama, double booking dicegah.
func (s *Service) CreateBooking(ctx context.Context, req CreateBookingRequest) (*Booking, error) {
    // (Contoh pseudocode kondisi sesudah fix)
    // 1) Mulai transaksi database.
    // 2) Lock record tiket (SELECT ... FOR UPDATE) / atau mutex di aplikasi.
    // 3) Cek ulang ketersediaan kuota di dalam transaksi yang sama.
    //    Jika tidak cukup → rollback → kembalian error "Kuota tiket tidak tersedia".
    // 4) Kurangi kuota, simpan booking_items beserta booking.
    // 5) Commit transaksi.
    // Hasil: double booking dicegah; hanya satu booking yang berhasil dalam kondisi kuota terpakai.
    // ...
    return booking, nil
}

// Perbedaaan utama sebelum vs sesudah fix:
// - Sebelum fix: tidak ada lock/transaction, sehingga double booking mungkin terjadi.
// - Sesudah fix: lock transaksi & pengecekan ulang kuota dalam satu transaksi, mencegah double booking.
