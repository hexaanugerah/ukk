package gallery

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/KawahPutih/backend/internal/middleware"
	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase *Usecase
}

func NewHandler(usecase *Usecase) *Handler {
	return &Handler{usecase: usecase}
}

// parseImageField membaca berkas gambar dari form multipart dan memvalidasi
// isinya (sniff content type dari byte awal, bukan sekadar header — klien
// tidak selalu mengirim Content-Type yang benar). required=false hanya boleh
// dipakai pada pembaruan (gambar lama dipakai).
func parseImageField(c *gin.Context, required bool) (*ImageUpload, bool) {
	fh, err := c.FormFile("image")
	if err != nil {
		if required {
			response.BadRequest(c, "File gambar wajib diunggah (field \"image\").", nil)
			return nil, false
		}
		return nil, true
	}
	if fh.Size > MaxImageSize {
		response.BadRequest(c, "Ukuran file gambar maksimal 5 MB.", nil)
		return nil, false
	}
	ext := filepath.Ext(fh.Filename)
	if ext == "" || !ValidateImageExt(ext) {
		response.BadRequest(c, "Format gambar tidak didukung. Gunakan JPG, PNG, WEBP, atau GIF.", nil)
		return nil, false
	}
	src, err := fh.Open()
	if err != nil {
		response.BadRequest(c, "Gagal membaca file gambar.", nil)
		return nil, false
	}

	// Sniff tipe dari 512 byte pertama agar ekstensi palsu (mis. .png untuk
	// berkas eksekusi) tetap tertolak.
	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	head = head[:n]
	if ct := http.DetectContentType(head); !strings.HasPrefix(ct, "image/") {
		src.Close()
		response.BadRequest(c, "Berkas yang diunggah bukan gambar.", nil)
		return nil, false
	}
	return &ImageUpload{Ext: strings.ToLower(ext), Reader: io.MultiReader(bytes.NewReader(head), src)}, true
}

func (h *Handler) Create(c *gin.Context) {
	req := CreateRequest{
		Title:     c.PostForm("title"),
		Category:  c.PostForm("category"),
		SortOrder: atoiOrZero(c.PostForm("sort_order")),
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	image, ok := parseImageField(c, true)
	if !ok {
		return
	}

	result, err := h.usecase.Create(c.Request.Context(), middleware.CurrentUserID(c), req, image)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.Created(c, "gallery item created", result)
}

func (h *Handler) Update(c *gin.Context) {
	req := UpdateRequest{
		Title:     c.PostForm("title"),
		Category:  c.PostForm("category"),
		SortOrder: atoiOrZero(c.PostForm("sort_order")),
	}
	if errs := Validate(req); errs != nil {
		response.ValidationFailed(c, errs)
		return
	}
	image, ok := parseImageField(c, false)
	if !ok {
		return
	}

	result, err := h.usecase.Update(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c), req, image)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "gallery item updated", result)
}

func (h *Handler) Get(c *gin.Context) {
	result, err := h.usecase.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "gallery item retrieved", result)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.usecase.Delete(c.Request.Context(), c.Param("id"), middleware.CurrentUserID(c)); err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OK(c, "gallery item deleted", nil)
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("limit"))
	result, total, err := h.usecase.List(c.Request.Context(), page, perPage, c.Query("category"))
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "gallery items retrieved", result, response.BuildMeta(page, perPage, total))
}

func (h *Handler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("limit"))
	result, total, err := h.usecase.Search(c.Request.Context(), c.Query("q"), page, perPage)
	if err != nil {
		apperror.WriteHTTP(c, err)
		return
	}
	response.OKWithMeta(c, "gallery items retrieved", result, response.BuildMeta(page, perPage, total))
}

func atoiOrZero(value string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(value))
	return n
}
