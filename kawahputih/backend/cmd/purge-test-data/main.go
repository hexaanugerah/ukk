// Command purge-test-data menghapus data uji (booking, pembayaran, e-tiket,
// riwayat scan, dan audit terkait) dari database.
//
// Default: HANYA melaporkan apa yang cocok (dry-run). Tambahkan --apply untuk
// benar-benar menghapus. Kredensial dibaca dari environment/.env backend.
//
//	go run ./cmd/purge-test-data                 # laporan saja
//	go run ./cmd/purge-test-data --apply         # hapus permanen
//	go run ./cmd/purge-test-data --markers=test,dummy --apply
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/KawahPutih/backend/internal/config"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type bookingRow struct {
	ID          string
	BookingCode string
	VisitorName string
	Email       string
	Status      string
}

func main() {
	apply := flag.Bool("apply", false, "benar-benar hapus data (default hanya laporan)")
	markersFlag := flag.String("markers", "test,testing,tes,dummy,coba,contoh,percobaan,qwerty,asdf", "penanda data uji, dipisah koma")
	flag.Parse()

	markers := splitMarkers(*markersFlag)
	if len(markers) == 0 {
		fail("tidak ada penanda yang diberikan")
	}

	cfg, err := config.Load()
	if err != nil {
		fail("gagal memuat konfigurasi: " + err.Error())
	}
	db, err := config.NewDatabase(cfg.Database, logger.Warn)
	if err != nil {
		fail("gagal konek database: " + err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		fail("gagal akses koneksi database: " + err.Error())
	}
	defer sqlDB.Close()

	fmt.Printf("Database : %s@%s:%s/%s\n", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	fmt.Printf("Penanda  : %s\n", strings.Join(markers, ", "))
	fmt.Printf("Mode     : %s\n\n", modeLabel(*apply))

	bookings, err := findTestBookings(db, markers)
	if err != nil {
		fail("gagal mencari booking uji: " + err.Error())
	}

	bookingIDs := make([]string, 0, len(bookings))
	for _, b := range bookings {
		bookingIDs = append(bookingIDs, b.ID)
	}
	// Tampilkan contoh lebih dulu agar bisa diperiksa sebelum dihapus.
	printSample(bookings)
	fmt.Println()

	// Pembayaran: milik booking uji ATAU referensinya mengandung penanda.
	paymentIDs, paymentRefs, err := findTestPayments(db, markers, bookingIDs)
	if err != nil {
		fail("gagal mencari pembayaran uji: " + err.Error())
	}
	// E-tiket: milik booking uji.
	ticketIDs, err := findTicketIDs(db, bookingIDs)
	if err != nil {
		fail("gagal mencari e-tiket uji: " + err.Error())
	}
	// Riwayat scan: milik booking/tiket uji ATAU nama pengunjung mengandung penanda.
	scanIDs, err := findScanRecordIDs(db, markers, bookingIDs, ticketIDs)
	if err != nil {
		fail("gagal mencari riwayat scan uji: " + err.Error())
	}
	// Audit: menyasar entitas uji yang akan dihapus.
	auditIDs, err := findAuditLogIDs(db, bookingIDs, paymentIDs, ticketIDs)
	if err != nil {
		fail("gagal mencari audit uji: " + err.Error())
	}
	webhookIDs, err := findWebhookLogIDs(db, paymentRefs)
	if err != nil {
		fail("gagal mencari log webhook uji: " + err.Error())
	}
	reviewIDs, err := findReviewIDs(db, bookingIDs)
	if err != nil {
		fail("gagal mencari review uji: " + err.Error())
	}
	itemCount, err := countBookingItems(db, bookingIDs)
	if err != nil {
		fail("gagal menghitung item booking: " + err.Error())
	}

	fmt.Println("Ringkasan yang akan dihapus:")
	fmt.Printf("  bookings            : %d\n", len(bookingIDs))
	fmt.Printf("  booking_items       : %d\n", itemCount)
	fmt.Printf("  payments            : %d\n", len(paymentIDs))
	fmt.Printf("  tickets (e-tiket)   : %d\n", len(ticketIDs))
	fmt.Printf("  scan_records        : %d\n", len(scanIDs))
	fmt.Printf("  audit_logs          : %d\n", len(auditIDs))
	fmt.Printf("  payment_webhook_logs: %d\n", len(webhookIDs))
	fmt.Printf("  reviews             : %d\n", len(reviewIDs))

	fmt.Println()
	fmt.Println("Sisa data di database:")
	printRemaining(db)

	total := len(bookingIDs) + itemCount + len(paymentIDs) + len(ticketIDs) + len(scanIDs) + len(auditIDs) + len(webhookIDs) + len(reviewIDs)
	if total == 0 {
		fmt.Println("\nTidak ada data uji yang cocok. Tidak ada yang dihapus.")
		return
	}

	if !*apply {
		fmt.Println("\nMode laporan (dry-run). Tidak ada yang dihapus.")
		fmt.Println("Jalankan ulang dengan --apply untuk menghapus data di atas.")
		return
	}

	if err := purge(db, bookingIDs, paymentIDs, ticketIDs, scanIDs, auditIDs, webhookIDs, reviewIDs); err != nil {
		fail("penghapusan gagal (transaksi dibatalkan): " + err.Error())
	}
	fmt.Printf("\nSelesai. %d baris data uji dihapus.\n", total)
}

// printRemaining menunjukkan jumlah baris yang masih ada setelah pembersihan.
func printRemaining(db *gorm.DB) {
	for _, t := range []string{"bookings", "payments", "tickets", "scan_records", "audit_logs"} {
		if !tableExists(db, t) {
			continue
		}
		var n int64
		if err := db.Raw(`SELECT COUNT(*) FROM ` + t).Scan(&n).Error; err != nil {
			fmt.Printf("  %-14s: (gagal dibaca)\n", t)
			continue
		}
		fmt.Printf("  %-14s: %d\n", t, n)
	}
}

func modeLabel(apply bool) string {
	if apply {
		return "HAPUS (--apply)"
	}
	return "LAPORAN SAJA (dry-run)"
}

func splitMarkers(raw string) []string {
	out := make([]string, 0, 8)
	for _, part := range strings.Split(raw, ",") {
		p := strings.ToLower(strings.TrimSpace(part))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// markerClause membangun "(LOWER(col) LIKE ? OR ...)" untuk daftar kolom & penanda.
func markerClause(cols []string, markers []string) (string, []interface{}) {
	parts := make([]string, 0, len(cols)*len(markers))
	args := make([]interface{}, 0, len(cols)*len(markers))
	for _, c := range cols {
		for _, m := range markers {
			parts = append(parts, fmt.Sprintf("LOWER(%s) LIKE ?", c))
			args = append(args, "%"+m+"%")
		}
	}
	return "(" + strings.Join(parts, " OR ") + ")", args
}

func findTestBookings(db *gorm.DB, markers []string) ([]bookingRow, error) {
	clause, args := markerClause([]string{"visitor_name", "email", "phone", "booking_code", "note"}, markers)
	var rows []bookingRow
	err := db.Raw(
		`SELECT id, booking_code, visitor_name, email, status FROM bookings WHERE `+clause+` ORDER BY created_at DESC`,
		args...,
	).Scan(&rows).Error
	return rows, err
}

func findTestPayments(db *gorm.DB, markers []string, bookingIDs []string) ([]string, []string, error) {
	like, args := markerClause([]string{"reference"}, markers)
	where := like
	if len(bookingIDs) > 0 {
		where = "(" + like + " OR booking_id IN ?)"
		args = append(args, bookingIDs)
	}

	var rows []struct {
		ID        string
		Reference string
	}
	if err := db.Raw(`SELECT id, reference FROM payments WHERE `+where, args...).Scan(&rows).Error; err != nil {
		return nil, nil, err
	}
	ids := make([]string, 0, len(rows))
	refs := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
		if r.Reference != "" {
			refs = append(refs, r.Reference)
		}
	}
	return ids, refs, nil
}

func findTicketIDs(db *gorm.DB, bookingIDs []string) ([]string, error) {
	if len(bookingIDs) == 0 {
		return nil, nil
	}
	var rows []struct{ ID string }
	if err := db.Raw(`SELECT id FROM tickets WHERE booking_id IN ?`, bookingIDs).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.ID)
	}
	return out, nil
}

func findScanRecordIDs(db *gorm.DB, markers, bookingIDs, ticketIDs []string) ([]string, error) {
	like, args := markerClause([]string{"visitor_name", "petugas_name"}, markers)
	where := like
	if len(bookingIDs) > 0 {
		where += " OR booking_id IN ?"
		args = append(args, bookingIDs)
	}
	if len(ticketIDs) > 0 {
		where += " OR ticket_id IN ?"
		args = append(args, ticketIDs)
	}

	var rows []struct{ ID string }
	if err := db.Raw(`SELECT id FROM scan_records WHERE `+where, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.ID)
	}
	return out, nil
}

