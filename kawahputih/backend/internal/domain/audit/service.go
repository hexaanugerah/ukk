package audit

import (
	"context"
	"encoding/json"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Log membuat entri audit. Tidak pernah menggagalkan operasi pemanggil karena error ditelan dan hanya dicatat internal. Panggil secara fire-and-forget.
func (s *Service) Log(ctx context.Context, actorID, actorRole, action, targetType, targetID, status string, metadata map[string]interface{}, ip, ua string) {
	// Best-effort: jangan blokir alur utama karena audit gagal. Di produksi, pertimbangkan queue async untuk lalu lintas tinggi.
	log := &AuditLog{
		ActorID:    actorID,
		ActorRole:  actorRole,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Status:     status,
		IPAddress:  ip,
		UserAgent:  ua,
	}
	if len(metadata) > 0 {
		if b, err := json.Marshal(metadata); err == nil {
			log.Metadata = string(b)
		}
	}
	_ = s.repo.Create(ctx, log) // Intentionally ignore error
}

// LogSuccess adalah pembungkus untuk operasi yang berhasil.
func (s *Service) LogSuccess(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string) {
	s.Log(ctx, actorID, actorRole, action, targetType, targetID, StatusSuccess, metadata, ip, ua)
}

// LogFailure adalah pembungkus untuk operasi yang gagal.
func (s *Service) LogFailure(ctx context.Context, actorID, actorRole, action, targetType, targetID string, metadata map[string]interface{}, ip, ua string) {
	s.Log(ctx, actorID, actorRole, action, targetType, targetID, StatusFailed, metadata, ip, ua)
}

// List mengembalikan audit log berhalaman dengan filter.
func (s *Service) List(ctx context.Context, page, limit int, actorID, action, targetType, targetID, status, from, to string) ([]AuditLog, int64, error) {
	offset := (page - 1) * limit
	if page < 1 {
		offset = 0
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, offset, limit, actorID, action, targetType, targetID, status, from, to)
}

// GetByID mengambil satu audit log berdasarkan ID.
func (s *Service) GetByID(ctx context.Context, id string) (*AuditLog, error) {
	return s.repo.GetByID(ctx, id)
}

// StatsByAction mengembalikan jumlah aksi dalam rentang tanggal.
func (s *Service) StatsByAction(ctx context.Context, from, to string) (map[string]int64, error) {
	return s.repo.CountByAction(ctx, from, to)
}

// StatsByActor mengembalikan jumlah aktor dalam rentang tanggal.
func (s *Service) StatsByActor(ctx context.Context, from, to string) (map[string]int64, error) {
	return s.repo.CountByActor(ctx, from, to)
}

// BuildMetadata adalah helper untuk membangun map metadata secara konsisten.
func BuildMetadata(keyValues ...interface{}) map[string]interface{} {
	if len(keyValues)%2 != 0 {
		return nil
	}
	m := make(map[string]interface{}, len(keyValues)/2)
	for i := 0; i < len(keyValues); i += 2 {
		key, ok := keyValues[i].(string)
		if !ok {
			continue
		}
		m[key] = keyValues[i+1]
	}
	return m
}

// BookingMetadata membangun metadata standar terkait booking.
func BookingMetadata(bookingCode, visitDate string, total int64, items []map[string]interface{}) map[string]interface{} {
	m := BuildMetadata(
		"booking_code", bookingCode,
		"visit_date", visitDate,
		"total", total,
		"items", items,
	)
	return m
}

// PaymentMetadata membangun metadata standar terkait pembayaran.
func PaymentMetadata(bookingCode, method string, amount int64, outcome string) map[string]interface{} {
	return BuildMetadata(
		"booking_code", bookingCode,
		"method", method,
		"amount", amount,
		"outcome", outcome,
	)
}

// TicketMetadata membangun metadata standar terkait tiket.
func TicketMetadata(bookingCode, qrData, gate, visitorName string) map[string]interface{} {
	return BuildMetadata(
		"booking_code", bookingCode,
		"qr_data", qrData,
		"gate", gate,
		"visitor_name", visitorName,
	)
}

// TicketTypeMetadata membangun metadata standar jenis tiket.
func TicketTypeMetadata(code, name, category string, price int64, capacity int) map[string]interface{} {
	return BuildMetadata(
		"code", code,
		"name", name,
		"category", category,
		"price", price,
		"capacity", capacity,
	)
}

// StatusChangeMetadata membangun metadata untuk perubahan status.
func StatusChangeMetadata(fromStatus, toStatus string) map[string]interface{} {
	return BuildMetadata(
		"from_status", fromStatus,
		"to_status", toStatus,
	)
}

// ContextKey untuk mengambil info audit dari context request.
const (
	CtxAuditActorID   = "audit_actor_id"
	CtxAuditActorRole = "audit_actor_role"
	CtxAuditIP        = "audit_ip"
	CtxAuditUA        = "audit_ua"
)

// ExtractAuditInfo menarik info audit dari context request.
func ExtractAuditInfo(ctx context.Context) (actorID, actorRole, ip, ua string) {
	if v := ctx.Value(CtxAuditActorID); v != nil {
		actorID, _ = v.(string)
	}
	if v := ctx.Value(CtxAuditActorRole); v != nil {
		actorRole, _ = v.(string)
	}
	if v := ctx.Value(CtxAuditIP); v != nil {
		ip, _ = v.(string)
	}
	if v := ctx.Value(CtxAuditUA); v != nil {
		ua, _ = v.(string)
	}
	return
}

// WithAuditInfo menambahkan info audit ke sebuah context.
func WithAuditInfo(ctx context.Context, actorID, actorRole, ip, ua string) context.Context {
	ctx = context.WithValue(ctx, CtxAuditActorID, actorID)
	ctx = context.WithValue(ctx, CtxAuditActorRole, actorRole)
	ctx = context.WithValue(ctx, CtxAuditIP, ip)
	ctx = context.WithValue(ctx, CtxAuditUA, ua)
	return ctx
}
