package payments

import (
	"context"
	"errors"
	"strings"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// ===================== Destination repository =====================

type DestinationRepository interface {
	Create(ctx context.Context, d *PaymentDestination) error
	Update(ctx context.Context, d *PaymentDestination) error
	FindByID(ctx context.Context, id string) (*PaymentDestination, error)
	List(ctx context.Context) ([]PaymentDestination, error)
	ListActive(ctx context.Context) ([]PaymentDestination, error)
	Delete(ctx context.Context, id string) error
}

type destinationRepository struct{ db *gorm.DB }

func NewDestinationRepository(db *gorm.DB) DestinationRepository {
	return &destinationRepository{db: db}
}

func (r *destinationRepository) Create(ctx context.Context, d *PaymentDestination) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *destinationRepository) Update(ctx context.Context, d *PaymentDestination) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *destinationRepository) FindByID(ctx context.Context, id string) (*PaymentDestination, error) {
	var d PaymentDestination
	err := r.db.WithContext(ctx).First(&d, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *destinationRepository) List(ctx context.Context) ([]PaymentDestination, error) {
	var out []PaymentDestination
	err := r.db.WithContext(ctx).Order("sort_order ASC, provider_name ASC").Find(&out).Error
	return out, err
}

func (r *destinationRepository) ListActive(ctx context.Context) ([]PaymentDestination, error) {
	var out []PaymentDestination
	err := r.db.WithContext(ctx).
		Where("active = ? AND account_number <> ?", true, "").
		Order("sort_order ASC, provider_name ASC").Find(&out).Error
	return out, err
}

func (r *destinationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&PaymentDestination{}, "id = ?", id).Error
}

// ===================== Destination service =====================

var ErrDestinationNotFound = apperror.NotFound("Tujuan pembayaran tidak ditemukan.").WithCode("PAYDEST_001")

type DestinationService struct {
	repo DestinationRepository
}

func NewDestinationService(repo DestinationRepository) *DestinationService {
	return &DestinationService{repo: repo}
}

func (s *DestinationService) Create(ctx context.Context, req DestinationRequest) (*PaymentDestination, error) {
	if errs := req.validate(); len(errs) > 0 {
		return nil, apperror.Validation("Data tujuan pembayaran tidak valid.").WithCode("PAYDEST_002")
	}
	d := &PaymentDestination{
		MethodCode:    strings.ToUpper(req.MethodCode),
		ProviderName:  strings.TrimSpace(req.ProviderName),
		AccountNumber: strings.TrimSpace(req.AccountNumber),
		AccountName:   strings.TrimSpace(req.AccountName),
		Active:        req.Active,
		SortOrder:     req.SortOrder,
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, apperror.Internal(err)
	}
	return d, nil
}

func (s *DestinationService) Update(ctx context.Context, id string, req DestinationRequest) (*PaymentDestination, error) {
	if errs := req.validate(); len(errs) > 0 {
		return nil, apperror.Validation("Data tujuan pembayaran tidak valid.").WithCode("PAYDEST_002")
	}
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing == nil {
		return nil, ErrDestinationNotFound
	}
	existing.MethodCode = strings.ToUpper(req.MethodCode)
	existing.ProviderName = strings.TrimSpace(req.ProviderName)
	existing.AccountNumber = strings.TrimSpace(req.AccountNumber)
	existing.AccountName = strings.TrimSpace(req.AccountName)
	existing.Active = req.Active
	existing.SortOrder = req.SortOrder
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, apperror.Internal(err)
	}
	return existing, nil
}

func (s *DestinationService) Delete(ctx context.Context, id string) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.Internal(err)
	}
	if existing == nil {
		return ErrDestinationNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *DestinationService) List(ctx context.Context) ([]PaymentDestination, error) {
	out, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return out, nil
}

// ListActive hanya untuk pengunjung: tujuan yang aktif DAN sudah diisi nomornya.
func (s *DestinationService) ListActive(ctx context.Context) ([]PaymentDestination, error) {
	out, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return out, nil
}

// ===================== Destination DTO & handler =====================

type DestinationRequest struct {
	MethodCode    string `json:"method_code" validate:"required,min=2,max=20"`
	ProviderName  string `json:"provider_name" validate:"required,min=1,max=100"`
	AccountNumber string `json:"account_number" validate:"required,min=1,max=100"`
	AccountName   string `json:"account_name" validate:"omitempty,max=100"`
	Active        bool   `json:"active"`
	SortOrder     int    `json:"sort_order"`
}

func (r DestinationRequest) validate() []response.FieldError {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		var out []response.FieldError
		for _, fe := range err.(validator.ValidationErrors) {
			out = append(out, response.FieldError{Field: fe.Field(), Message: fe.Field() + " is invalid"})
		}
		return out
	}
	if !ValidMethod(strings.ToUpper(strings.TrimSpace(r.MethodCode))) {
		return []response.FieldError{{Field: "method_code", Message: "metode pembayaran tidak didukung"}}
	}
	return nil
}

type DestinationHandler struct {
	service *DestinationService
}

func NewDestinationHandler(service *DestinationService) *DestinationHandler {
	return &DestinationHandler{service: service}
}

// ListActive adalah endpoint publik: tujuan pembayaran aktif untuk pengunjung.
func (h *DestinationHandler) ListActive(c *gin.Context) {
	out, err := h.service.ListActive(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment destinations retrieved", out)
}

func (h *DestinationHandler) List(c *gin.Context) {
	out, err := h.service.List(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment destinations retrieved", out)
}

func (h *DestinationHandler) Create(c *gin.Context) {
	var req DestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "payment destination created", result)
}

func (h *DestinationHandler) Update(c *gin.Context) {
	var req DestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	result, err := h.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment destination updated", result)
}

func (h *DestinationHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "payment destination deleted", nil)
}
