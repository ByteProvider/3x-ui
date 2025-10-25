package controller

import (
	"github.com/mhsanaei/3x-ui/v2/web/service"

	"github.com/gin-gonic/gin"
)

// XraySettingController handles Xray configuration and settings operations.
type XraySettingController struct {
	XraySettingService service.XraySettingService
	SettingService     service.SettingService
	InboundService     service.InboundService
	OutboundService    service.OutboundService
	XrayService        service.XrayService
	WarpService        service.WarpService
}

// NewXraySettingController creates a new XraySettingController and initializes its routes.
func NewXraySettingController(g *gin.RouterGroup) *XraySettingController {
	a := &XraySettingController{}
	a.initRouter(g)
	return a
}

// initRouter sets up the routes for Xray settings management.
func (a *XraySettingController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/xray")
	g.GET("/getDefaultJsonConfig", a.getDefaultXrayConfig)
	g.GET("/getOutboundsTraffic", a.getOutboundsTraffic)
	g.GET("/getXrayResult", a.getXrayResult)

	g.POST("/", a.getXraySetting)
	g.POST("/warp/:action", a.warp)
	g.POST("/update", a.updateSetting)
	g.POST("/resetOutboundsTraffic", a.resetOutboundsTraffic)
}

// getXraySetting retrieves the Xray configuration template and inbound tags.
// @Summary      Get Xray settings
// @Description  Retrieve the Xray configuration template and inbound tags
// @Tags         xray
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /xray/ [post]
func (a *XraySettingController) getXraySetting(c *gin.Context) {
	xraySetting, err := a.SettingService.GetXrayConfigTemplate()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getSettings"), err)
		return
	}
	inboundTags, err := a.InboundService.GetInboundTags()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getSettings"), err)
		return
	}
	xrayResponse := "{ \"xraySetting\": " + xraySetting + ", \"inboundTags\": " + inboundTags + " }"
	jsonObj(c, xrayResponse, nil)
}

// updateSetting updates the Xray configuration settings.
// @Summary      Update Xray settings
// @Description  Update the Xray configuration settings
// @Tags         xray
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        xraySetting  formData  string  true  "Xray settings JSON"
// @Success      200          {object}  entity.Msg
// @Failure      400          {object}  entity.Msg
// @Router       /xray/update [post]
func (a *XraySettingController) updateSetting(c *gin.Context) {
	xraySetting := c.PostForm("xraySetting")
	err := a.XraySettingService.SaveXraySetting(xraySetting)
	jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifySettings"), err)
}

// getDefaultXrayConfig retrieves the default Xray configuration.
// @Summary      Get default Xray config
// @Description  Retrieve the default Xray configuration
// @Tags         xray
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /xray/getDefaultJsonConfig [get]
func (a *XraySettingController) getDefaultXrayConfig(c *gin.Context) {
	defaultJsonConfig, err := a.SettingService.GetDefaultXrayConfig()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getSettings"), err)
		return
	}
	jsonObj(c, defaultJsonConfig, nil)
}

// getXrayResult retrieves the current Xray service result.
// @Summary      Get Xray result
// @Description  Retrieve the current Xray service result
// @Tags         xray
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /xray/getXrayResult [get]
func (a *XraySettingController) getXrayResult(c *gin.Context) {
	jsonObj(c, a.XrayService.GetXrayResult(), nil)
}

// warp handles Warp-related operations based on the action parameter.
// @Summary      Handle Warp operations
// @Description  Handle Warp-related operations (data, del, config, reg, license)
// @Tags         xray
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        action      path      string  true   "Warp action (data, del, config, reg, license)"
// @Param        privateKey  formData  string  false  "Private key (for reg action)"
// @Param        publicKey   formData  string  false  "Public key (for reg action)"
// @Param        license     formData  string  false  "License (for license action)"
// @Success      200         {object}  entity.Msg
// @Failure      400         {object}  entity.Msg
// @Router       /xray/warp/{action} [post]
func (a *XraySettingController) warp(c *gin.Context) {
	action := c.Param("action")
	var resp string
	var err error
	switch action {
	case "data":
		resp, err = a.WarpService.GetWarpData()
	case "del":
		err = a.WarpService.DelWarpData()
	case "config":
		resp, err = a.WarpService.GetWarpConfig()
	case "reg":
		skey := c.PostForm("privateKey")
		pkey := c.PostForm("publicKey")
		resp, err = a.WarpService.RegWarp(skey, pkey)
	case "license":
		license := c.PostForm("license")
		resp, err = a.WarpService.SetWarpLicense(license)
	}

	jsonObj(c, resp, err)
}

// getOutboundsTraffic retrieves the traffic statistics for outbounds.
// @Summary      Get outbounds traffic
// @Description  Retrieve the traffic statistics for outbounds
// @Tags         xray
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /xray/getOutboundsTraffic [get]
func (a *XraySettingController) getOutboundsTraffic(c *gin.Context) {
	outboundsTraffic, err := a.OutboundService.GetOutboundsTraffic()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getOutboundTrafficError"), err)
		return
	}
	jsonObj(c, outboundsTraffic, nil)
}

// resetOutboundsTraffic resets the traffic statistics for the specified outbound tag.
// @Summary      Reset outbound traffic
// @Description  Reset the traffic statistics for the specified outbound tag
// @Tags         xray
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        tag  formData  string  true  "Outbound tag"
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /xray/resetOutboundsTraffic [post]
func (a *XraySettingController) resetOutboundsTraffic(c *gin.Context) {
	tag := c.PostForm("tag")
	err := a.OutboundService.ResetOutboundTraffic(tag)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.resetOutboundTrafficError"), err)
		return
	}
	jsonObj(c, "", nil)
}
