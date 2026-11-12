package auth

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=150"`
	Email    string `json:"email" validate:"required,email,max=150"`
	Phone    string `json:"phone" validate:"omitempty,max=20"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RequestEmailVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

// ChangePasswordRequest dipakai pengguna yang sudah login dari role mana pun.
// Password lama hanya wajib diisi (tanpa batas panjang) agar akun lama dengan
// password pendek tetap bisa diganti; kekuatan password baru divalidasi terpisah.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    int64        `json:"expires_at"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Role       string `json:"role"`
	Gate       string `json:"gate,omitempty"`
	Shift      string `json:"shift,omitempty"`
	IsVerified bool   `json:"is_verified"`
}

type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"omitempty,min=3,max=150"`
	Email string `json:"email" validate:"omitempty,email,max=150"`
	Phone string `json:"phone" validate:"omitempty,max=20"`
}
