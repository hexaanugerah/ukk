// KODE FIX (setelah diperbaiki).
// Snapshot dari backend/pkg/jwtutil/jwtutil.go: versi SETELAH perbaikan
// terkait bug TC25 (rotasi refresh token tidak tepat — token identik & revoke tidak memutus sesi).
// File ini hanya dokumen/arsip, bukan bagian dari build.
package jwtutil_fix_snapshot

// generate — SETELAH FIX: tambahkan jti (ID unik) untuk mencegah token identik.
// Kini setiap token refresh memiliki jti yang unik sehingga revoke bisa memutus sesi.
func (m *Manager) generate(userID, email string, roles []string, secret []byte, ttl time.Duration) (string, time.Time, error) {
    expiresAt := time.Now().Add(ttl)
    claims := Claims{
        UserID: userID, Email: email, Roles: roles,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    m.issuer,
            ID:        m.randomJTI(),           // <-- SETELAH FIX: tambahkan jti unik
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            ExpiresAt: jwt.NewNumericDate(expiresAt),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(secret)
    return signed, expiresAt, err
}

// generate (versi sebelum fix) TIDAK memiliki jti.
// Perbedaannya: sebelum fix, dalam satu detik dua panggilan menghasilkan token identik (tanpa jti).

// helper pembuatan jti (contoh placeholder — implementasi sesungguhnya bisa pakai crypto/rand atau UUID)
func (m *Manager) randomJTI() string {
    // (implementasi placeholder; sesungguhnya pakai crypto/rand atau uuid)
    return fmt.Sprintf("jti-%d", time.Now().UnixNano())
}

// RefreshTokenPair — SETELAH FIX:
// revoke dipanggil, dan karena sekarang ada jti, revoke bisa memutus sesi dengan tepat.
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
    // --- SETELAH FIX: revoke tepat karena jti unik ---
    if err := s.repo.RevokeRefreshToken(ctx, hash); err != nil {
        s.log.Error("revoke old refresh token failed", zap.Error(err))
    }
    return user, stored.UserAgent, stored.IPAddress, nil
}

// RefreshTokenPair (versi sebelum fix) memiliki masalah revoke tidak memutus sesi jika token identik.
// Perbedaannya: sebelum fix, revoke dipanggil tapi token identik (tanpa jti) menyebabkan revoke tidak memutus sesi.
