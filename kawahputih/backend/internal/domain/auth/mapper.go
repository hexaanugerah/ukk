package auth

// mapper.go satu-satunya tempat penerjemahan entity GORM (User) ke DTO (UserResponse); handler hanya melihat DTO.
// role adalah role efektif tunggal pengguna: role staf tertinggi bila ada (ADMIN/super_admin > MANAGER > PETUGAS), selain itu PENGUNJUNG. JWT tetap membawa semua role untuk dicek middleware.
func toUserResponse(u *User, roles []string) UserResponse {
	role := rolePENGUNJUNG
	for _, r := range roles {
		switch r {
		case "ADMIN", "admin", "super_admin", "SUPER_ADMIN":
			// super_admin diterjemahkan ADMIN agar panel frontend (route guard
			// level ADMIN) mengikuti cakupan akses backend RequireRole("ADMIN", "super_admin").
			role = "ADMIN"
		case "MANAGER", "manager":
			if role != "ADMIN" {
				role = "MANAGER"
			}
		case "PETUGAS", "petugas":
			if role != "ADMIN" && role != "MANAGER" {
				role = "PETUGAS"
			}
		}
	}
	return UserResponse{
		ID: u.ID, Name: u.Name, Email: u.Email, Phone: u.Phone,
		Role: role, Gate: u.Gate, Shift: u.Shift, IsVerified: u.IsVerified,
	}
}

const rolePENGUNJUNG = "PENGUNJUNG"
