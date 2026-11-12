package gallery

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

	if req.ImageFile != "" && !isValidImageFileRef(req.ImageFile) {
		response.BadRequest(c, "image_file tidak valid", nil)
		return
	}

	result, err := h.usecase.Create(c.Request.Context(), middleware.CurrentUserID(c), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "gallery item created", result)
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
	if req.ImageFile != "" && !isValidImageFileRef(req.ImageFile) {
		response.BadRequest(c, "image_file tidak valid", nil)
		return
	}

	result, err := h.usecase.Update(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c), req)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "gallery item updated", result)
}

func (h *Handler) Get(c *gin.Context) {
	result, err := h.usecase.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "gallery item retrieved", result)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.usecase.Delete(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c)); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "gallery item deleted", nil)
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("limit"))
	result, total, err := h.usecase.List(c.Request.Context(), page, perPage, c.Query("category"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "gallery items retrieved", result, response.BuildMeta(page, perPage, total))
}

func (h *Handler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("limit"))
	result, total, err := h.usecase.Search(c.Request.Context(), c.Query("q"), page, perPage)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "gallery items retrieved", result, response.BuildMeta(page, perPage, total))
}
