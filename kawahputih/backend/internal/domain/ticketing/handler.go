package ticketing

import (
	"strconv"

	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase *Usecase
}

func NewHandler(usecase *Usecase) *Handler { return &Handler{usecase: usecase} }

// Validate memverifikasi tiket yang di-scan dan menandainya terpakai bila valid.
func (h *Handler) Validate(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	// staffName: JWT hanya membawa ID; nama tampilan petugas diambil dari request bila ada, selain itu ID.
	staffName := req.Petugas
	if staffName == "" {
		staffName = middleware.CurrentUserID(c)
	}
	result, err := h.usecase.ValidateTicket(c.Request.Context(), req, middleware.CurrentUserID(c), staffName)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, result.Message, result)
}

// History mengembalikan riwayat verifikasi (petugas + admin).
func (h *Handler) History(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	petugas := c.Query("petugas")
	records, total, err := h.usecase.History(c.Request.Context(), page, limit,
		c.Query("gate"), petugas, c.Query("from"), c.Query("to"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "scan history retrieved", records, response.BuildMeta(page, limit, total))
}

// Stats mengembalikan penghitung scanner untuk dasbor petugas.
func (h *Handler) Stats(c *gin.Context) {
	stats, err := h.usecase.Stats(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "scan stats retrieved", stats)
}

// ManagerMonitoring mengembalikan agregat operasional petugas untuk manager/admin.
func (h *Handler) ManagerMonitoring(c *gin.Context) {
	result, err := h.usecase.Monitoring(c.Request.Context(), c.Query("date"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "monitoring retrieved", result)
}

// LookupTicket adalah pencarian detail hanya-baca (berdasarkan UUID tiket, awalan kode booking, atau payload QR). Berbeda dari Validate, tidak pernah membakar tiket atau menulis record scan.
func (h *Handler) LookupTicket(c *gin.Context) {
	ref := c.Query("ref")
	if len(ref) < 4 {
		response.BadRequest(c, "provide at least 4 characters of ticket id / QR code", nil)
		return
	}
	t, err := h.usecase.LookupTicket(c.Request.Context(), ref)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "ticket retrieved", t)
}

// canViewAllTickets: admin (kelola tiket) dan petugas (detail tiket) boleh
// melihat semua e-tiket. Peran lain hanya boleh melihat miliknya sendiri.
func canViewAllTickets(c *gin.Context) bool {
	return middleware.HasRole(c, "ADMIN") || middleware.HasRole(c, "PETUGAS")
}

func (h *Handler) ensureTicketAccess(c *gin.Context, bookingID string) bool {
	if canViewAllTickets(c) {
		return true
	}
	ok, err := h.usecase.CanAccess(c.Request.Context(), bookingID, middleware.CurrentUserID(c), false)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return false
	}
	if !ok {
		response.Forbidden(c, "Anda tidak memiliki akses.")
		return false
	}
	return true
}

// ListMine mengembalikan e-tiket milik pengguna yang sedang login.
func (h *Handler) ListMine(c *gin.Context) {
	tickets, err := h.usecase.ListMyTickets(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "tickets retrieved", tickets)
}

func (h *Handler) GetTicket(c *gin.Context) {
	t, err := h.usecase.GetTicket(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	if !h.ensureTicketAccess(c, t.BookingID) {
		return
	}
	response.OK(c, "ticket retrieved", t)
}

func (h *Handler) GetTicketByCode(c *gin.Context) {
	t, err := h.usecase.GetTicketByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	if !h.ensureTicketAccess(c, t.BookingID) {
		return
	}
	response.OK(c, "ticket retrieved", t)
}

func (h *Handler) GetTicketByBooking(c *gin.Context) {
	t, err := h.usecase.GetTicketByBooking(c.Request.Context(), c.Param("bookingId"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	if !h.ensureTicketAccess(c, t.BookingID) {
		return
	}
	response.OK(c, "ticket retrieved", t)
}

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *jwtutil.Manager) {
	requireAuth := middleware.RequireAuth(jwtManager)

	// PETUGAS: verifikasi (mutasi), hanya scan gerbang. Admin tidak memindai tiket.
	petugas := rg.Group("/petugas", requireAuth, middleware.RequireRole("PETUGAS"))
	petugas.POST("/scanner/validate", h.Validate)

	// Riwayat/detail scanner adalah area petugas. Manager mendapat data monitoring
	// agregat lewat halaman monitoring, bukan riwayat pemindaian mentah.
	petugasRead := rg.Group("/petugas", requireAuth, middleware.RequireRole("PETUGAS"))
	petugasRead.GET("/scanner/history", h.History)
	petugasRead.GET("/scanner/stats", h.Stats)
	petugasRead.GET("/scanner/ticket", h.LookupTicket)

	// Manager hanya menerima angka agregat untuk monitoring, bukan riwayat scan mentah.
	manager := rg.Group("/manager", requireAuth, middleware.RequireRole("MANAGER", "ADMIN", "super_admin"))
	manager.GET("/monitoring", h.ManagerMonitoring)

	// E-Tiket: pengunjung bisa melihat tiketnya sendiri, admin/petugas bisa melihat semua
	eTickets := rg.Group("/e-tickets", requireAuth)
	eTickets.GET("/mine", h.ListMine)
	eTickets.GET("/:id", h.GetTicket)
	eTickets.GET("/code/:code", h.GetTicketByCode)
	eTickets.GET("/booking/:bookingId", h.GetTicketByBooking)
}
