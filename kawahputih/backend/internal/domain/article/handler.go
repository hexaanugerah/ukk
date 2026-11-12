package article

import (
	"strconv"

	"github.com/KawahPutih/backend/internal/middleware"
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
	result, err := h.usecase.Create(c.Request.Context(), middleware.CurrentUserID(c), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "article created as draft", result)
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
	result, err := h.usecase.Update(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "article updated", result)
}

func (h *Handler) Publish(c *gin.Context) {
	if err := h.usecase.Publish(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c)); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "article published", nil)
}

func (h *Handler) Archive(c *gin.Context) {
	if err := h.usecase.Archive(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c)); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "article archived", nil)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.usecase.Delete(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c)); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "article deleted", nil)
}

func (h *Handler) Get(c *gin.Context) {
	result, err := h.usecase.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "article retrieved", result)
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("limit"))
	result, total, err := h.usecase.List(c.Request.Context(), page, perPage, c.Query("status"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "articles retrieved", result, response.BuildMeta(page, perPage, total))
}
