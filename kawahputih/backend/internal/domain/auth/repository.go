package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Repository adalah kontrak persistensi yang dipakai lapisan service (service bergantung pada interface, bukan *gorm.DB langsung).
type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	FindUserByID(ctx context.Context, id string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByGoogleID(ctx context.Context, googleID string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateUser(ctx context.Context, user *User) error

	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	FindValidRefreshToken(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, hash string) error
	RevokeAllRefreshTokens(ctx context.Context, userID string) error
	TouchLastLogin(ctx context.Context, userID string, at time.Time) error

	CreateAuthToken(ctx context.Context, token *AuthToken) error
	FindValidAuthToken(ctx context.Context, hash, purpose string) (*AuthToken, error)
	MarkAuthTokenUsed(ctx context.Context, id string) error
	InvalidatePendingAuthTokens(ctx context.Context, userID, purpose string) error
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) CreateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) FindUserByID(ctx context.Context, id string) (*User, error) {
	return r.findUser(ctx, "id = ?", id)
}

func (r *repository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	return r.findUser(ctx, "email = ?", email)
}

func (r *repository) FindUserByGoogleID(ctx context.Context, googleID string) (*User, error) {
	return r.findUser(ctx, "google_id = ?", googleID)
}

func (r *repository) findUser(ctx context.Context, where string, arg interface{}) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).First(&u, where, arg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *repository) UpdateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *repository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *repository) FindValidRefreshToken(ctx context.Context, hash string) (*RefreshToken, error) {
	var t RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ? AND revoked = ? AND expires_at > ?", hash, false, time.Now().Unix()).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) RevokeRefreshToken(ctx context.Context, hash string) error {
	return r.db.WithContext(ctx).Model(&RefreshToken{}).Where("token_hash = ?", hash).Update("revoked", true).Error
}

func (r *repository) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}

func (r *repository) TouchLastLogin(ctx context.Context, userID string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("last_login_at", at).Error
}

func (r *repository) CreateAuthToken(ctx context.Context, token *AuthToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *repository) FindValidAuthToken(ctx context.Context, hash, purpose string) (*AuthToken, error) {
	var t AuthToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND purpose = ? AND used_at IS NULL AND expires_at > ?", hash, purpose, time.Now().Unix()).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) MarkAuthTokenUsed(ctx context.Context, id string) error {
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Model(&AuthToken{}).Where("id = ?", id).Update("used_at", now).Error
}

func (r *repository) InvalidatePendingAuthTokens(ctx context.Context, userID, purpose string) error {
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Model(&AuthToken{}).
		Where("user_id = ? AND purpose = ? AND used_at IS NULL", userID, purpose).
		Update("used_at", now).Error
}
