package user

type SetActiveRequest struct {
	IsActive bool `json:"is_active"`
}

type AdminUserResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Phone      string   `json:"phone"`
	IsActive   bool     `json:"is_active"`
	IsVerified bool     `json:"is_verified"`
	Roles      []string `json:"roles"`
}
