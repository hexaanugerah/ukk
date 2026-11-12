package role

type AssignRolesRequest struct {
	Roles []string `json:"roles" validate:"required,min=1,dive,required"`
}

type RoleResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions,omitempty"`
}
