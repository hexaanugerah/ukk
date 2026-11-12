package destinationinfo

import (
	"context"
	"time"

	"github.com/KawahPutih/backend/pkg/apperror"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, authorID string, req CreateRequest) (*DestinationInfo, error) {
	// Cek apakah judul sudah ada
	existing, err := s.repo.FindByTitle(ctx, req.Title)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing != nil {
		return nil, ErrTitleExists
	}

	info := &DestinationInfo{
		Title:           req.Title,
		Description:     req.Description,
		History:         req.History,
		DestinationDesc: req.DestinationDesc,
		Location:        req.Location,
		OperatingHours:  req.OperatingHours,
		Contact:         req.Contact,
		AdditionalInfo:  req.AdditionalInfo,
		Status:          StatusDraft,
		AuthorID:        authorID,
		CreatedBy:       authorID,
		UpdatedBy:       authorID,
	}
	if err := s.repo.Create(ctx, info); err != nil {
		return nil, apperror.Internal(err)
	}
	return info, nil
}

func (s *Service) Update(ctx context.Context, id, actorID string, req UpdateRequest) (*DestinationInfo, error) {
	info, err := s.mustFind(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cek apakah judul berubah dan judul baru sudah dipakai
	if info.Title != req.Title {
		existing, err := s.repo.FindByTitle(ctx, req.Title)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		if existing != nil && existing.ID != id {
			return nil, ErrTitleExists
		}
		info.Title = req.Title
	}

	info.Description = req.Description
	info.History = req.History
	info.DestinationDesc = req.DestinationDesc
	info.Location = req.Location
	info.OperatingHours = req.OperatingHours
	info.Contact = req.Contact
	info.AdditionalInfo = req.AdditionalInfo
	info.UpdatedBy = actorID

	if err := s.repo.Update(ctx, info); err != nil {
		return nil, apperror.Internal(err)
	}
	return info, nil
}

func (s *Service) Publish(ctx context.Context, id, actorID string) error {
	info, err := s.mustFind(ctx, id)
	if err != nil {
		return err
	}

	// Validasi field wajib sebelum publikasi
	if info.Title == "" || info.Description == "" {
		return ErrCannotPublish
	}

	now := time.Now().Unix()
	info.Status, info.PublishedAt, info.UpdatedBy = StatusPublished, &now, actorID
	if err := s.repo.Update(ctx, info); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Archive(ctx context.Context, id, actorID string) error {
	info, err := s.mustFind(ctx, id)
	if err != nil {
		return err
	}
	info.Status, info.UpdatedBy = StatusArchived, actorID
	if err := s.repo.Update(ctx, info); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id, actorID string) error {
	info, err := s.mustFind(ctx, id)
	if err != nil {
		return err
	}
	info.DeletedBy = &actorID
	if err := s.repo.Update(ctx, info); err != nil {
		return apperror.Internal(err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*DestinationInfo, error) {
	return s.mustFind(ctx, id)
}

func (s *Service) GetPublished(ctx context.Context) (*DestinationInfo, error) {
	return s.repo.GetPublished(ctx)
}

func (s *Service) List(ctx context.Context, offset, limit int, status string) ([]*DestinationInfo, int64, error) {
	items, total, err := s.repo.List(ctx, offset, limit, status)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return items, total, nil
}

func (s *Service) mustFind(ctx context.Context, id string) (*DestinationInfo, error) {
	info, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if info == nil {
		return nil, ErrNotFound
	}
	return info, nil
}
