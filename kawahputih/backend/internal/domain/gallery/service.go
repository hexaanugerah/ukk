package gallery

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// MaxImageSize membatasi ukuran berkas gambar galeri.
const MaxImageSize = 5 << 20 // 5 MB

// allowedImageExts adalah ekstensi gambar yang boleh diunggah.
var allowedImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
}

type Service struct {
	repo        Repository
	attachment  ImageStore
	defaultBase string
	mu          sync.Mutex
}

func NewService(repo Repository, attach ImageStore, defaultBase string) *Service {
	if defaultBase == "" {
		defaultBase = "/uploads/gallery"
	}
	return &Service{
		repo:        repo,
		attachment:  attach,
		defaultBase: defaultBase,
	}
}

// ImageUpload adalah berkas gambar yang diunggah admin.
type ImageUpload struct {
	Ext    string // termasuk titik, contoh ".png" — sudah divalidasi
	Reader io.Reader
}

func (s *Service) Create(ctx context.Context, uploaderID string, req CreateRequest, image *ImageUpload) (*Item, error) {
	if errs := validateStruct(req); len(errs) > 0 {
		return nil, apperror.Validation("Data galeri tidak valid.").WithCode("GALLERY_003")
	}
	if image == nil || image.Ext == "" || image.Reader == nil {
		return nil, apperror.Validation("File gambar wajib diunggah.").WithCode("GALLERY_004")
	}

	imageURL, err := s.storeImage(ctx, uploaderID, image)
	if err != nil {
		return nil, err
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

func (s *Service) Update(ctx context.Context, id, actorID string, req UpdateRequest, image *ImageUpload) (*Item, error) {
	if errs := validateStruct(req); len(errs) > 0 {
		return nil, apperror.Validation("Data galeri tidak valid.").WithCode("GALLERY_003")
	}
	item, err := s.mustFind(ctx, id)
	if err != nil {
		return nil, err
	}

	if image != nil {
		newURL, err := s.storeImage(ctx, actorID, image)
		if err != nil {
			return nil, err
		}
		item.ImageURL = newURL
	}

	item.Title = req.Title
	item.Category = req.Category
	item.SortOrder = req.SortOrder
	item.UpdatedBy = actorID
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, apperror.Internal(err)
	}
	return item, nil
}

func (s *Service) storeImage(ctx context.Context, uploaderID string, image *ImageUpload) (string, error) {
	if s.attachment == nil {
		return "", apperror.Internal(errors.New("gallery image store not configured"))
	}
	target, err := s.attachment.Save(ctx, uploaderID, image.Ext, image.Reader)
	if err != nil {
		return "", apperror.Validation("Gagal menyimpan file galeri: " + err.Error()).WithCode("GALLERY_001")
	}
	return target, nil
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

// ImageStore menyimpan berkas gambar galeri dan mengembalikan URL publiknya.
type ImageStore interface {
	Save(ctx context.Context, uploaderID, ext string, src io.Reader) (string, error)
}

func validateStruct(s interface{}) []response.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var out []response.FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		out = append(out, response.FieldError{Field: fe.Field(), Message: fe.Field() + " is invalid"})
	}
	return out
}

func ValidateImageExt(ext string) bool {
	return allowedImageExts[strings.ToLower(ext)]
}

type localDiskStore struct {
	dir       string
	urlPrefix string
	mu        sync.Mutex
}

// NewLocalDiskStore menyimpan berkas ke dir pada disk dan mengembalikan
// URL dengan awalan urlPrefix (harus didaftarkan sebagai rute statis).
func NewLocalDiskStore(dir, urlPrefix string) ImageStore {
	return &localDiskStore{dir: dir, urlPrefix: urlPrefix}
}

func (l *localDiskStore) Save(_ context.Context, _, ext string, src io.Reader) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.dir == "" || l.urlPrefix == "" {
		return "", errors.New("gallery image store path not configured")
	}
	if err := ensureDir(l.dir); err != nil {
		return "", err
	}

	name := newImageFileName(ext)
	if err := writeFile(filepath.Join(l.dir, name), src); err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join(l.urlPrefix, name)), nil
}
