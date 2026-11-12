package gallery

import (
	"context"
	"fmt"
	"sync"

	"github.com/KawahPutih/backend/pkg/apperror"
)

type Service struct {
	repo        Repository
	attachment  AttachmentStore
	defaultBase string
	mu          sync.Mutex
}

func NewService(repo Repository, attach AttachmentStore, defaultBase string) *Service {
	if defaultBase == "" {
		defaultBase = "/uploads/gallery"
	}
	return &Service{
		repo:        repo,
		attachment:  attach,
		defaultBase: defaultBase,
	}
}

func (s *Service) Create(ctx context.Context, uploaderID string, req CreateRequest) (*Item, error) {
	if err := ValidateRequest(req); err != nil {
		return nil, err
	}

	imageURL := req.ImageURL
	if req.ImageFile != "" {
		if s.attachment == nil {
			return nil, apperror.Internal(fmt.Errorf("gallery attachment store not configured"))
		}
		target, err := s.attachment.Store(ctx, req.ImageFile, uploaderID)
		if err != nil {
			return nil, apperror.Validation("Gagal menyimpan file galeri: " + err.Error()).WithCode("GALLERY_001")
		}
		imageURL = target
	}

	if imageURL == "" {
		return nil, apperror.Validation("Gambar tidak boleh kosong.").WithCode("GALLERY_010")
	}

	item := &Item{
		Title:      req.Title,
		ImageURL:   imageURL,
		Category:   req.Category,
		SortOrder:  req.SortOrder,
		UploadedBy: uploaderID,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, apperror.Internal(err)
	}
	return item, nil
}

func (s *Service) Update(ctx context.Context, id, actorID string, req UpdateRequest) (*Item, error) {
	if err := ValidateRequest(req); err != nil {
		return nil, err
	}
	item, err := s.mustFind(ctx, id)
	if err != nil {
		return nil, err
	}

	newURL := item.ImageURL
	if req.ImageFile != "" {
		if s.attachment == nil {
			return nil, apperror.Internal(fmt.Errorf("gallery attachment store not configured"))
		}
		target, err := s.attachment.Store(ctx, req.ImageFile, actorID)
		if err != nil {
			return nil, apperror.Validation("Gagal memperbarui file galeri: " + err.Error()).WithCode("GALLERY_002")
		}
		newURL = target
	} else if req.ImageURL != "" {
		newURL = req.ImageURL
	}
	if newURL == "" {
		return nil, apperror.Validation("Gambar tidak boleh kosong.").WithCode("GALLERY_011")
	}

	item.Title = req.Title
	item.ImageURL = newURL
	item.Category = req.Category
	item.SortOrder = req.SortOrder
	item.UpdatedBy = actorID
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, apperror.Internal(err)
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, id, actorID string) error {
	item, err := s.mustFind(ctx, id)
	if err != nil {
		return err
	}
	item.DeletedBy = &actorID
	if err := s.repo.Update(ctx, item); err != nil {
		return apperror.Internal(err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (*Item, error) {
	return s.mustFind(ctx, id)
}

func (s *Service) List(ctx context.Context, offset, limit int, category string) ([]Item, int64, error) {
	items, total, err := s.repo.List(ctx, offset, limit, category)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return items, total, nil
}

func (s *Service) Search(ctx context.Context, query string, offset, limit int) ([]Item, int64, error) {
	items, total, err := s.repo.Search(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return items, total, nil
}

func (s *Service) mustFind(ctx context.Context, id string) (*Item, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if item == nil {
		return nil, ErrItemNotFound
	}
	return item, nil
}

type AttachmentStore interface {
	Store(ctx context.Context, fileRef, uploaderID string) (string, error)
}

func ValidateRequest(req interface{}) error {
	if len(Validate(req)) > 0 {
		return apperror.Validation("Data galeri tidak valid.").WithCode("GALLERY_003")
	}
	switch value := req.(type) {
	case CreateRequest:
		if (value.ImageURL == "") == (value.ImageFile == "") {
			return apperror.Validation("Pilih satu sumber gambar: URL atau referensi file.").WithCode("GALLERY_003")
		}
		if (value.ImageURL != "" && !isHTTPImageURL(value.ImageURL)) || (value.ImageFile != "" && !isValidImageFileRef(value.ImageFile)) {
			return apperror.Validation("Sumber gambar tidak valid.").WithCode("GALLERY_003")
		}
	case UpdateRequest:
		if value.ImageURL != "" && value.ImageFile != "" {
			return apperror.Validation("Pilih satu sumber gambar: URL atau referensi file.").WithCode("GALLERY_003")
		}
		if (value.ImageURL != "" && !isHTTPImageURL(value.ImageURL)) || (value.ImageFile != "" && !isValidImageFileRef(value.ImageFile)) {
			return apperror.Validation("Sumber gambar tidak valid.").WithCode("GALLERY_003")
		}
	default:
		return apperror.Validation("Jenis permintaan galeri tidak didukung.").WithCode("GALLERY_003")
	}
	return nil
}

type localDiskStore struct {
	base string
	mu   sync.Mutex
}

func NewLocalDiskStore(base string) AttachmentStore {
	return &localDiskStore{base: base}
}

func (l *localDiskStore) Store(ctx context.Context, fileRef, uploaderID string) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.base == "" {
		return "", fmt.Errorf("gallery attachment base path not configured")
	}
	return fmt.Sprintf("%s/demo-%s.jpg", l.base, uploaderID), nil
}
