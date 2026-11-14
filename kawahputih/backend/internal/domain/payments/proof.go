package payments

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Bukti pembayaran: hanya gambar (bukti transfer/screenshot) atau PDF.
const maxProofSize = 5 << 20 // 5 MB

var allowedProofExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".pdf": true,
}

// UploadProof menerima multipart field "proof" berisi bukti pembayaran, lalu
// menyimpannya di UPLOAD_DIR/payments dan menautkannya ke payment booking.
func (h *Handler) UploadProof(c *gin.Context) {
	bookingID := c.Param("bookingId")
	fh, err := c.FormFile("proof")
	if err != nil {
		response.BadRequest(c, "File bukti pembayaran wajib diunggah (field \"proof\").", nil)
		return
	}
	if fh.Size > maxProofSize {
		response.BadRequest(c, "Ukuran bukti pembayaran maksimal 5 MB.", nil)
		return
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedProofExt[ext] {
		response.BadRequest(c, "Format bukti tidak didukung. Gunakan JPG, PNG, WEBP, atau PDF.", nil)
		return
	}

	src, err := fh.Open()
	if err != nil {
		response.BadRequest(c, "Gagal membaca file bukti pembayaran.", nil)
		return
	}
	defer src.Close()

	// Sniff tipe dari byte awal agar ekstensi palsu tetap tertolak.
	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	head = head[:n]
	ct := http.DetectContentType(head)
	if ext != ".pdf" && !strings.HasPrefix(ct, "image/") {
		response.BadRequest(c, "Berkas bukti pembayaran bukan gambar yang valid.", nil)
		return
	}

	dir := filepath.Join(h.uploadDir, proofDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	name, err := randomFileName(ext)
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	target := filepath.Join(dir, name)
	out, err := os.Create(target)
	if err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	if _, err := io.Copy(out, io.MultiReader(strings.NewReader(string(head)), src)); err != nil {
		out.Close()
		_ = os.Remove(target)
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}
	if err := out.Close(); err != nil {
		apperror.WriteHTTP(c, apperror.Internal(err))
		return
	}

	proofURL := proofURLPrefix + "/" + name
	dto, err := h.usecase.AttachProof(c.Request.Context(), bookingID, middleware.CurrentUserID(c), privilegedRead(c), proofURL)
	if err != nil {
		_ = os.Remove(target)
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "bukti pembayaran diunggah", dto)
}

func randomFileName(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("payments: generate proof file name: %w", err)
	}
	return hex.EncodeToString(buf) + ext, nil
}
