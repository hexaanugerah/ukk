package middleware

import (
	"strings"

	"github.com/KawahPutih/backend/internal/domain/audit"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

const (
	ctxClaimsKey = "jwt_claims"
)

// AuditMiddleware mengambil info pengguna dari JWT dan menambahkannya ke context untuk audit logging. Pasang setelah RequireAuth.
func AuditMiddleware(jwtManager *jwtutil.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil info pengguna dari context yang ada (diisi RequireAuth)
		actorID := CurrentUserID(c)
		roles := CurrentRoles(c)
		actorRole := ""
		if len(roles) > 0 {
			actorRole = roles[0]
		}

		// Ambil IP dan User-Agent
		ip := c.ClientIP()
		ua := c.GetHeader("User-Agent")

		// Tambahkan ke context untuk service audit
		ctx := audit.WithAuditInfo(c.Request.Context(), actorID, actorRole, ip, ua)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OptionalAuditMiddleware seperti AuditMiddleware tapi tidak mewajibkan auth. Menambah info audit bila tersedia.
func OptionalAuditMiddleware(jwtManager *jwtutil.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		actorID := ""
		actorRole := ""

		if header != "" && strings.HasPrefix(header, "Bearer ") {
			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := jwtManager.ParseAccessToken(token)
			if err == nil {
				actorID = claims.UserID
				if len(claims.Roles) > 0 {
					actorRole = claims.Roles[0]
				}
			}
		}

		ip := c.ClientIP()
		ua := c.GetHeader("User-Agent")

		ctx := audit.WithAuditInfo(c.Request.Context(), actorID, actorRole, ip, ua)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetAuditActorID mengembalikan ID aktor audit dari context.
func GetAuditActorID(c *gin.Context) string {
	return CurrentUserID(c)
}

// GetAuditActorRole mengembalikan role aktor audit dari context.
func GetAuditActorRole(c *gin.Context) string {
	roles := CurrentRoles(c)
	if len(roles) > 0 {
		return roles[0]
	}
	return ""
}
