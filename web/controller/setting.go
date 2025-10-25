package controller

import (
	"errors"
	"time"

	"github.com/mhsanaei/3x-ui/v2/util/crypto"
	"github.com/mhsanaei/3x-ui/v2/web/entity"
	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/session"

	"github.com/gin-gonic/gin"
)

// updateUserForm represents the form for updating user credentials.
type updateUserForm struct {
	OldUsername string `json:"oldUsername" form:"oldUsername"`
	OldPassword string `json:"oldPassword" form:"oldPassword"`
	NewUsername string `json:"newUsername" form:"newUsername"`
	NewPassword string `json:"newPassword" form:"newPassword"`
}

// SettingController handles settings and user management operations.
type SettingController struct {
	settingService service.SettingService
	userService    service.UserService
	panelService   service.PanelService
}

// NewSettingController creates a new SettingController and initializes its routes.
func NewSettingController(g *gin.RouterGroup) *SettingController {
	a := &SettingController{}
	a.initRouter(g)
	return a
}

// initRouter sets up the routes for settings management.
func (a *SettingController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/setting")

	g.POST("/all", a.getAllSetting)
	g.POST("/defaultSettings", a.getDefaultSettings)
	g.POST("/update", a.updateSetting)
	g.POST("/updateUser", a.updateUser)
	g.POST("/restartPanel", a.restartPanel)
	g.GET("/getDefaultJsonConfig", a.getDefaultXrayConfig)
	g.GET("/getApiKey", a.getApiKey)
	g.POST("/generateApiKey", a.generateApiKey)
}

// getAllSetting retrieves all current settings.
// @Summary      Get all settings
// @Description  Retrieve all current panel settings
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg{obj=entity.AllSetting}
// @Failure      400  {object}  entity.Msg
// @Router       /setting/all [post]
func (a *SettingController) getAllSetting(c *gin.Context) {
	allSetting, err := a.settingService.GetAllSetting()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getSettings"), err)
		return
	}
	jsonObj(c, allSetting, nil)
}

// getDefaultSettings retrieves the default settings based on the host.
// @Summary      Get default settings
// @Description  Retrieve the default settings based on the host
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /setting/defaultSettings [post]
func (a *SettingController) getDefaultSettings(c *gin.Context) {
	result, err := a.settingService.GetDefaultSettings(c.Request.Host)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getSettings"), err)
		return
	}
	jsonObj(c, result, nil)
}

// updateSetting updates all settings with the provided data.
// @Summary      Update settings
// @Description  Update all settings with the provided data
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        settings  body      entity.AllSetting  true  "All settings"
// @Success      200       {object}  entity.Msg
// @Failure      400       {object}  entity.Msg
// @Router       /setting/update [post]
func (a *SettingController) updateSetting(c *gin.Context) {
	allSetting := &entity.AllSetting{}
	err := c.ShouldBind(allSetting)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifySettings"), err)
		return
	}
	err = a.settingService.UpdateAllSetting(allSetting)
	jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifySettings"), err)
}

// updateUser updates the current user's username and password.
// @Summary      Update user credentials
// @Description  Update the current user's username and password
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        user  body      updateUserForm  true  "User credentials"
// @Success      200   {object}  entity.Msg
// @Failure      400   {object}  entity.Msg
// @Router       /setting/updateUser [post]
func (a *SettingController) updateUser(c *gin.Context) {
	form := &updateUserForm{}
	err := c.ShouldBind(form)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifySettings"), err)
		return
	}
	user := session.GetLoginUser(c)
	if user.Username != form.OldUsername || !crypto.CheckPasswordHash(user.Password, form.OldPassword) {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifyUserError"), errors.New(I18nWeb(c, "pages.settings.toasts.originalUserPassIncorrect")))
		return
	}
	if form.NewUsername == "" || form.NewPassword == "" {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifyUserError"), errors.New(I18nWeb(c, "pages.settings.toasts.userPassMustBeNotEmpty")))
		return
	}
	err = a.userService.UpdateUser(user.Id, form.NewUsername, form.NewPassword)
	if err == nil {
		user.Username = form.NewUsername
		user.Password, _ = crypto.HashPasswordAsBcrypt(form.NewPassword)
		session.SetLoginUser(c, user)
	}
	jsonMsg(c, I18nWeb(c, "pages.settings.toasts.modifyUser"), err)
}

// restartPanel restarts the panel service after a delay.
// @Summary      Restart panel
// @Description  Restart the panel service after a delay
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /setting/restartPanel [post]
func (a *SettingController) restartPanel(c *gin.Context) {
	err := a.panelService.RestartPanel(time.Second * 3)
	jsonMsg(c, I18nWeb(c, "pages.settings.restartPanelSuccess"), err)
}

// getDefaultXrayConfig retrieves the default Xray configuration.
// @Summary      Get default Xray config
// @Description  Retrieve the default Xray configuration
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /setting/getDefaultJsonConfig [get]
func (a *SettingController) getDefaultXrayConfig(c *gin.Context) {
	defaultJsonConfig, err := a.settingService.GetDefaultXrayConfig()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getSettings"), err)
		return
	}
	jsonObj(c, defaultJsonConfig, nil)
}

// getApiKey retrieves the current user's API key
// @Summary      Get API key
// @Description  Retrieve the current user's API key
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg{obj=string}
// @Failure      401  {object}  entity.Msg
// @Router       /setting/getApiKey [get]
func (a *SettingController) getApiKey(c *gin.Context) {
	user := session.GetLoginUser(c)
	if user == nil {
		jsonMsg(c, "Unauthorized", errors.New("user not logged in"))
		return
	}
	
	apiKey, err := a.userService.GetApiKey(user.Id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.getApiKey"), err)
		return
	}
	jsonObj(c, apiKey, nil)
}

// generateApiKey generates a new API key for the current user
// @Summary      Generate API key
// @Description  Generate a new API key for the current user
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg{obj=string}
// @Failure      401  {object}  entity.Msg
// @Router       /setting/generateApiKey [post]
func (a *SettingController) generateApiKey(c *gin.Context) {
	user := session.GetLoginUser(c)
	if user == nil {
		jsonMsg(c, "Unauthorized", errors.New("user not logged in"))
		return
	}
	
	apiKey, err := a.userService.GenerateApiKey(user.Id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.settings.toasts.generateApiKey"), err)
		return
	}
	jsonObj(c, apiKey, nil)
}
