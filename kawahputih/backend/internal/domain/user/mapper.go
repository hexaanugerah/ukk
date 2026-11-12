package user

func toResponse(u AdminUser, roles []string) AdminUserResponse {
	return AdminUserResponse{
		ID: u.ID, Name: u.Name, Email: u.Email, Phone: u.Phone,
		IsActive: u.IsActive, IsVerified: u.IsVerified, Roles: roles,
	}
}
