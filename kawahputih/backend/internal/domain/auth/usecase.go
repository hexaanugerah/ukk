package auth

import (
	"context"

	"go.uber.org/zap"
)

// RoleProvider adalah cara usecase auth menjangkau domain role tanpa mengimpornya langsung: auth mendefinisikan interface yang dibutuhkan, service role memenuhinya secara struktural, dan bootstrap memasang instance konkretnya.
type RoleProvider interface {
	AssignDefaultRole(ctx context.Context, userID string) error
	GetRoleNames(ctx context.Context, userID string) ([]string, error)
}

// Usecase menghubungkan handler ke service dan satu-satunya lapisan yang boleh mengoordinasi lintas domain. Service di atas tidak pernah memanggil RoleProvider, hanya Usecase.
type Usecase struct {
	service *Service
	roles   RoleProvider
}

func NewUsecase(service *Service, roles RoleProvider) *Usecase {
	return &Usecase{service: service, roles: roles}
}

func (u *Usecase) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	user, err := u.service.CreateUser(ctx, req.Name, req.Email, req.Phone, req.Password)
	if err != nil {
		return nil, err
	}
	if err := u.roles.AssignDefaultRole(ctx, user.ID); err != nil {
		return nil, err
	}
	// AUTH-07: registrasi otomatis mengirim email verifikasi. Kegagalan kirim
	// email tidak boleh menggagalkan registrasi — pengguna masih bisa meminta
	// ulang lewat POST /auth/email/verify/request.
	if err := u.service.RequestEmailVerification(ctx, user.Email); err != nil {
		u.service.log.Warn("send verification email after register failed", zap.String("email", user.Email), zap.Error(err))
	}
	roleNames, err := u.roles.GetRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return u.service.IssueTokenPair(ctx, user, roleNames, "", "")
}

func (u *Usecase) Login(ctx context.Context, req LoginRequest, userAgent, ip string) (*AuthResponse, error) {
	user, err := u.service.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	roleNames, err := u.roles.GetRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	u.service.RecordLogin(ctx, user.ID)
	return u.service.IssueTokenPair(ctx, user, roleNames, userAgent, ip)
}

// Me menyusun payload /auth/me: profil + role efektif. Dipakai frontend untuk cek sesi saat halaman dimuat.
func (u *Usecase) Me(ctx context.Context, userID string) (*UserResponse, error) {
	user, err := u.service.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleNames, err := u.roles.GetRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	resp := toUserResponse(user, roleNames)
	return &resp, nil
}

// UpdateProfile memperbarui profil pemanggil dan mengembalikan payload profil terbaru.
func (u *Usecase) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*UserResponse, error) {
	user, err := u.service.UpdateProfile(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	roleNames, err := u.roles.GetRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	resp := toUserResponse(user, roleNames)
	return &resp, nil
}

func (u *Usecase) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
	return u.service.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword)
}

func (u *Usecase) Refresh(ctx context.Context, rawRefreshToken string) (*AuthResponse, error) {
	user, userAgent, ip, err := u.service.RefreshTokenPair(ctx, rawRefreshToken)
	if err != nil {
		return nil, err
	}
	roleNames, err := u.roles.GetRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return u.service.IssueTokenPair(ctx, user, roleNames, userAgent, ip)
}

func (u *Usecase) Logout(ctx context.Context, rawRefreshToken string) error {
	return u.service.Logout(ctx, rawRefreshToken)
}

func (u *Usecase) RequestEmailVerification(ctx context.Context, email string) error {
	return u.service.RequestEmailVerification(ctx, email)
}

func (u *Usecase) VerifyEmail(ctx context.Context, token string) error {
	return u.service.VerifyEmail(ctx, token)
}

func (u *Usecase) ForgotPassword(ctx context.Context, email string) error {
	return u.service.ForgotPassword(ctx, email)
}

func (u *Usecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	return u.service.ResetPassword(ctx, token, newPassword)
}

func (u *Usecase) GoogleAuthURL(state string) (string, error) {
	return u.service.GoogleAuthURL(state)
}

func (u *Usecase) GoogleLogin(ctx context.Context, code string) (*AuthResponse, error) {
	user, err := u.service.GoogleExchange(ctx, code)
	if err != nil {
		return nil, err
	}
	if err := u.roles.AssignDefaultRole(ctx, user.ID); err != nil {
		// no-op bila role sudah ada, lihat AssignDefaultRole di domain role
		return nil, err
	}
	roleNames, err := u.roles.GetRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return u.service.IssueTokenPair(ctx, user, roleNames, "", "")
}
