package ticketing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/KawahPutih/backend/pkg/apperror"
	"go.uber.org/zap"
)

type BookingUsedMarker interface {
	SetCheckedIn(ctx context.Context, bookingID, gate string, at time.Time) error
	// IsOwner dipakai untuk membatasi akses e-tiket milik pengunjung lain.
	IsOwner(ctx context.Context, bookingID, userID string) (bool, error)
}

type AuditLogger interface {
	LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string)
	LogFailure(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string)
}

type Service struct {
	repo     Repository
	bookings BookingUsedMarker
	log      *zap.Logger
	audit    AuditLogger
}

func NewService(repo Repository, bookings BookingUsedMarker, log *zap.Logger, audit AuditLogger) *Service {
	return &Service{repo: repo, bookings: bookings, log: log, audit: audit}
}

// IssueForBooking membuat e-tiket untuk booking berbayar. Dipanggil domain payments saat pembayaran berhasil. Idempoten: penerbitan ulang mengembalikan tiket yang ada.
func (s *Service) IssueForBooking(ctx context.Context, snap *booking.Snapshot) (*TicketResponse, error) {
	existing, err := s.repo.FindByBookingID(ctx, snap.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing != nil {
		return toTicketResponse(existing), nil
	}

	// Booking multi-item menggabungkan setiap baris untuk nama tiket yang bisa dicetak.
	names := make([]string, 0, len(snap.Items))
	for _, item := range snap.Items {
		names = append(names, fmt.Sprintf("%s x%d", item.TicketName, item.Quantity))
	}
	ticketName := strings.Join(names, ", ")
	if len(ticketName) > 300 {
		ticketName = ticketName[:297] + "..."
	}

	qr, err := s.uniqueQRData(ctx)
	if err != nil {
		return nil, err
	}
	t := &Ticket{
		BookingID:   snap.ID,
		BookingCode: snap.BookingCode,
		VisitorName: snap.VisitorName,
		Email:       snap.Email,
		VisitDate:   snap.VisitDate,
		TicketName:  ticketName,
		Quantity:    totalQuantity(snap),
		Total:       snap.Total,
		Status:      TicketValid,
		GeneratedAt: time.Now(),
		QRData:      qr,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, apperror.Internal(err)
	}

	// Audit log: tiket diterbitkan
	if s.audit != nil {
		actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
		s.audit.LogSuccess(ctx, actorID, actorRole, audit.ActionTicketIssue, "ticket", t.ID,
			audit.TicketMetadata(snap.BookingCode, t.QRData, "", snap.VisitorName), ip, ua)
	}

	return toTicketResponse(t), nil
}

func totalQuantity(snap *booking.Snapshot) int {
	total := 0
	for _, item := range snap.Items {
		total += item.Quantity
	}
	return total
}

func (s *Service) uniqueQRData(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		buf := make([]byte, 9)
		if _, err := rand.Read(buf); err != nil {
			return "", apperror.Internal(err)
		}
		qr := "KPT-" + strings.ToUpper(hex.EncodeToString(buf))
		existing, err := s.repo.FindByQR(ctx, qr)
		if err != nil {
			return "", apperror.Internal(err)
		}
		if existing == nil {
			return qr, nil
		}
	}
	return "", apperror.Internal(fmt.Errorf("could not generate unique QR data"))
}

