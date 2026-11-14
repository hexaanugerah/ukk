package payments

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ===================== Settings / methods / webhook log repository =====================

// ConfigRepository menyimpan konfigurasi gateway dan kanal metode pembayaran.
type ConfigRepository interface {
	GetSetting(ctx context.Context, provider string) (*PaymentSetting, error)
	SaveSetting(ctx context.Context, s *PaymentSetting) error
	ListMethods(ctx context.Context) ([]PaymentMethodConfig, error)
	FindMethod(ctx context.Context, code string) (*PaymentMethodConfig, error)
	SaveMethod(ctx context.Context, m *PaymentMethodConfig) error
	CreateWebhookLog(ctx context.Context, l *PaymentWebhookLog) error
	FindWebhookLog(ctx context.Context, notificationID string) (*PaymentWebhookLog, error)
	SaveWebhookLog(ctx context.Context, l *PaymentWebhookLog) error
}

type configRepository struct{ db *gorm.DB }

func NewConfigRepository(db *gorm.DB) ConfigRepository { return &configRepository{db: db} }

func (r *configRepository) GetSetting(ctx context.Context, provider string) (*PaymentSetting, error) {
	var s PaymentSetting
	err := r.db.WithContext(ctx).Where("provider = ?", provider).Order("created_at ASC").First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *configRepository) SaveSetting(ctx context.Context, s *PaymentSetting) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *configRepository) ListMethods(ctx context.Context) ([]PaymentMethodConfig, error) {
	var out []PaymentMethodConfig
	err := r.db.WithContext(ctx).Order("display_order ASC, method_code ASC").Find(&out).Error
	return out, err
}

func (r *configRepository) FindMethod(ctx context.Context, code string) (*PaymentMethodConfig, error) {
	var m PaymentMethodConfig
	err := r.db.WithContext(ctx).Where("method_code = ?", code).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *configRepository) SaveMethod(ctx context.Context, m *PaymentMethodConfig) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *configRepository) CreateWebhookLog(ctx context.Context, l *PaymentWebhookLog) error {
	// Unique index pada notification_id membuat notifikasi ganda gagal di sini,
	// sehingga pemrosesan pembayaran tetap idempotent walau webhook dikirim ulang.
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *configRepository) FindWebhookLog(ctx context.Context, notificationID string) (*PaymentWebhookLog, error) {
	var l PaymentWebhookLog
	err := r.db.WithContext(ctx).Where("notification_id = ?", notificationID).First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *configRepository) SaveWebhookLog(ctx context.Context, l *PaymentWebhookLog) error {
	return r.db.WithContext(ctx).Save(l).Error
}

// ===================== Settings service =====================

var (
	ErrMethodNotFound = apperror.NotFound("Metode pembayaran tidak ditemukan.").WithCode("PAY_004")
	ErrSettingsUnset  = apperror.NotFound("Pengaturan pembayaran belum tersedia.").WithCode("PAY_005")
)

type SettingsService struct {
	repo      ConfigRepository
	gateway   MidtransGateway
	env       string
	webhookURL string
}

func NewSettingsService(repo ConfigRepository, gateway MidtransGateway, env, frontendURL string) *SettingsService {
	return &SettingsService{
		repo:       repo,
		gateway:    gateway,
		env:        env,
		webhookURL: strings.TrimRight(frontendURL, "/") + "/api/v1/payments/webhook",
	}
}

