package visitorrule

import (
	"strconv"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase *Usecase
}

func NewHandler(usecase *Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	result, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "visitor rule created", ToResponse(result))
}

func (h *Handler) Update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	result, err := h.usecase.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "visitor rule updated", ToResponse(result))
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.usecase.Delete(c.Request.Context(), c.Param("id")); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "visitor rule deleted", nil)
}

func (h *Handler) Get(c *gin.Context) {
	result, err := h.usecase.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "visitor rule retrieved", ToResponse(result))
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("limit"))
	category := c.Query("category")
	publishedOnly := c.Query("published_only") == "true"
	result, total, err := h.usecase.List(c.Request.Context(), page, perPage, category, publishedOnly)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "visitor rules retrieved", ToResponseList(result), response.BuildMeta(page, perPage, total))
}

func (h *Handler) ListPublished(c *gin.Context) {
	category := c.Query("category")
	result, err := h.usecase.ListPublished(c.Request.Context(), category)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "published visitor rules retrieved", ToResponseList(result))
}
