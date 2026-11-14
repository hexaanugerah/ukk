package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/KawahPutih/backend/internal/config"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/mailer"
	"github.com/KawahPutih/backend/pkg/validation"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
)

const bcryptCost = 12
const (
	verificationTokenTTL = 24 * time.Hour
	resetTokenTTL        = 1 * time.Hour
)

// Service memegang semua aturan bisnis auth: hashing, validasi kredensial, penerbitan/rotasi token, dan OAuth. Tidak menyentuh gin.Context dan tidak mengoordinasi domain lain.
type Service struct {
	repo        Repository
	jwtManager  *jwtutil.Manager
	mailer      mailer.Mailer
	oauthCfg    config.GoogleOAuthConfig
	frontendURL string
	log         *zap.Logger
}

func NewService(repo Repository, jwtManager *jwtutil.Manager, m mailer.Mailer, oauthCfg config.GoogleOAuthConfig, frontendURL string, log *zap.Logger) *Service {
	return &Service{repo: repo, jwtManager: jwtManager, mailer: m, oauthCfg: oauthCfg, frontendURL: frontendURL, log: log}
}

func (s *Service) CreateUser(ctx context.Context, name, email, phone, password string) (*User, error) {
	// Normalisasi field identitas sebelum disimpan/dicek: email dicocokkan tanpa membedakan huruf besar/kecil dan dipangkas, agar A@x.com tidak membuat akun kedua di samping a@x.com.
	name = validation.SanitizeInput(name)
	email = validation.SanitizeEmail(email)
	phone = validation.SanitizePhone(phone)
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	user := &User{Name: name, Email: email, Phone: phone, PasswordHash: string(hash), IsActive: true}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, apperror.Internal(err)
	}
	return user, nil
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	// Normalisasi sama seperti CreateUser, agar login tidak membedakan huruf besar/kecil.
	email = validation.SanitizeEmail(email)
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrAccountDeactivated
	}
	// AUTH-06: akun yang belum verifikasi email tidak boleh login. Akun Google
	// selalu IsVerified=true (lihat GoogleExchange), jadi hanya alur password
	// yang perlu gerbang ini.
	if !user.IsVerified {
		return nil, ErrEmailNotVerified
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) IssueTokenPair(ctx context.Context, user *User, roles []string, userAgent, ip string) (*AuthResponse, error) {
	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, roles)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, roles)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	rt := &RefreshToken{UserID: user.ID, TokenHash: hashToken(refreshToken), ExpiresAt: refreshExpiresAt.Unix(), UserAgent: userAgent, IPAddress: ip}
	if err := s.repo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, apperror.Internal(err)
	}
	return &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresAt: expiresAt.Unix(), User: toUserResponse(user, roles)}, nil
}

// RecordLogin mengisi kolom operasional last_login_at. Kegagalan hanya dicatat, tidak dikembalikan, karena gagal tulis timestamp tidak boleh menghalangi autentikasi.
func (s *Service) RecordLogin(ctx context.Context, userID string) {
	now := time.Now()
	if err := s.repo.TouchLastLogin(ctx, userID, now); err != nil {
		s.log.Error("record last login failed", zap.String("user_id", userID), zap.Error(err))
	}
}

// GetProfile mengembalikan profil pemanggil. Role diambil dari lapisan usecase (domain role) sehingga service bebas panggilan lintas domain.
func (s *Service) GetProfile(ctx context.Context, userID string) (*User, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile menerapkan perubahan profil pemanggil. Perubahan email dicek keunikannya dan wajib verifikasi ulang.
func (s *Service) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*User, error) {
	user, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Email != "" && !strings.EqualFold(req.Email, user.Email) {
		newEmail := validation.SanitizeEmail(req.Email)
		exists, err := s.repo.ExistsByEmail(ctx, newEmail)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		if exists {
			return nil, ErrEmailAlreadyExists
		}
		user.Email = newEmail
		user.IsVerified = false
	}
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, apperror.Internal(err)
	}
	return user, nil
}

// ChangePassword mengganti password pengguna yang sedang login (berlaku untuk semua role). Password baru selalu di-hash ulang dengan bcrypt; hash lama tidak pernah dipakai kembali, dan seluruh refresh token dicabut agar sesi lain tidak tetap hidup dengan kredensial lama.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return apperror.Internal(err)
	}
	if user == nil {
		return ErrUserNotFound
	}
	// Akun Google-only belum punya password; jangan izinkan ganti password tanpa verifikasi apa pun.
	if user.PasswordHash == "" {
		return ErrPasswordNotSet
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)) != nil {
		return ErrCurrentPasswordWrong
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newPassword)) == nil {
		return ErrPasswordReused
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return apperror.Internal(err)
	}
	user.PasswordHash = string(hash)
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return apperror.Internal(err)
	}
	// Password berubah = semua sesi lama tidak boleh dipakai lagi.
	if err := s.repo.RevokeAllRefreshTokens(ctx, user.ID); err != nil {
		s.log.Error("revoke refresh tokens after password change failed", zap.String("user_id", user.ID), zap.Error(err))
	}
	return nil
}