// GetByBookingID mengembalikan e-tiket booking (nil bila belum diterbitkan).
func (s *Service) GetByBookingID(ctx context.Context, bookingID string) (*TicketResponse, error) {
	t, err := s.repo.FindByBookingID(ctx, bookingID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if t == nil {
		return nil, nil
	}
	return toTicketResponse(t), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*TicketResponse, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if t == nil {
		return nil, nil
	}
	return toTicketResponse(t), nil
}

// CanAccess menentukan apakah peminta boleh melihat e-tiket sebuah booking.
// Petugas/admin yang privileged boleh melihat semua; pengunjung hanya miliknya.
func (s *Service) CanAccess(ctx context.Context, bookingID, userID string, privileged bool) (bool, error) {
	if privileged {
		return true, nil
	}
	if s.bookings == nil {
		return false, nil
	}
	return s.bookings.IsOwner(ctx, bookingID, userID)
}

// ListMine mengembalikan semua e-tiket milik pengunjung, dipakai dasbor "Tiket Saya".
func (s *Service) ListMine(ctx context.Context, userID string) ([]TicketResponse, error) {
	tickets, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	out := make([]TicketResponse, 0, len(tickets))
	for i := range tickets {
		out = append(out, *toTicketResponse(&tickets[i]))
	}
	return out, nil
}

func (s *Service) GetByCode(ctx context.Context, qrData string) (*TicketResponse, error) {
	t, err := s.repo.FindByQR(ctx, qrData)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if t == nil {
		return nil, nil
	}
	return toTicketResponse(t), nil
}

// qrPayloadRef menerima payload QR berbentuk JSON (mis. {"t":"<ticket-id>","b":"<booking-id>"})
// dan mengembalikan referensi tiket di dalamnya. Payload teks biasa dikembalikan apa adanya.
func qrPayloadRef(raw string) string {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "{") {
		return raw
	}
	var payload struct {
		TicketID  string `json:"t"`
		ID        string `json:"id"`
		BookingID string `json:"b"`
		Booking   string `json:"booking_id"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return raw
	}
	switch {
	case payload.TicketID != "":
		return payload.TicketID
	case payload.ID != "":
		return payload.ID
	case payload.BookingID != "":
		return payload.BookingID
	case payload.Booking != "":
		return payload.Booking
	}
	return raw
}

// GetByIDOrQR menyelesaikan referensi yang di-scan: UUID tiket, payload QR JSON, atau payload QR teks (dengan atau tanpa awalan KPT-).
func (s *Service) GetByIDOrQR(ctx context.Context, ref string) (*Ticket, error) {
	ref = qrPayloadRef(ref)
	ref = strings.TrimSpace(ref)
	ref = strings.TrimPrefix(ref, "KPT-")
	t, err := s.repo.FindByID(ctx, ref)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if t != nil {
		return t, nil
	}
	t, err = s.repo.FindByQR(ctx, "KPT-"+strings.ToUpper(ref))
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if t == nil {
		// percobaan terakhir: anggap input mentah sebagai payload QR lengkap
		t, err = s.repo.FindByQR(ctx, strings.ToUpper(ref))
		if err != nil {
			return nil, apperror.Internal(err)
		}
	}
	// Payload JSON dari e-tiket lama membawa booking id, bukan ticket id.
	if t == nil {
		t, err = s.repo.FindByBookingID(ctx, ref)
		if err != nil {
			return nil, apperror.Internal(err)
		}
	}
	return t, nil
}

// ValidateResult adalah putusan scanner.
type ValidateResult struct {
	Status  string
	Message string
	Ticket  *TicketResponse
}

func (s *Service) Validate(ctx context.Context, ref, gate, petugasID, petugasName string) (*ValidateResult, error) {
	t, err := s.GetByIDOrQR(ctx, ref)
	if err != nil {
		return nil, err
	}
	if t == nil {
		// Kegagalan pencarian tetap dicatat supaya riwayat aktivitas petugas lengkap.
		s.logValidation(ctx, nil, ref, gate, ScanInvalid, "Data tiket tidak ditemukan.")
		return &ValidateResult{Status: ScanInvalid, Message: "Data tiket tidak ditemukan."}, nil
	}

	record := func(status string) {
		scan := &ScanRecord{
			TicketID: t.ID, BookingID: t.BookingID, VisitorName: t.VisitorName,
			Gate: gate, PetugasID: petugasID, PetugasName: petugasName,
			Status: status, ScannedAt: time.Now(),
		}
		if err := s.repo.CreateScan(ctx, scan); err != nil {
			s.log.Error("scan record persist failed", zap.String("ticket_id", t.ID), zap.Error(err))
		}
	}

	switch t.Status {
	case TicketCancelled:
		record(ScanCancelled)
		s.logValidation(ctx, t, "", gate, ScanCancelled, "Tiket dibatalkan.")
		return &ValidateResult{Status: ScanCancelled, Message: "Tiket dibatalkan.", Ticket: toTicketResponse(t)}, nil
	case TicketExpired:
		record(ScanExpired)
		s.logValidation(ctx, t, "", gate, ScanExpired, "Tiket sudah kedaluwarsa.")
		return &ValidateResult{Status: ScanExpired, Message: "Tiket sudah kedaluwarsa.", Ticket: toTicketResponse(t)}, nil
	case TicketUsed:
		record(ScanUsed)
		s.logValidation(ctx, t, "", gate, ScanUsed, "Tiket sudah digunakan.")
		return &ValidateResult{Status: ScanUsed, Message: "Tiket sudah digunakan.", Ticket: toTicketResponse(t)}, nil
	}

	// Aturan bisnis: tiket hanya bisa di-scan pada tanggal kunjungannya.
	today := time.Now().Format("2006-01-02")
	if t.VisitDate != today {
		record(ScanInvalid)
		message := fmt.Sprintf("Tiket berlaku untuk %s, bukan hari ini.", t.VisitDate)
		s.logValidation(ctx, t, "", gate, ScanInvalid, message)
		return &ValidateResult{
			Status:  ScanInvalid,
			Message: message,
			Ticket:  toTicketResponse(t),
		}, nil
	}

	// Valid: bakar tiket (sekali pakai), tandai booking sudah check-in.
	now := time.Now()
	t.Status = TicketUsed
	t.CheckedInAt = &now
	t.CheckedInGate = gate
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, apperror.Internal(err)
	}
	if s.bookings != nil {
		if err := s.bookings.SetCheckedIn(ctx, t.BookingID, gate, now); err != nil {
			s.log.Error("booking checked-in stamp failed", zap.String("booking_id", t.BookingID), zap.Error(err))
		}
	}
	record(ScanValid)

	s.logValidation(ctx, t, "", gate, ScanValid, "Tiket valid, pengunjung dipersilakan masuk.")

	return &ValidateResult{Status: ScanValid, Message: "Tiket valid, pengunjung dipersilakan masuk.", Ticket: toTicketResponse(t)}, nil
}

// logValidation mencatat setiap hasil scan (valid maupun ditolak) ke audit log.
// Tanpa ini, riwayat aktivitas admin hanya memuat scan yang berhasil.
func (s *Service) logValidation(ctx context.Context, t *Ticket, fallbackRef, gate, status, message string) {
	if s.audit == nil {
		return
	}
	actorID, actorRole, ip, ua := audit.ExtractAuditInfo(ctx)
	targetID := fallbackRef
	// Kolom target_id hanya 36 karakter (UUID); payload QR mentah bisa lebih panjang.
	if len(targetID) > 36 {
		targetID = targetID[:36]
	}
	metadata := audit.BuildMetadata(
		"scan_status", status,
		"gate", gate,
		"message", message,
	)
	if t != nil {
		targetID = t.ID
		metadata["booking_code"] = t.BookingCode
		metadata["qr_data"] = t.QRData
		metadata["visitor_name"] = t.VisitorName
	}
	if status == ScanValid {
		s.audit.LogSuccess(ctx, actorID, actorRole, audit.ActionTicketValidate, "ticket", targetID, metadata, ip, ua)
		return
	}
	s.audit.LogFailure(ctx, actorID, actorRole, audit.ActionTicketValidate, "ticket", targetID, metadata, ip, ua)
}

// MonitoringResponse adalah ringkasan agregat untuk halaman monitoring manager.
// Tidak memuat payload QR atau identitas pengunjung, hanya angka operasional.
type MonitoringResponse struct {
	Totals     map[string]int64       `json:"totals"`
	TotalValid int64                  `json:"total_valid"`
	ValidToday int64                  `json:"valid_today"`
	PerPetugas []MonitoringPetugasRow `json:"per_petugas"`
	Hourly     []MonitoringHourlyRow  `json:"hourly"`
}

type MonitoringPetugasRow struct {
	Name         string     `json:"name"`
	Gate         string     `json:"gate"`
	Total        int64      `json:"total"`
	Valid        int64      `json:"valid"`
	LastActivity *time.Time `json:"last_activity,omitempty"`
}

type MonitoringHourlyRow struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

// Monitoring mengagregasi scan_records untuk dashboard monitoring manager:
// jumlah validasi per petugas dan per jam pada tanggal kunjungan tertentu.
func (s *Service) Monitoring(ctx context.Context, date string) (*MonitoringResponse, error) {
	scans, _, err := s.repo.ListScans(ctx, 0, 10000, "", "", "", "")
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	result := &MonitoringResponse{
		Totals:     map[string]int64{},
		PerPetugas: make([]MonitoringPetugasRow, 0),
		Hourly:     make([]MonitoringHourlyRow, 0, 10),
	}
	byPetugas := map[string]*MonitoringPetugasRow{}
	for i := range scans {
		scan := &scans[i]
		result.Totals[scan.Status]++
		key := scan.PetugasID
		if key == "" {
			key = scan.PetugasName
		}
		row := byPetugas[key]
		if row == nil {
			row = &MonitoringPetugasRow{Name: scan.PetugasName, Gate: scan.Gate}
			byPetugas[key] = row
		}
		if row.Gate == "" {
			row.Gate = scan.Gate
		}
		row.Total++
		if scan.Status == ScanValid {
			row.Valid++
		}
		if row.LastActivity == nil || scan.ScannedAt.After(*row.LastActivity) {
			at := scan.ScannedAt
			row.LastActivity = &at
		}
	}
	result.TotalValid = result.Totals[ScanValid]

	for _, row := range byPetugas {
		result.PerPetugas = append(result.PerPetugas, *row)
	}
	sort.Slice(result.PerPetugas, func(i, j int) bool {
		if result.PerPetugas[i].Valid == result.PerPetugas[j].Valid {
			return result.PerPetugas[i].Name < result.PerPetugas[j].Name
		}
		return result.PerPetugas[i].Valid > result.PerPetugas[j].Valid
	})

	counts := map[int]int64{}
	for i := range scans {
		scan := &scans[i]
		if scan.Status != ScanValid || scan.ScannedAt.Format("2006-01-02") != date {
			continue
		}
		counts[scan.ScannedAt.Hour()]++
		result.ValidToday++
	}
	for hour := 7; hour <= 16; hour++ {
		result.Hourly = append(result.Hourly, MonitoringHourlyRow{Hour: hour, Count: counts[hour]})
	}
	return result, nil
}

// SyncStatusForBooking menjaga tiket tetap konsisten dengan override status booking oleh admin (refund menjadi CANCELLED, expire menjadi EXPIRED, confirm menjadi VALID).
func (s *Service) SyncStatusForBooking(ctx context.Context, bookingID, bookingStatus string) error {
	var ticketStatus string
	switch bookingStatus {
	case booking.StatusCancelled:
		ticketStatus = TicketCancelled
	case booking.StatusExpired:
		ticketStatus = TicketExpired
	case booking.StatusPaid, booking.StatusConfirmed:
		ticketStatus = TicketValid
	case booking.StatusUsed:
		ticketStatus = TicketUsed
	default:
		return nil
	}
	if _, err := s.repo.SetStatusForBooking(ctx, bookingID, ticketStatus); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

// riwayat & statistik scanner

func (s *Service) History(ctx context.Context, page, limit int, gate, petugasID, from, to string) ([]ScanRecord, int64, error) {
	offset := (page - 1) * limit
	if page < 1 {
		offset = 0
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.ListScans(ctx, offset, limit, gate, petugasID, from, to)
}

// Stats mengembalikan penghitung dasbor petugas.
func (s *Service) Stats(ctx context.Context) (map[string]int64, error) {
	stats, err := s.repo.CountScansByStatus(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	today := time.Now().Format("2006-01-02")
	scansToday, err := s.repo.CountScansOn(ctx, today)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	stats["total"] = sumCounts(stats)
	stats["today"] = scansToday
	return stats, nil
}

func sumCounts(m map[string]int64) int64 {
	var total int64
	for _, v := range m {
		total += v
	}
	return total
}

// ValidTicketsOn dipakai analytics (perkiraan pengunjung untuk suatu tanggal).
func (s *Service) ValidTicketsOn(ctx context.Context, date string) (int64, error) {
	return s.repo.CountValidOn(ctx, date)
}
