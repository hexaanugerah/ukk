package middleware

import (
	"strings"

	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	ctxUserIDKey = "auth_user_id"
	ctxRolesKey  = "auth_user_roles"
)

func RequireAuth(jwtManager *jwtutil.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "Silakan login terlebih dahulu.")
			c.Abort()
			return
		}
		claims, err := jwtManager.ParseAccessToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			response.Unauthorized(c, "Sesi tidak valid atau sudah kedaluwarsa. Silakan login kembali.")
			c.Abort()
			return
		}
		c.Set(ctxUserIDKey, claims.UserID)
		c.Set(ctxRolesKey, claims.Roles)
		c.Next()
	}
}

func RequireRole(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[strings.ToUpper(r)] = struct{}{}
	}
	return func(c *gin.Context) {
		rolesRaw, exists := c.Get(ctxRolesKey)
		if !exists {
			response.Forbidden(c, "Anda tidak memiliki akses.")
			c.Abort()
			return
		}
		roles, _ := rolesRaw.([]string)
		for _, r := range roles {
			if _, ok := allowedSet[strings.ToUpper(r)]; ok {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "Anda tidak memiliki akses.")
		c.Abort()
	}
}

func CurrentUserID(c *gin.Context) string {
	v, _ := c.Get(ctxUserIDKey)
	id, _ := v.(string)
	return id
}

// CurrentRoles mengembalikan claim role JWT untuk request, dipakai endpoint yang butuh logika kepemilikan lebih halus daripada RequireRole.
func CurrentRoles(c *gin.Context) []string {
	v, _ := c.Get(ctxRolesKey)
	roles, _ := v.([]string)
	return roles
}

// HasRole menyatakan apakah salah satu role request cocok (tanpa membedakan huruf besar/kecil).
func HasRole(c *gin.Context, role string) bool {
	for _, r := range CurrentRoles(c) {
		if strings.EqualFold(r, role) {
			return true
		}
	}
	return false
}
