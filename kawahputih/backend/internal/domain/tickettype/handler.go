package tickettype

import (
	"context"
	"strconv"
	"time"

	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// SalesCounter memungkinkan domain tickettype bertanya berapa tiap jenis tiket yang sudah terjual untuk tanggal X, tanpa mengimpor domain booking. Bootstrap memasang adapter service booking; bila nil, available sama dengan kapasitas.
type SalesCounter interface {
	SoldByTicketOn(ctx context.Context, date string) map[string]int64
}

type Handler struct {
	service *Service
	sales   SalesCounter
}

func NewHandler(service *Service, sales SalesCounter) *Handler {
	return &Handler{service: service, sales: sales}
}

// withAvailability menambahkan `available` = kapasitas dikurangi tiket yang sudah terjual hari ini. Validasi booking di server tetap yang menentukan untuk tanggal tertentu.
func (h *Handler) withAvailability(c *gin.Context, items []Response) []Response {
	if h.sales == nil || len(items) == 0 {
		return items
	}
	today := time.Now().Format("2006-01-02")
	sold := h.sales.SoldByTicketOn(c.Request.Context(), today)
	for i := range items {
		if s, ok := sold[items[i].ID]; ok {
			avail := items[i].Capacity - int(s)
			if avail < 0 {
				avail = 0
			}
			items[i].Available = avail
		} else {
			items[i].Available = items[i].Capacity
		}
	}
	return items
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	// Frontend juga mengirim per_page (ticket service); terima keduanya.
	if limit == 0 {
		if pp, _ := strconv.Atoi(c.Query("per_page")); pp > 0 {
			limit = pp
		}
	}
	limit, offset := clampPage(page, limit)
	items, total, err := h.service.List(c.Request.Context(), offset, limit, c.Query("status"), c.Query("category"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	responses := toResponseList(items)
	response.OKWithMeta(c, "tickets retrieved", h.withAvailability(c, responses), response.BuildMeta(page, limit, total))
}

func (h *Handler) ListActive(c *gin.Context) {
	items, err := h.service.ListActive(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "active tickets retrieved", h.withAvailability(c, toResponseList(items)))
}

func (h *Handler) Categories(c *gin.Context) {
	cats, err := h.service.Categories(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "ticket categories retrieved", cats)
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	resp := toResponseList([]TicketType{*item})
	response.OK(c, "ticket retrieved", h.withAvailability(c, resp)[0])
}

func (h *Handler) GetByCode(c *gin.Context) {
	item, err := h.service.GetByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	resp := toResponseList([]TicketType{*item})
	response.OK(c, "ticket retrieved", h.withAvailability(c, resp)[0])
}

func (h *Handler) Create(c *gin.Context) {
	var req UpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	item, err := h.service.Create(c.Request.Context(), middleware.CurrentUserID(c), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "ticket type created", toResponse(item))
}

func (h *Handler) Update(c *gin.Context) {
	var req UpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	item, err := h.service.Update(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "ticket type updated", toResponse(item))
}

func (h *Handler) SetStatus(c *gin.Context) {
	var req SetStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	item, err := h.service.SetStatus(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c), req.Status)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "ticket type status updated", toResponse(item))
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "ticket type deleted", nil)
}

func clampPage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return limit, (page - 1) * limit
}
