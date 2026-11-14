package visitors

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Nama role kanonik yang boleh dipilih saat membuat akun staf. Didefinisikan
// lokal agar domain visitors tidak mengimpor domain lain.
const (
	staffRoleAdmin   = "ADMIN"
	staffRoleManager = "MANAGER"
	staffRolePetugas = "PETUGAS"
)

// Pola validasi untuk akun staf baru.
var (
	emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	phonePattern = regexp.MustCompile(`^08\d{8,11}$`)
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// pengunjung

func (h *Handler) ListVisitors(c *gin.Context) {
	rows, err := h.service.repo.ListVisitors(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OK(c, "visitors retrieved", rows)
}

func (h *Handler) GetVisitor(c *gin.Context) {
	row, err := h.service.repo.FindVisitorByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	if row == nil {
		response.NotFound(c, "visitor not found")
		return
	}
	response.OK(c, "visitor retrieved", row)
}

func (h *Handler) VisitorStats(c *gin.Context) {
	stats, err := h.service.repo.VisitorStats(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OK(c, "visitor stats retrieved", stats)
}

// staf

func (h *Handler) ListStaff(c *gin.Context) {
	rows, err := h.service.repo.ListStaff(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OK(c, "staff retrieved", rows)
}

func (h *Handler) GetStaff(c *gin.Context) {
	row, err := h.service.repo.FindStaffByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	if row == nil {
		response.NotFound(c, "staff not found")
		return
	}
	response.OK(c, "staff retrieved", row)
}

func (h *Handler) StaffStats(c *gin.Context) {
	stats, err := h.service.repo.StaffStats(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OK(c, "staff stats retrieved", stats)
}

func (h *Handler) SetStaffStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if req.Status != "AKTIF" && req.Status != "NONAKTIF" {
		response.ValidationFailed(c, []response.FieldError{{Field: "status", Message: "status must be AKTIF or NONAKTIF"}})
		return
	}
	if err := h.service.repo.SetStaffActive(c.Request.Context(), c.Param("id"), req.Status == "AKTIF"); err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	response.OK(c, "staff status updated", nil)
}

type createStaffRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Gate     string `json:"gate"`
	Shift    string `json:"shift"`
}

// CreateStaff membuat akun staf baru (ADMIN/MANAGER/PETUGAS) dari panel admin.
func (h *Handler) CreateStaff(c *gin.Context) {
	var req createStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Phone = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(req.Phone, " ", ""), "-", ""))
	req.Role = strings.TrimSpace(strings.ToUpper(req.Role))

	var errs []response.FieldError
	if len(req.Name) < 3 {
		errs = append(errs, response.FieldError{Field: "name", Message: "Nama minimal 3 karakter"})
	}
	if !emailPattern.MatchString(req.Email) {
		errs = append(errs, response.FieldError{Field: "email", Message: "Email tidak valid"})
	}
	if !phonePattern.MatchString(req.Phone) {
		errs = append(errs, response.FieldError{Field: "phone", Message: "Nomor HP tidak valid (format 08xxxxxxxxxx)"})
	}
	errs = append(errs, validatePasswordStrength(req.Password)...)
	switch req.Role {
	case staffRoleAdmin, staffRoleManager, staffRolePetugas:
	default:
		errs = append(errs, response.FieldError{Field: "role", Message: "Role harus ADMIN, MANAGER, atau PETUGAS"})
	}
	if len(errs) > 0 {
		response.ValidationFailed(c, errs)
		return
	}

	err := h.service.repo.CreateStaff(c.Request.Context(), CreateStaffInput{
		Name: req.Name, Email: req.Email, Phone: req.Phone, Password: req.Password,
		Role: req.Role, Gate: strings.TrimSpace(req.Gate), Shift: strings.TrimSpace(req.Shift),
	})
	switch {
	case err == nil:
	case errors.Is(err, ErrEmailExists):
		response.Conflict(c, "Email sudah terdaftar. Gunakan email lain.")
		return
	case errors.Is(err, ErrRoleNotFound):
		response.ValidationFailed(c, []response.FieldError{{Field: "role", Message: "Role tidak ditemukan"}})
		return
	default:
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}

	response.Created(c, "Petugas baru berhasil ditambahkan", nil)
}

// validatePasswordStrength menegakkan kebijakan password yang sama dengan domain
// auth (8-72 byte, minimal huruf besar, huruf kecil, dan angka). Didefinisikan
// lokal agar domain visitors tidak mengimpor domain auth.
func validatePasswordStrength(password string) []response.FieldError {
	if len(password) < 8 || len([]byte(password)) > 72 {
		return []response.FieldError{{Field: "password", Message: "Password minimal 8 dan maksimal 72 karakter"}}
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return []response.FieldError{{Field: "password", Message: "Password harus mengandung huruf besar, huruf kecil, dan angka"}}
	}
	return nil
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return page, limit
}

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	// Data hanya-baca yang dipakai admin DAN manager: manager menjalankan monitoring dan pelaporan, jadi harus bisa mendaftar staf/pengunjung. Mutasi tetap khusus admin.
	read := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "MANAGER", "super_admin"))
	read.GET("/visitors", h.ListVisitors)
	read.GET("/visitors/stats", h.VisitorStats)
	read.GET("/visitors/:id", h.GetVisitor)
	read.GET("/staff", h.ListStaff)
	read.GET("/staff/stats", h.StaffStats)
	read.GET("/staff/:id", h.GetStaff)

	admin := rg.Group("/admin", requireAuth, middleware.RequireRole("ADMIN", "super_admin"))
	admin.POST("/staff", h.CreateStaff)
	admin.PATCH("/staff/:id/status", h.SetStaffStatus)
}
