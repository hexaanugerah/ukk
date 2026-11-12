package auth

import (
	"context"
	"testing"
	"time"

	"github.com/KawahPutih/backend/internal/config"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupAuthTest menyiapkan repository + service auth di atas SQLite in-memory,
// sehingga hash password benar-benar diuji lewat jalur produksi (Service -> Repository).
func setupAuthTest(t *testing.T) (*Service, Repository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	// SQLite in-memory menghilangkan isi database saat koneksi berbeda dibuka;
	// satu koneksi membuat seluruh tes memakai database yang sama.
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(&User{}, &RefreshToken{}, &AuthToken{}))

	repo := NewRepository(db)
	service := NewService(repo, nil, nil, config.GoogleOAuthConfig{}, "http://localhost:3000", zap.NewNop())
	return service, repo
}

// TestCreateUserHashesPassword memastikan password tidak pernah tersimpan sebagai teks biasa.
func TestCreateUserHashesPassword(t *testing.T) {
	service, repo := setupAuthTest(t)
	ctx := context.Background()

	user, err := service.CreateUser(ctx, "Pengunjung", "pengunjung@example.com", "081200000001", "RahasiaKu123")
	require.NoError(t, err)

	stored, err := repo.FindUserByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.NotEqual(t, "RahasiaKu123", stored.PasswordHash, "password must never be stored in plain text")
	assert.NotContains(t, stored.PasswordHash, "RahasiaKu123")

	cost, err := bcrypt.Cost([]byte(stored.PasswordHash))
	require.NoError(t, err)
	assert.Equal(t, bcryptCost, cost, "stored hash must use the project's bcrypt cost")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("RahasiaKu123")))
}

// TestChangePassword mengenali semua role karena autentikasi memakai tabel users yang sama;
// yang diuji adalah alur ganti password: verifikasi password lama, hash ulang, sesi dicabut.
func TestChangePassword(t *testing.T) {
	service, repo := setupAuthTest(t)
	ctx := context.Background()

	user, err := service.CreateUser(ctx, "Petugas", "petugas@example.com", "081200000002", "PasswordLama123")
	require.NoError(t, err)

	refresh := &RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken("refresh-token-uji"),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	require.NoError(t, repo.CreateRefreshToken(ctx, refresh))

	oldHash := user.PasswordHash

	// Password lama salah: tidak boleh mengubah apa pun.
	err = service.ChangePassword(ctx, user.ID, "PasswordSalah123", "PasswordBaru123")
	assert.ErrorIs(t, err, ErrCurrentPasswordWrong)
	unchanged, err := repo.FindUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, oldHash, unchanged.PasswordHash)

	// Password baru sama dengan yang lama ditolak.
	err = service.ChangePassword(ctx, user.ID, "PasswordLama123", "PasswordLama123")
	assert.ErrorIs(t, err, ErrPasswordReused)

	// Ganti password berhasil.
	require.NoError(t, service.ChangePassword(ctx, user.ID, "PasswordLama123", "PasswordBaru123"))

	updated, err := repo.FindUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.NotEqual(t, oldHash, updated.PasswordHash, "new password must be re-hashed with a fresh salt")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("PasswordBaru123")))

	// Kredensial lama tidak berlaku, kredensial baru berlaku.
	_, err = service.Authenticate(ctx, "petugas@example.com", "PasswordLama123")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	_, err = service.Authenticate(ctx, "petugas@example.com", "PasswordBaru123")
	assert.NoError(t, err)

	// Semua refresh token lama dicabut sehingga sesi di perangkat lain berakhir.
	stored, err := repo.FindValidRefreshToken(ctx, hashToken("refresh-token-uji"))
	require.NoError(t, err)
	assert.Nil(t, stored, "refresh tokens must be revoked after a password change")
}

// TestChangePasswordGoogleOnlyAccount memastikan akun tanpa password tidak bisa mengganti password
// hanya dengan menebak password lama yang kosong.
func TestChangePasswordGoogleOnlyAccount(t *testing.T) {
	service, repo := setupAuthTest(t)
	ctx := context.Background()

	googleID := "google-123"
	user := &User{Name: "Akun Google", Email: "google@example.com", GoogleID: &googleID, IsActive: true, IsVerified: true}
	require.NoError(t, repo.CreateUser(ctx, user))

	err := service.ChangePassword(ctx, user.ID, "", "PasswordBaru123")
	assert.ErrorIs(t, err, ErrPasswordNotSet)
}
