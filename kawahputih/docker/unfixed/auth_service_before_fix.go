// KODE ASLI (kondisi SEBELUM FIX).
// Snapshot dari backend/internal/domain/auth/service.go: versi sebelum perbaikan
// terkait bug TC06 (login tanpa cek IsVerified) & TC07 (register tanpa kirim email verifikasi).
// File ini hanya dokumen/arsip, bukan bagian dari build. Dipakai untuk dokumentasi Sprint 3.
package auth_snapshot

// From internal/domain/auth/service.go — Authenticate (status quo sebelum fix)
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
    // --- kondisi SEBELUM FIX: tidak ada pemeriksaan user.IsVerified di sini ---
    return user, nil
}

// CreateUser — kondisi SEBELUM FIX: register TIDAK otomatis mengirim email verifikasi.
// (email verifikasi hanya bisa dipicu manual lewat endpoint /auth/email/verify/request)
func (s *Service) CreateUser(ctx context.Context, name, email, phone, password string) (*User, error) {
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
    user := &User{
        Name: name, Email: email, Phone: phone,
        PasswordHash: string(hash), IsActive: true,
        // --- kondisi SEBELUM FIX: field verifikasi/IsVerified belum diatur, dan
        //     tidak ada s.RequestEmailVerification(...) yang dipicu di sini ---
    }
    if err := s.repo.CreateUser(ctx, user); err != nil {
        return nil, apperror.Internal(err)
    }
    return user, nil
}
