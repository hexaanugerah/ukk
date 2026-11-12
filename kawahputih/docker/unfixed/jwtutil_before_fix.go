// KODE ASLI (kondisi SEBELUM FIX).
// Snapshot dari backend/pkg/jwtutil/jwtutil.go: versi sebelum perbaikan
// terkait bug TC25 (rotasi refresh token tidak tepat — token identik & revoke tidak memutus sesi).
// File ini hanya dokumen/arsip, bukan bagian dari build.
package jwtutil_snapshot

// generate — kondisi SEBELUM FIX: tanpa jti (ID unik).
// Karena IssuedAt menggunakan time.Now() dan JWT HS256 saja, dua panggilan dalam satu detik
// menghasilkan token refresh identik sehingga endpoint refresh/reuse bisa lolos revoke.
func (m *Manager) generate(userID, email string, roles []string, secret []byte, ttl time.Duration) (string, time.Time, error) {
    expiresAt := time.Now().Add(ttl)
    claims := Claims{
        UserID: userID, Email: email, Roles: roles,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    m.issuer,
            IssuedAt:  jwt.NewNumericDate(time.Now()), // <-- tanpa jti; dalam 1 detik token identik
            ExpiresAt: jwt.NewNumericDate(expiresAt),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(secret)
    return signed, expiresAt, err
}

// RefreshTokenPair — kondisi SEBELUM FIX:
// revoke dipanggil, namun jika token identik (duplikat hash) row duplikat tetap valid
// mengingat tidak ada jti/kunci unik yang dicek.
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
        // kondisi SEBELUM FIX: jika token identik / row duplikat, revoke tidak memutus sesi
        s.log.Error("revoke old refresh token failed", zap.Error(err))
    }
    return user, stored.UserAgent, stored.IPAddress, nil
}