func findAuditLogIDs(db *gorm.DB, bookingIDs, paymentIDs, ticketIDs []string) ([]string, error) {
	all := make([]string, 0, len(bookingIDs)+len(paymentIDs)+len(ticketIDs))
	all = append(all, bookingIDs...)
	all = append(all, paymentIDs...)
	all = append(all, ticketIDs...)
	if len(all) == 0 {
		return nil, nil
	}
	var rows []struct{ ID string }
	if err := db.Raw(`SELECT id FROM audit_logs WHERE target_id IN ?`, all).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.ID)
	}
	return out, nil
}

func findWebhookLogIDs(db *gorm.DB, paymentRefs []string) ([]string, error) {
	if len(paymentRefs) == 0 || !tableExists(db, "payment_webhook_logs") {
		return nil, nil
	}
	var rows []struct{ ID string }
	if err := db.Raw(`SELECT id FROM payment_webhook_logs WHERE transaction_id IN ?`, paymentRefs).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.ID)
	}
	return out, nil
}

func findReviewIDs(db *gorm.DB, bookingIDs []string) ([]string, error) {
	if len(bookingIDs) == 0 || !tableHasColumn(db, "reviews", "booking_id") {
		return nil, nil
	}
	var rows []struct{ ID string }
	if err := db.Raw(`SELECT id FROM reviews WHERE booking_id IN ?`, bookingIDs).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.ID)
	}
	return out, nil
}

