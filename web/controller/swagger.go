package controller

import (
	"net/http"

	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/mhsanaei/3x-ui/v2/docs"
)

// SwaggerController handles Swagger documentation routes
type SwaggerController struct {
	BaseController
	settingService service.SettingService
}

// NewSwaggerController creates a new SwaggerController and initializes its routes
func NewSwaggerController(g *gin.RouterGroup) *SwaggerController {
	a := &SwaggerController{}
	a.initRouter(g)
	return a
}

// checkSwaggerEnabled is a middleware that checks if Swagger is enabled
func (a *SwaggerController) checkSwaggerEnabled(c *gin.Context) {
	enabled, err := a.settingService.GetSwaggerEnable()
	if err != nil || !enabled {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Next()
}

// initRouter sets up the Swagger documentation routes
func (a *SwaggerController) initRouter(g *gin.RouterGroup) {
	swagger := g.Group("/swagger")
	swagger.Use(a.checkSwaggerEnabled)
	
	// Serve Swagger UI
	swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

