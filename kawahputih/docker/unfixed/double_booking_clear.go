// KODE PLACEHOLDER — DOUBLE BOOKING (klarifikasi & perbandingan sebelum vs sesudah fix)
// File ini hanya arsip/dokumentasi. Bukan bagian dari build, bukan kode asli yang ada.

package booking_clear_snapshot

// ============================================================================
// SITUASI SEBELUM FIX
// ============================================================================
// Single booking berjalan normal; tapi tanpa koordinasi yang tepat, jika dua
// permintaan booking masuk hampir bersamaan (parallel), keduanya bisa lolos
// pengecekan kuota saat yang sama, sehingga kuota tiket melebihi kapasitas.

// Pseudocode: sebelum fix (single-threaded-looking tapi rentan paralel)
func CreateBooking_BEFORE(ctx context.Context, ticketID string, qty int) (string, error) {
    // 1) Hitung sisa kuota tiket (baca jumlah terjual + stok).
    sisa := countTiketSold(ticketID)         // Baca jumlah terjual saat ini
    cap := capTicket(ticketID)                // Kapasitas maks

    // 2) Cek kuota → bila cukup, proceed.
    if sisa+qty > cap {
        return "", fmt.Errorf("kuota tiket tidak tersedia")
    }

    // 3) Buat booking.
    //    ---- GAP: di sini terdapat jendela waktu (race window) ----
    //    Permintaan lain bisa membaca kuota yang sama SEBELUM perubahan
    //    dari request ini disimpan, sehingga kedua request lolos.
    bookingID := insertBooking(ticketID, qty)

    // 4) Simpan traveler, dll.
    insertBookingItems(bookingID, ticketID, qty)

    return bookingID, nil
}

// Catatan: tanpa LOCK / transaksi terisolasi, dua permintaan bisa membaca
// kuota yang sama (misal: sisa=2, cap=3, qty=2) → keduanya lolos, total
// terjual jadi 4 > cap=3 → double booking terjadi.

// ============================================================================
// SITUASI SESUDAH FIX
// ============================================================================
// Menggunakan transaksi database + kunci eksklusif / row lock sehingga
// proses pengecekan dan pengurangan kuota berjalan atomik. Double booking
// dicegah: hanya satu booking yang berhasil jika kuota terbatas.

// Pseudocode: sesudah fix (dengan transaksi & lock)
func CreateBooking_AFTER(ctx context.Context, ticketID string, qty int) (string, error) {
    // 1) Mulai transaksi (dengan isolation yang tepat).
    tx, err := beginTransaction(ctx)
    if err != nil {
        return "", err
    }
    defer tx.Rollback() // rollback otomatis jika belum commit

    // 2) Lock record tiket (misal: SELECT ... FOR UPDATE, atau gunakan
    //    mutex aplikasi jika DB tidak mendukung FOR UPDATE di konteks ini).
    //    Hanya satu request yang bisa membaca & memodifikasi kuota dalam satu waktu.
    lockTicketRow(ticketID, tx)

    // 3) Cek ulang kuota di dalam transaksi yang sama (setelah lock).
    sisa := countTiketSold(ticketID)     // sesudah lock, nilai stabil
    cap := capTicket(ticketID)

    if sisa+qty > cap {
        return "", fmt.Errorf("kuota tiket tidak tersedia") // rollback di defer
    }

    // 4) Kurangi kuota / tandai terjual (atomik dalam transaksi).
    //    (misal: update tabel ticket_statistics / jurnal penjualan)
    decrementQuota(ticketID, qty, tx)

    // 5) Simpan booking & booking_items.
    bookingID := insertBooking(ticketID, qty, tx)
    insertBookingItems(bookingID, ticketID, qty, tx)

    // 6) Commit.
    if err := tx.Commit(); err != nil {
        return "", err
    }
    return bookingID, nil
}

// Perbedaaan utama:
// - Sebelum fix: pengecekan kuota & insert dipisah; ada race terhadap kuota.
// - Sesudah fix: lock + pengecekan ulang + insert dalam satu transaksi atomik
//   sehingga double booking dicegah.

// ============================================================================
// CONTOH SKENARIO DOUBLE BOOKING (ilustrasi)
// ============================================================================
// Misal: tiket kapasitas 3, terjual 1 (sisa 2).
//   Request A: qty=2
//   Request B: qty=2  (datang sedikit lebih lambat, tapi masih dalam jendela race)
//
// Sebelum fix:
//   - A baca sisa=2 → lolos (2+2=4 > 3 tapi pengecekan salah karena belum
//     diperbarui). Setelah A selesai, kuota berkurang.
//   - B baca sisa yang sama (belum termutakhir) → lolos.
//   - Total terjual = 4 > 3 → double booking.
//
// Sesudah fix (dengan lock):
//   - A acquire lock → baca sisa=2 → lolos (kuota 2, cap 3, sisa+2=4 >3? Tidak, sisa 2 + qty 2 = 4 > cap 3, jadi error, mengecek: mana yang benar? Jika sisa 2+2=4 > cap3 → A gagal; kalau A lolos sisa harus 1+2=3 <= cap3; ilustrasi perlu angka lebih tepat).
//
// (Catatan: suatu ilustrasi yang tepat bisa disesuaikan nanti jika ditemukan kasus nyata.)

// ============================================================================
// STATUS DOKUMENTASI SPRINT 3
// ============================================================================
// - Double booking → placeholder ini dibuat karena kode asli terkait double
//   booking belum ditemukan/ditemukan di project saat ini.
// - Temuan: concurrency/race condition dalam booking kuota tersedia sebagai
//   potensi rawan double booking jika tidak ada lock/transaction.
// - Aksi: belum diedit; pending menemuan kode asli atau penambahan mekanisme
//   pencegahan double booking (lock/transaction).
