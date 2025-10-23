package controller

import (
	"net/http"

	"github.com/mhsanaei/3x-ui/v2/web/middleware"
	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/session"

	"github.com/gin-gonic/gin"
)

// APIController handles the main API routes for the 3x-ui panel, including inbounds and server management.
type APIController struct {
	BaseController
	inboundController *InboundController
	serverController  *ServerController
	Tgbot             service.Tgbot
}

// NewAPIController creates a new APIController instance and initializes its routes.
func NewAPIController(g *gin.RouterGroup) *APIController {
	a := &APIController{}
	a.initRouter(g)
	return a
}

// checkAPIAuth is a middleware that returns 404 for unauthenticated API requests
// to hide the existence of API endpoints from unauthorized users
// It supports both session-based auth and API key auth
func (a *APIController) checkAPIAuth(c *gin.Context) {
	// First check session login
	if session.IsLogin(c) {
		c.Next()
		return
	}
	
	// If not logged in via session, return 404 to hide API existence
	c.AbortWithStatus(http.StatusNotFound)
}

// initRouter sets up the API routes for inbounds, server, and other endpoints.
func (a *APIController) initRouter(g *gin.RouterGroup) {
	// Main API group
	api := g.Group("/panel/api")
	api.Use(middleware.ApiKeyAuth())
	api.Use(a.checkAPIAuth)

	// Inbounds API
	inbounds := api.Group("/inbounds")
	a.inboundController = NewInboundController(inbounds)

	// Server API
	server := api.Group("/server")
	a.serverController = NewServerController(server)

	// Extra routes
	api.GET("/backuptotgbot", a.BackuptoTgbot)
}

// BackuptoTgbot sends a backup of the panel data to Telegram bot admins.
// @Summary      Send backup to Telegram
// @Description  Send a backup of the panel database to Telegram bot admins
// @Tags         backup
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      401  {object}  entity.Msg
// @Router       /backuptotgbot [get]
func (a *APIController) BackuptoTgbot(c *gin.Context) {
	a.Tgbot.SendBackupToAdmins()
}
