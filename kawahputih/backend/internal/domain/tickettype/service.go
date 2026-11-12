package tickettype

import (
	"context"
	"fmt"
	"strings"

	"github.com/KawahPutih/backend/pkg/apperror"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, offset, limit int, status, category string) ([]TicketType, int64, error) {
	if status != "" && status != StatusActive && status != StatusInactive {
		return nil, 0, invalidStatus()
	}
	out, total, err := s.repo.List(ctx, offset, limit, status, category)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return out, total, nil
}

func (s *Service) ListActive(ctx context.Context) ([]TicketType, error) {
	out, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return out, nil
}

func (s *Service) Categories(ctx context.Context) ([]string, error) {
	cats, err := s.repo.CountCategories(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return cats, nil
}

func (s *Service) Get(ctx context.Context, id string) (*TicketType, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if t == nil {
		return nil, ErrNotFound
	}
	return t, nil
}

func (s *Service) GetByCode(ctx context.Context, code string) (*TicketType, error) {
	t, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if t == nil {
		return nil, ErrNotFound
	}
	return t, nil
}

// GetMany menyelesaikan sekumpulan ID dalam satu panggilan, dipakai domain booking (lewat adapter bootstrap) untuk menghitung harga pesanan multi-item.
func (s *Service) GetMany(ctx context.Context, ids []string) ([]TicketType, error) {
	out := make([]TicketType, 0, len(ids))
	for _, id := range ids {
		t, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		if t == nil {
			return nil, ErrNotFound
		}
		out = append(out, *t)
	}
	return out, nil
}

func validCategory(c string) bool {
	for _, v := range Categories {
		if v == c {
			return true
		}
	}
	return false
}

func (s *Service) Create(ctx context.Context, actorID string, req UpsertRequest) (*TicketType, error) {
	if req.Price <= 0 {
		return nil, invalidPrice()
	}
	if req.Capacity <= 0 {
		return nil, invalidCapacity()
	}
	if !validCategory(req.Category) {
		return nil, ErrInvalidCategory
	}
	status, err := parseStatus(req.Status)
	if err != nil {
		return nil, err
	}
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		code = generateCode(req.Name)
	}
	existing, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing != nil {
		return nil, ErrCodeExists
	}
	t := &TicketType{
		Code: code, Name: req.Name, Description: req.Description, Category: req.Category,
		Price: req.Price, Capacity: req.Capacity, IsActive: status == StatusActive,
		CreatedBy: actorID, UpdatedBy: actorID,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, apperror.Internal(err)
	}
	return t, nil
}

func (s *Service) Update(ctx context.Context, id, actorID string, req UpsertRequest) (*TicketType, error) {
	t, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Price <= 0 {
		return nil, invalidPrice()
	}
	if req.Capacity <= 0 {
		return nil, invalidCapacity()
	}
	if !validCategory(req.Category) {
		return nil, ErrInvalidCategory
	}
	status, err := parseStatus(req.Status)
	if err != nil {
		return nil, err
	}
	t.Name = req.Name
	t.Description = req.Description
	t.Category = req.Category
	t.Price = req.Price
	t.Capacity = req.Capacity
	t.IsActive = status == StatusActive
	t.UpdatedBy = actorID
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, apperror.Internal(err)
	}
	return t, nil
}

func (s *Service) SetStatus(ctx context.Context, id, actorID, status string) (*TicketType, error) {
	t, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	st, err := parseStatus(status)
	if err != nil {
		return nil, err
	}
	t.IsActive = st == StatusActive
	t.UpdatedBy = actorID
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, apperror.Internal(err)
	}
	return t, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	t, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	// Hanya soft-delete; booking lama tetap merujuk barisnya.
	if err := s.repo.Delete(ctx, t.ID); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func parseStatus(status string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "", StatusActive:
		return StatusActive, nil
	case StatusInactive:
		return StatusInactive, nil
	}
	return "", invalidStatus()
}

// generateCode membuat kode yang mudah dibaca dari nama; keunikan terhadap DB dicek pemanggil (ErrCodeExists).
func generateCode(name string) string {
	base := "TKT-"
	n := strings.ToUpper(name)
	for _, r := range n {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			base += string(r)
		}
	}
	if len(base) > 14 {
		base = base[:14]
	}
	return fmt.Sprintf("%s-%04d", base, len(name)%10000)
}
