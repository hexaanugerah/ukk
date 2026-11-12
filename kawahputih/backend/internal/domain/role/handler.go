package role

import (
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
	roles, err := h.usecase.ListRoles(c.Request.Context())
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "roles retrieved", roles)
}

func (h *Handler) AssignToUser(c *gin.Context) {
	var req AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	if err := h.usecase.AssignRoles(c.Request.Context(), c.Param("userId"), req.Roles); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "roles updated", nil)
}
