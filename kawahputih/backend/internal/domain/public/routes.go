package public

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes memasang endpoint publik yang tidak butuh autentikasi.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	public := rg.Group("/public")
	public.GET("/destinations", h.GetDestinations)
	public.GET("/destinations/:id", h.GetDestinationByID)
	public.GET("/destinations/slug/:slug", h.GetDestinationBySlug)
	public.GET("/ticket-types", h.GetPrices)
	public.GET("/prices", h.GetPrices)
	public.GET("/facilities", h.GetFacilities)
	public.GET("/faq", h.GetFAQs)
	public.GET("/reviews", h.GetReviews)
	public.GET("/content", h.GetContents)
	public.GET("/content/:id", h.GetContentByID)
	public.GET("/content/slug/:slug", h.GetContentBySlug)
}
