// KODE ASLI (potongan dari backend/internal/domain/auth/service.go).
// Snapshot utuh: Authenticate, CreateUser, RequestEmailVerification, VerifyEmail,
// ForgotPassword, ResetPassword.
// File ini hanya arsip/dokumentasi. Bukan bagian dari build, bukan kode yang dijalankan.
package auth_asli_snapshot

import (
	"context"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/mailer"
	"github.com/KawahPutih/backend/pkg/validation"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Placeholder tipe (replika, bukan definisi asli).
type Service struct {
	// isi replika; tidak dijalankan.
}

// Placeholder fungsi (replika, bukan definisi asli).
func (s *Service) issueToken(ctx context.Context, userID string, purpose string, ttl int) (string, error) {
	return "", nil
}

// Placeholder hashToken (replika).
func hashToken(token string) string {
	return ""
}

// Placeholder logger (replika).
var _ *zap.Logger

// ============================================================================
// KODE ASLI: Authenticate
// ============================================================================
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
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// ============================================================================
// KODE ASLI: CreateUser
// ============================================================================
func (s *Service) CreateUser(ctx context.Context, name, email, phone, password string) (*User, error) {
	// Normalisasi field identitas sebelum disimpan/dicek: email dicocokkan tanpa
	// membedakan huruf besar/kecil dan dipangkas, agar A@x.com tidak membuat akun
	// kedua di samping a@x.com.
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

// ============================================================================
// KODE ASLI: RequestEmailVerification
// ============================================================================
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
	_ = s.mailer.Send(ctx, user.Email, "Verifikasi Email - KPR Tourism", "Klik untuk verifikasi (berlaku 24 jam): "+link)
	return nil
}

// ============================================================================
// KODE ASLI: VerifyEmail
// ============================================================================
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

// ============================================================================
// KODE ASLI: ForgotPassword
// ============================================================================
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
	_ = s.mailer.Send(ctx, user.Email, "Reset Password - KPR Tourism", "Klik untuk reset (berlaku 1 jam): "+link)
	return nil
}

// ============================================================================
// KODE ASLI: ResetPassword
// ============================================================================
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
	return nil
}

// Placeholder konstanta (replika).
var (
	bcryptCost          = bcrypt.DefaultCost
	PurposeVerifyEmail  = "verify_email"
	PurposeResetPassword = "reset_password"
	verificationTokenTTL = 24 * 60 * 60
	resetTokenTTL        = 60 * 60
)

// Placeholder User struct (replika).
type User struct {
	ID           string
	Name         string
	Email        string
	Phone        string
	PasswordHash string
	IsActive     bool
	IsVerified   bool
}

// Placeholder konstanta error (replika).
var (
	ErrInvalidCredentials = apperror.New("invalid credentials")
	ErrEmailAlreadyExists = apperror.New("email already exists")
	ErrAccountDeactivated = apperror.New("account deactivated")
	ErrInvalidToken       = apperror.New("invalid token")
	ErrUserNotFound       = apperror.New("user not found")
)

// Placeholder repo interface (replika).
type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user *User) error
	FindUserByID(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	FindValidAuthToken(ctx context.Context, hash, purpose string) (*AuthToken, error)
	MarkAuthTokenUsed(ctx context.Context, id int64) error
}

// Placeholder AuthToken (replika).
type AuthToken struct {
	ID        int64
	UserID    string
	Purpose   string
	UsedAt    *int64
	ExpiresAt int64
}

// Placeholder mailer (replika).
var _ mailer.Mailer

// Placeholder frontend URL (replika).
var s = struct{ frontendURL string }{frontendURL: ""}
var _ = struct{ Send func(ctx context.Context, to, subject, body string) error }{Send: func(ctx context.Context, to, subject, body string) error { return nil }}