// ensureDefaults membuat baris pengaturan & daftar metode awal bila belum ada.
func (s *SettingsService) ensureDefaults(ctx context.Context) error {
	setting, err := s.repo.GetSetting(ctx, ProviderMidtrans)
	if err != nil {
		return err
	}
	if setting == nil {
		envName := "sandbox"
		if strings.EqualFold(s.env, "production") {
			envName = "production"
		}
		setting = &PaymentSetting{
			Provider:         ProviderMidtrans,
			Environment:      envName,
			Enabled:          true,
			SettlementStatus: "PENDING",
		}
		if err := s.repo.SaveSetting(ctx, setting); err != nil {
			return err
		}
	}

	methods, err := s.repo.ListMethods(ctx)
	if err != nil {
		return err
	}
	if len(methods) > 0 {
		return nil
	}
	defaults := []struct {
		code, name string
		order      int
	}{
		{MethodQRIS, "QRIS", 1},
		{MethodGoPay, "GoPay", 2},
		{MethodOVO, "OVO", 3},
		{MethodDANA, "DANA", 4},
		{MethodShopeePay, "ShopeePay", 5},
		{MethodBCAVA, "BCA Virtual Account", 6},
		{MethodBNIVA, "BNI Virtual Account", 7},
		{MethodMandiriVA, "Mandiri Virtual Account", 8},
		{MethodBRIVA, "BRI Virtual Account", 9},
		{MethodCard, "Kartu Kredit / Debit", 10},
		{MethodAlfamart, "Alfamart", 11},
		{MethodIndomaret, "Indomaret", 12},
	}
	for _, d := range defaults {
		m := &PaymentMethodConfig{
			Provider: ProviderMidtrans, MethodCode: d.code, MethodName: d.name,
			Enabled: true, DisplayOrder: d.order,
		}
		if err := s.repo.SaveMethod(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (s *SettingsService) Get(ctx context.Context) (*PaymentSettingsView, error) {
	if err := s.ensureDefaults(ctx); err != nil {
		return nil, apperror.Internal(err)
	}
	setting, err := s.repo.GetSetting(ctx, ProviderMidtrans)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if setting == nil {
		return nil, ErrSettingsUnset
	}
	configured := s.gateway != nil && s.gateway.Configured()
	return &PaymentSettingsView{
		Provider:                setting.Provider,
		Environment:             setting.Environment,
		Enabled:                 setting.Enabled,
		Connected:               configured && setting.Enabled,
		Configured:              configured,
		MerchantName:            setting.MerchantName,
		SettlementBank:          setting.SettlementBank,
		SettlementAccountMasked: setting.SettlementAccountMasked,
		SettlementStatus:        setting.SettlementStatus,
		WebhookURL:              s.webhookURL,
		LastSyncedAt:            setting.LastSyncedAt,
	}, nil
}

func (s *SettingsService) Update(ctx context.Context, req UpdateSettingsRequest) (*PaymentSettingsView, error) {
	if err := s.ensureDefaults(ctx); err != nil {
		return nil, apperror.Internal(err)
	}
	setting, err := s.repo.GetSetting(ctx, ProviderMidtrans)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if setting == nil {
		return nil, ErrSettingsUnset
	}
	if req.Enabled != nil {
		setting.Enabled = *req.Enabled
	}
	if v := strings.TrimSpace(req.MerchantName); v != "" {
		setting.MerchantName = v
	}
	if v := strings.TrimSpace(req.SettlementBank); v != "" {
		setting.SettlementBank = v
	}
	// Nomor rekening pencairan hanya boleh diubah lewat mekanisme resmi provider.
	// Aplikasi hanya menyimpan versi tersamar untuk ditampilkan.
	if v := strings.TrimSpace(req.SettlementAccountMasked); v != "" {
		setting.SettlementAccountMasked = maskAccount(v)
	}
	now := time.Now()
	setting.LastSyncedAt = &now
	if err := s.repo.SaveSetting(ctx, setting); err != nil {
		return nil, apperror.Internal(err)
	}
	return s.Get(ctx)
}

func (s *SettingsService) ListMethods(ctx context.Context) ([]PaymentMethodConfig, error) {
	if err := s.ensureDefaults(ctx); err != nil {
		return nil, apperror.Internal(err)
	}
	out, err := s.repo.ListMethods(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return out, nil
}

// ListEnabledMethods mengembalikan metode aktif untuk pengunjung.
func (s *SettingsService) ListEnabledMethods(ctx context.Context) ([]PaymentMethodConfig, error) {
	methods, err := s.ListMethods(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PaymentMethodConfig, 0, len(methods))
	for _, m := range methods {
		if m.Enabled {
			out = append(out, m)
		}
	}
	return out, nil
}

func (s *SettingsService) UpdateMethod(ctx context.Context, code string, req UpdateMethodRequest) (*PaymentMethodConfig, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if !ValidMethod(code) {
		return nil, ErrInvalidMethod
	}
	m, err := s.repo.FindMethod(ctx, code)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if m == nil {
		return nil, ErrMethodNotFound
	}
	if req.Enabled != nil {
		m.Enabled = *req.Enabled
	}
	if req.DisplayOrder != nil {
		m.DisplayOrder = *req.DisplayOrder
	}
	if err := s.repo.SaveMethod(ctx, m); err != nil {
		return nil, apperror.Internal(err)
	}
	return m, nil
}

// IsMethodEnabled dipakai backend untuk menolak metode yang dinonaktifkan admin.
func (s *SettingsService) IsMethodEnabled(ctx context.Context, code string) (bool, error) {
	m, err := s.repo.FindMethod(ctx, strings.ToUpper(strings.TrimSpace(code)))
	if err != nil {
		return false, apperror.Internal(err)
	}
	if m == nil {
		return true, nil // belum dikonfigurasi → ikuti daftar bawaan
	}
	return m.Enabled, nil
}

// maskAccount menyamarkan nomor rekening: hanya 4 digit terakhir terlihat.
func maskAccount(account string) string {
	clean := strings.ReplaceAll(strings.TrimSpace(account), " ", "")
	if len(clean) <= 4 {
		return "****"
	}
	return "****" + clean[len(clean)-4:]
}

// ===================== Webhook log =====================

// RecordWebhookLog mencatat notifikasi provider. Bila notification_id sudah ada,
// notifikasi dianggap duplikat dan MUST NOT diproses ulang.
func (s *SettingsService) RecordWebhookLog(ctx context.Context, notificationID, transactionID, eventType string) (isDuplicate bool, logID string, err error) {
	if s == nil || s.repo == nil {
		return false, "", nil
	}
	existing, err := s.repo.FindWebhookLog(ctx, notificationID)
	if err != nil {
		return false, "", apperror.Internal(err)
	}
	// Hanya log yang sudah selesai diproses yang dianggap duplikat. Log yang
	// masih RECEIVED (mis. proses terhenti) boleh diproses ulang saat retry.
	if existing != nil && (existing.ProcessingStatus == WebhookProcessed || existing.ProcessingStatus == WebhookDuplicate) {
		return true, existing.ID, nil
	}
	l := &PaymentWebhookLog{
		Provider:         ProviderMidtrans,
		TransactionID:    transactionID,
		NotificationID:   notificationID,
		EventType:        eventType,
		ProcessingStatus: WebhookReceived,
		ReceivedAt:       time.Now(),
	}
	if existing != nil {
		l = existing
		l.ReceivedAt = time.Now()
		l.ProcessingStatus = WebhookReceived
	}
	if err := s.repo.SaveWebhookLog(ctx, l); err != nil {
		// Bentrok unique → balapan antar-instance; perlakukan sebagai duplikat.
		return true, "", nil
	}
	return false, l.ID, nil
}

func (s *SettingsService) MarkWebhookLog(ctx context.Context, id, status string) {
	if s == nil || s.repo == nil || id == "" {
		return
	}
	l, err := s.repo.FindWebhookLog(ctx, id)
	if err != nil || l == nil {
		return
	}
	now := time.Now()
	l.ProcessingStatus = status
	l.ProcessedAt = &now
	_ = s.repo.SaveWebhookLog(ctx, l)
}

// ===================== Handler =====================

type SettingsHandler struct {
	service *SettingsService
}

func NewSettingsHandler(service *SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

func (h *SettingsHandler) Get(c *gin.Context) {
	view, err := h.service.Get(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment settings retrieved", view)
}

func (h *SettingsHandler) Update(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	view, err := h.service.Update(c.Request.Context(), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment settings updated", view)
}

func (h *SettingsHandler) ListMethods(c *gin.Context) {
	methods, err := h.service.ListMethods(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment methods retrieved", methods)
}

// ListPublicMethods hanya mengembalikan metode aktif untuk halaman pembayaran.
func (h *SettingsHandler) ListPublicMethods(c *gin.Context) {
	methods, err := h.service.ListEnabledMethods(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment methods retrieved", methods)
}

func (h *SettingsHandler) UpdateMethod(c *gin.Context) {
	var req UpdateMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	m, err := h.service.UpdateMethod(c.Request.Context(), c.Param("code"), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment method updated", m)
}

// formatAmount converts an int64 rupiah amount to the decimal string Midtrans expects.
func formatAmount(amount int64) string {
	return fmt.Sprintf("%d.00", amount)
}