func (s *Service) RefreshTokenPair(ctx context.Context, rawRefreshToken string) (*User, string, string, error) {
	claims, err := s.jwtManager.ParseRefreshToken(rawRefreshToken)
	if err != nil {
		return nil, "", "", ErrInvalidToken
	}
	hash := hashToken(rawRefreshToken)
	stored, err := s.repo.FindValidRefreshToken(ctx, hash)
	if err != nil {
		return nil, "", "", apperror.Internal(err)
	}
	if stored == nil {
		return nil, "", "", ErrTokenRevoked
	}
	user, err := s.repo.FindUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, "", "", apperror.Internal(err)
	}
	if user == nil || !user.IsActive {
		return nil, "", "", ErrAccountDeactivated
	}
	if err := s.repo.RevokeRefreshToken(ctx, hash); err != nil {
		s.log.Error("revoke old refresh token failed", zap.Error(err))
	}
	return user, stored.UserAgent, stored.IPAddress, nil
}

func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	if err := s.repo.RevokeRefreshToken(ctx, hashToken(rawRefreshToken)); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

// Verifikasi email

func (s *Service) RequestEmailVerification(ctx context.Context, email string) error {
	email = validation.SanitizeEmail(email)
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return apperror.Internal(err)
	}
	if user == nil || user.IsVerified {
		return nil // never reveal whether the email exists
	}
	raw, err := s.issueToken(ctx, user.ID, PurposeVerifyEmail, verificationTokenTTL)
	if err != nil {
		return err
	}
	link := s.frontendURL + "/verify-email?token=" + raw
	_ = s.mailer.Send(ctx, user.Email, "Verifikasi Email ", "Klik untuk verifikasi (berlaku 24 jam): "+link)
	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	token, err := s.repo.FindValidAuthToken(ctx, hashToken(rawToken), PurposeVerifyEmail)
	if err != nil {
		return apperror.Internal(err)
	}
	if token == nil {
		return ErrInvalidToken
	}
	user, err := s.repo.FindUserByID(ctx, token.UserID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}
	user.IsVerified = true
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return apperror.Internal(err)
	}
	_ = s.repo.MarkAuthTokenUsed(ctx, token.ID)
	return nil
}

// Reset password

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = validation.SanitizeEmail(email)
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return apperror.Internal(err)
	}
	if user == nil {
		return nil
	}
	raw, err := s.issueToken(ctx, user.ID, PurposeResetPassword, resetTokenTTL)
	if err != nil {
		return err
	}
	link := s.frontendURL + "/reset-password?token=" + raw
	_ = s.mailer.Send(ctx, user.Email, "Reset Password ", "Klik untuk reset (berlaku 1 jam): "+link)
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	token, err := s.repo.FindValidAuthToken(ctx, hashToken(rawToken), PurposeResetPassword)
	if err != nil {
		return apperror.Internal(err)
	}
	if token == nil {
		return ErrInvalidToken
	}
	user, err := s.repo.FindUserByID(ctx, token.UserID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return apperror.Internal(err)
	}
	user.PasswordHash = string(hash)
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return apperror.Internal(err)
	}
	_ = s.repo.MarkAuthTokenUsed(ctx, token.ID)
	_ = s.repo.RevokeAllRefreshTokens(ctx, user.ID) // reset kills every existing session
	return nil
}

func (s *Service) issueToken(ctx context.Context, userID, purpose string, ttl time.Duration) (string, error) {
	_ = s.repo.InvalidatePendingAuthTokens(ctx, userID, purpose)
	raw, err := randomHex(32)
	if err != nil {
		return "", apperror.Internal(err)
	}
	record := &AuthToken{UserID: userID, Purpose: purpose, TokenHash: hashToken(raw), ExpiresAt: time.Now().Add(ttl).Unix()}
	if err := s.repo.CreateAuthToken(ctx, record); err != nil {
		return "", apperror.Internal(err)
	}
	return raw, nil
}

// Google OAuth

func (s *Service) googleConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID: s.oauthCfg.ClientID, ClientSecret: s.oauthCfg.ClientSecret, RedirectURL: s.oauthCfg.RedirectURL,
		Scopes:   []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: googleoauth.Endpoint,
	}
}

func (s *Service) GoogleAuthURL(state string) (string, error) {
	if s.oauthCfg.ClientID == "" {
		return "", ErrOAuthNotConfigured
	}
	return s.googleConfig().AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GoogleExchange menukar kode OAuth menjadi data user, membuat atau menautkan sesuai kebutuhan, TANPA menerbitkan token. Penerbitan token adalah tugas usecase.
func (s *Service) GoogleExchange(ctx context.Context, code string) (*User, error) {
	if s.oauthCfg.ClientID == "" {
		return nil, ErrOAuthNotConfigured
	}
	cfg := s.googleConfig()
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, apperror.Unauthorized("Gagal memproses login Google. Silakan coba kembali.")
	}
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, apperror.Internal(err)
	}
	defer resp.Body.Close()
	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, apperror.Internal(err)
	}
	if !info.VerifiedEmail {
		return nil, ErrOAuthEmailUnverified
	}

	user, err := s.repo.FindUserByGoogleID(ctx, info.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if user != nil {
		if !user.IsActive {
			return nil, ErrAccountDeactivated
		}
		return user, nil
	}

	existing, err := s.repo.FindUserByEmail(ctx, info.Email)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing != nil {
		existing.GoogleID = &info.ID
		if existing.AvatarURL == "" {
			existing.AvatarURL = info.Picture
		}
		existing.IsVerified = true
		if err := s.repo.UpdateUser(ctx, existing); err != nil {
			return nil, apperror.Internal(err)
		}
		return existing, nil
	}

	newUser := &User{Name: info.Name, Email: info.Email, GoogleID: &info.ID, AvatarURL: info.Picture, IsActive: true, IsVerified: true}
	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		return nil, apperror.Internal(err)
	}
	return newUser, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
