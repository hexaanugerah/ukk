package user

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

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("limit"))
	users, total, err := h.usecase.List(c.Request.Context(), page, perPage, c.Query("search"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "users retrieved", users, response.BuildMeta(page, perPage, total))
}

func (h *Handler) SetActive(c *gin.Context) {
	var req SetActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if err := h.usecase.SetActive(c.Request.Context(), c.Param("id"), req.IsActive, middleware.CurrentUserID(c)); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "user status updated", nil)
}