func countBookingItems(db *gorm.DB, bookingIDs []string) (int, error) {
	if len(bookingIDs) == 0 {
		return 0, nil
	}
	var n int64
	err := db.Raw(`SELECT COUNT(*) FROM booking_items WHERE booking_id IN ?`, bookingIDs).Scan(&n).Error
	return int(n), err
}

func tableExists(db *gorm.DB, name string) bool {
	var n int64
	if err := db.Raw(
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`, name,
	).Scan(&n).Error; err != nil {
		return false
	}
	return n > 0
}

func tableHasColumn(db *gorm.DB, table, column string) bool {
	var n int64
	if err := db.Raw(
		`SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`,
		table, column,
	).Scan(&n).Error; err != nil {
		return false
	}
	return n > 0
}

func printSample(bookings []bookingRow) {
	const limit = 10
	fmt.Println("Contoh booking yang terdeteksi sebagai data uji:")
	if len(bookings) == 0 {
		fmt.Println("  (tidak ada)")
		return
	}
	for i, b := range bookings {
		if i >= limit {
			fmt.Printf("  ... dan %d lainnya\n", len(bookings)-limit)
			break
		}
		fmt.Printf("  - %s | %s | %s | %s\n", b.BookingCode, b.VisitorName, b.Email, b.Status)
	}
}

// purge menghapus anak dulu lalu induk di dalam satu transaksi.
func purge(db *gorm.DB, bookingIDs, paymentIDs, ticketIDs, scanIDs, auditIDs, webhookIDs, reviewIDs []string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		steps := []struct {
			table string
			ids   []string
		}{
			{"audit_logs", auditIDs},
			{"scan_records", scanIDs},
			{"tickets", ticketIDs},
			{"payment_webhook_logs", webhookIDs},
			{"payments", paymentIDs},
			{"booking_items", nil}, // dihapus lewat booking_id di bawah
			{"reviews", reviewIDs},
			{"bookings", bookingIDs},
		}
		for _, step := range steps {
			if step.table == "booking_items" {
				if len(bookingIDs) == 0 {
					continue
				}
				if err := tx.Exec(`DELETE FROM booking_items WHERE booking_id IN ?`, bookingIDs).Error; err != nil {
					return fmt.Errorf("%s: %w", step.table, err)
				}
				continue
			}
			if len(step.ids) == 0 {
				continue
			}
			if err := tx.Exec(`DELETE FROM `+step.table+` WHERE id IN ?`, step.ids).Error; err != nil {
				return fmt.Errorf("%s: %w", step.table, err)
			}
		}
		return nil
	})
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "ERROR: "+msg)
	os.Exit(1)
}
