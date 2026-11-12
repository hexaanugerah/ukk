package public

import (
	"strconv"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetDestinations(c *gin.Context) {
	items, err := h.service.GetDestinations(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "destinations retrieved", ToDestinationResponseList(items))
}

func (h *Handler) GetDestinationByID(c *gin.Context) {
	item, err := h.service.GetDestinationByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "destination retrieved", ToDestinationResponse(item))
}

func (h *Handler) GetDestinationBySlug(c *gin.Context) {
	item, err := h.service.GetDestinationBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "destination retrieved", ToDestinationResponse(item))
}

func (h *Handler) GetFacilities(c *gin.Context) {
	category := c.Query("category")
	items, err := h.service.GetFacilities(c.Request.Context(), category)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "facilities retrieved", ToFacilityResponseList(items))
}

func (h *Handler) GetFAQs(c *gin.Context) {
	category := c.Query("category")
	items, err := h.service.GetFAQs(c.Request.Context(), category)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "faqs retrieved", ToFAQResponseList(items))
}

func (h *Handler) GetReviews(c *gin.Context) {
	destinationID := c.Query("destination_id")
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 10
	}
	items, err := h.service.GetReviews(c.Request.Context(), destinationID, limit)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "reviews retrieved", ToReviewResponseList(items))
}

func (h *Handler) GetContents(c *gin.Context) {
	contentType := c.Query("type")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, total, err := h.service.GetContents(c.Request.Context(), contentType, page, limit)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "contents retrieved", ToContentResponseList(items), response.BuildMeta(page, limit, total))
}

func (h *Handler) GetContentBySlug(c *gin.Context) {
	item, err := h.service.GetContentBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "content retrieved", ToContentResponse(item))
}

func (h *Handler) GetContentByID(c *gin.Context) {
	item, err := h.service.GetContentByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "content retrieved", ToContentResponse(item))
}

func (h *Handler) GetPrices(c *gin.Context) {
	items, err := h.service.GetPrices(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "prices retrieved", items)
}
