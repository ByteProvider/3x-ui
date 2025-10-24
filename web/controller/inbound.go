package controller

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/session"
	webutil "github.com/mhsanaei/3x-ui/v2/web/util"

	"github.com/gin-gonic/gin"
)

// InboundController handles HTTP requests related to Xray inbounds management.
type InboundController struct {
	inboundService service.InboundService
	xrayService    service.XrayService
	settingService service.SettingService
}

// NewInboundController creates a new InboundController and sets up its routes.
func NewInboundController(g *gin.RouterGroup) *InboundController {
	a := &InboundController{}
	a.initRouter(g)
	return a
}

// initRouter initializes the routes for inbound-related operations.
func (a *InboundController) initRouter(g *gin.RouterGroup) {

	g.GET("/list", a.getInbounds)
	g.GET("/get/:id", a.getInbound)
	g.GET("/getClientTraffics/:email", a.getClientTraffics)
	g.GET("/getClientTrafficsById/:id", a.getClientTrafficsById)

	g.POST("/add", a.addInbound)
	g.POST("/del/:id", a.delInbound)
	g.POST("/update/:id", a.updateInbound)
	g.POST("/clientIps/:email", a.getClientIps)
	g.POST("/clearClientIps/:email", a.clearClientIps)
	g.POST("/addClient", a.addInboundClient)
	g.POST("/addClientWithLink", a.addInboundClientWithLink)
	g.POST("/:id/delClient/:clientId", a.delInboundClient)
	g.POST("/updateClient/:clientId", a.updateInboundClient)
	g.POST("/:id/resetClientTraffic/:email", a.resetClientTraffic)
	g.POST("/resetAllTraffics", a.resetAllTraffics)
	g.POST("/resetAllClientTraffics/:id", a.resetAllClientTraffics)
	g.POST("/delDepletedClients/:id", a.delDepletedClients)
	g.POST("/import", a.importInbound)
	g.POST("/onlines", a.onlines)
	g.POST("/lastOnline", a.lastOnline)
	g.POST("/updateClientTraffic/:email", a.updateClientTraffic)
	g.POST("/:id/delClientByEmail/:email", a.delInboundClientByEmail)
}

// getInbounds retrieves the list of inbounds for the logged-in user.
// @Summary      List all inbounds
// @Description  Get list of all inbounds for the authenticated user
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg{obj=[]model.Inbound}
// @Failure      400  {object}  entity.Msg
// @Failure      401  {object}  entity.Msg
// @Router       /inbounds/list [get]
func (a *InboundController) getInbounds(c *gin.Context) {
	user := session.GetLoginUser(c)
	inbounds, err := a.inboundService.GetInbounds(user.Id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
	jsonObj(c, inbounds, nil)
}

// getInbound retrieves a specific inbound by its ID.
// @Summary      Get inbound by ID
// @Description  Get detailed information about a specific inbound
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "Inbound ID"
// @Success      200  {object}  entity.Msg{obj=model.Inbound}
// @Failure      400  {object}  entity.Msg
// @Failure      404  {object}  entity.Msg
// @Router       /inbounds/get/{id} [get]
func (a *InboundController) getInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "get"), err)
		return
	}
	inbound, err := a.inboundService.GetInbound(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
	jsonObj(c, inbound, nil)
}

// getClientTraffics retrieves client traffic information by email.
// @Summary      Get client traffic by email
// @Description  Get traffic statistics for a specific client by email address
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        email  path      string  true  "Client email address"
// @Success      200    {object}  entity.Msg{obj=xray.ClientTraffic}
// @Failure      400    {object}  entity.Msg
// @Router       /inbounds/getClientTraffics/{email} [get]
func (a *InboundController) getClientTraffics(c *gin.Context) {
	email := c.Param("email")
	clientTraffics, err := a.inboundService.GetClientTrafficByEmail(email)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.trafficGetError"), err)
		return
	}
	jsonObj(c, clientTraffics, nil)
}

// getClientTrafficsById retrieves client traffic information by inbound ID.
// @Summary      Get client traffic by inbound ID
// @Description  Get traffic statistics for all clients in a specific inbound
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Inbound ID"
// @Success      200  {object}  entity.Msg{obj=[]xray.ClientTraffic}
// @Failure      400  {object}  entity.Msg
// @Router       /inbounds/getClientTrafficsById/{id} [get]
func (a *InboundController) getClientTrafficsById(c *gin.Context) {
	id := c.Param("id")
	clientTraffics, err := a.inboundService.GetClientTrafficByID(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.trafficGetError"), err)
		return
	}
	jsonObj(c, clientTraffics, nil)
}

// addInbound creates a new inbound configuration.
// @Summary      Add new inbound
// @Description  Create a new inbound with specified configuration
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        inbound  body      model.Inbound  true  "Inbound configuration"
// @Success      200      {object}  entity.Msg{obj=model.Inbound}
// @Failure      400      {object}  entity.Msg
// @Router       /inbounds/add [post]
func (a *InboundController) addInbound(c *gin.Context) {
	inbound := &model.Inbound{}
	err := c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundCreateSuccess"), err)
		return
	}
	user := session.GetLoginUser(c)
	inbound.UserId = user.Id
	if inbound.Listen == "" || inbound.Listen == "0.0.0.0" || inbound.Listen == "::" || inbound.Listen == "::0" {
		inbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)
	} else {
		inbound.Tag = fmt.Sprintf("inbound-%v:%v", inbound.Listen, inbound.Port)
	}

	inbound, needRestart, err := a.inboundService.AddInbound(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundCreateSuccess"), inbound, nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// delInbound deletes an inbound configuration by its ID.
// @Summary      Delete inbound
// @Description  Delete an inbound by its ID
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "Inbound ID"
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /inbounds/del/{id} [post]
func (a *InboundController) delInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundDeleteSuccess"), err)
		return
	}
	needRestart, err := a.inboundService.DelInbound(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundDeleteSuccess"), id, nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// updateInbound updates an existing inbound configuration.
// @Summary      Update inbound
// @Description  Update an existing inbound configuration
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id       path      int            true  "Inbound ID"
// @Param        inbound  body      model.Inbound  true  "Updated inbound configuration"
// @Success      200      {object}  entity.Msg{obj=model.Inbound}
// @Failure      400      {object}  entity.Msg
// @Router       /inbounds/update/{id} [post]
func (a *InboundController) updateInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}
	inbound := &model.Inbound{
		Id: id,
	}
	err = c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}
	inbound, needRestart, err := a.inboundService.UpdateInbound(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), inbound, nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// getClientIps retrieves the IP addresses associated with a client by email.
// @Summary      Get client IP addresses
// @Description  Retrieve IP addresses recorded for a specific client
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        email  path      string  true  "Client email address"
// @Success      200    {object}  entity.Msg
// @Failure      400    {object}  entity.Msg
// @Router       /inbounds/clientIps/{email} [post]
func (a *InboundController) getClientIps(c *gin.Context) {
	email := c.Param("email")

	ips, err := a.inboundService.GetInboundClientIps(email)
	if err != nil || ips == "" {
		jsonObj(c, "No IP Record", nil)
		return
	}

	jsonObj(c, ips, nil)
}

// clearClientIps clears the IP addresses for a client by email.
// @Summary      Clear client IP addresses
// @Description  Clear all recorded IP addresses for a specific client
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        email  path      string  true  "Client email address"
// @Success      200    {object}  entity.Msg
// @Failure      400    {object}  entity.Msg
// @Router       /inbounds/clearClientIps/{email} [post]
func (a *InboundController) clearClientIps(c *gin.Context) {
	email := c.Param("email")

	err := a.inboundService.ClearClientIps(email)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.updateSuccess"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.logCleanSuccess"), nil)
}

// addInboundClient adds a new client to an existing inbound.
// @Summary      Add client to inbound
// @Description  Add a new client to an existing inbound configuration
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        data  body      model.Inbound  true  "Client data"
// @Success      200   {object}  entity.Msg
// @Failure      400   {object}  entity.Msg
// @Router       /inbounds/addClient [post]
func (a *InboundController) addInboundClient(c *gin.Context) {
	data := &model.Inbound{}
	err := c.ShouldBind(data)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}

	needRestart, err := a.inboundService.AddInboundClient(data)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientAddSuccess"), nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// AddClientRequest represents a simplified request to add a client with only essential fields
type AddClientRequest struct {
	InboundID  int    `json:"inboundId" form:"inboundId" binding:"required"` // Inbound ID
	Email      string `json:"email" form:"email" binding:"required"`          // Client email
	TotalGB    int64  `json:"totalGB" form:"totalGB"`                         // Total traffic in GB (0 for unlimited)
	ExpiryTime int64  `json:"expiryTime" form:"expiryTime"`                   // Expiration time in milliseconds (0 for no expiration)
	LimitIP    int    `json:"limitIp" form:"limitIp"`                         // IP limit (0 for unlimited)
	TgID       int64  `json:"tgId" form:"tgId"`                               // Telegram ID
	SubID      string `json:"subId" form:"subId"`                             // Subscription ID
}

// addInboundClientWithLink adds a new client to an existing inbound and returns the connection link.
// @Summary      Add client to inbound with link (simplified)
// @Description  Add a new client with only email and inbound ID - other fields are auto-generated. Returns the connection link (vless://, vmess://, trojan://, ss://)
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Param        X-API-Key  header    string              true  "API Key for authentication"
// @Param        request    body      AddClientRequest    true  "Client request with email and inbound ID"
// @Success      200        {object}  entity.Msg{obj=map[string]any}  "Returns link, email, and client details"
// @Failure      400        {object}  entity.Msg
// @Failure      401        {object}  entity.Msg
// @Router       /panel/api/inbounds/addClientWithLink [post]
func (a *InboundController) addInboundClientWithLink(c *gin.Context) {
	request := &AddClientRequest{}
	err := c.ShouldBindJSON(request)
	if err != nil {
		jsonMsg(c, "Invalid request parameters", err)
		return
	}

	// Get the inbound first to determine the protocol
	inbound, err := a.inboundService.GetInbound(request.InboundID)
	if err != nil {
		jsonMsg(c, "Inbound not found", err)
		return
	}

	// Generate client with default values
	client, err := webutil.GenerateClientDefaults(inbound.Protocol, request.Email, request.TotalGB, request.ExpiryTime, request.LimitIP, request.TgID, request.SubID)
	if err != nil {
		jsonMsg(c, "Failed to generate client", err)
		return
	}

	// Create the inbound data structure for adding client
	var settings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &settings)
	if settings == nil {
		settings = make(map[string]any)
	}
	
	// Add the new client to clients array
	clientsArray := []any{client}
	settings["clients"] = clientsArray
	
	newSettings, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		jsonMsg(c, "Failed to marshal settings", err)
		return
	}

	data := &model.Inbound{
		Id:       request.InboundID,
		Settings: string(newSettings),
	}

	needRestart, err := a.inboundService.AddInboundClient(data)
	if err != nil {
		jsonMsg(c, "Failed to add client", err)
		return
	}

	// Get the address - try settings first, then fall back to request host
	address := ""
	subDomain, err := a.settingService.GetSubDomain()
	if err == nil && subDomain != "" {
		address = subDomain
	} else {
		// Fall back to request host
		address = c.Request.Host
		if address == "" {
			address = "localhost"
		}
		// Remove port from host if present
		if idx := strings.Index(address, ":"); idx != -1 {
			address = address[:idx]
		}
	}
	
	link := webutil.GetClientLink(inbound, request.Email, address)

	// Convert client map to a clean response
	clientMap, _ := client.(map[string]any)
	
	response := map[string]any{
		"link":   link,
		"email":  request.Email,
		"client": clientMap,
	}

	jsonMsgObj(c, "Client added successfully", response, nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// delInboundClient deletes a client from an inbound by inbound ID and client ID.
// @Summary      Delete client from inbound
// @Description  Remove a client from an inbound by client ID
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id        path      int     true  "Inbound ID"
// @Param        clientId  path      string  true  "Client ID"
// @Success      200       {object}  entity.Msg
// @Failure      400       {object}  entity.Msg
// @Router       /inbounds/{id}/delClient/{clientId} [post]
func (a *InboundController) delInboundClient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}
	clientId := c.Param("clientId")

	needRestart, err := a.inboundService.DelInboundClient(id, clientId)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientDeleteSuccess"), nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// updateInboundClient updates a client's configuration in an inbound.
// @Summary      Update inbound client
// @Description  Update configuration for a specific client in an inbound
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        clientId  path      string         true  "Client ID"
// @Param        data      body      model.Inbound  true  "Updated client data"
// @Success      200       {object}  entity.Msg
// @Failure      400       {object}  entity.Msg
// @Router       /inbounds/updateClient/{clientId} [post]
func (a *InboundController) updateInboundClient(c *gin.Context) {
	clientId := c.Param("clientId")

	inbound := &model.Inbound{}
	err := c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}

	needRestart, err := a.inboundService.UpdateInboundClient(inbound, clientId)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientUpdateSuccess"), nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// resetClientTraffic resets the traffic counter for a specific client in an inbound.
// @Summary      Reset client traffic
// @Description  Reset traffic counter for a specific client
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id     path      int     true  "Inbound ID"
// @Param        email  path      string  true  "Client email"
// @Success      200    {object}  entity.Msg
// @Failure      400    {object}  entity.Msg
// @Router       /inbounds/{id}/resetClientTraffic/{email} [post]
func (a *InboundController) resetClientTraffic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}
	email := c.Param("email")

	needRestart, err := a.inboundService.ResetClientTraffic(id, email)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.resetInboundClientTrafficSuccess"), nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// resetAllTraffics resets all traffic counters across all inbounds.
// @Summary      Reset all traffics
// @Description  Reset traffic counters for all inbounds and clients
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /inbounds/resetAllTraffics [post]
func (a *InboundController) resetAllTraffics(c *gin.Context) {
	err := a.inboundService.ResetAllTraffics()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	} else {
		a.xrayService.SetToNeedRestart()
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.resetAllTrafficSuccess"), nil)
}

// resetAllClientTraffics resets traffic counters for all clients in a specific inbound.
// @Summary      Reset all client traffics
// @Description  Reset traffic counters for all clients in a specific inbound
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "Inbound ID"
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /inbounds/resetAllClientTraffics/{id} [post]
func (a *InboundController) resetAllClientTraffics(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}

	err = a.inboundService.ResetAllClientTraffics(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	} else {
		a.xrayService.SetToNeedRestart()
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.resetAllClientTrafficSuccess"), nil)
}

// importInbound imports an inbound configuration from provided data.
// @Summary      Import inbound
// @Description  Import an inbound configuration from JSON data
// @Tags         inbounds
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Security     ApiKeyAuth
// @Param        data  formData  string  true  "Inbound JSON data"
// @Success      200   {object}  entity.Msg{obj=model.Inbound}
// @Failure      400   {object}  entity.Msg
// @Router       /inbounds/import [post]
func (a *InboundController) importInbound(c *gin.Context) {
	inbound := &model.Inbound{}
	err := json.Unmarshal([]byte(c.PostForm("data")), inbound)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	user := session.GetLoginUser(c)
	inbound.Id = 0
	inbound.UserId = user.Id
	if inbound.Listen == "" || inbound.Listen == "0.0.0.0" || inbound.Listen == "::" || inbound.Listen == "::0" {
		inbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)
	} else {
		inbound.Tag = fmt.Sprintf("inbound-%v:%v", inbound.Listen, inbound.Port)
	}

	for index := range inbound.ClientStats {
		inbound.ClientStats[index].Id = 0
		inbound.ClientStats[index].Enable = true
	}

	needRestart := false
	inbound, needRestart, err = a.inboundService.AddInbound(inbound)
	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundCreateSuccess"), inbound, err)
	if err == nil && needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// delDepletedClients deletes clients in an inbound who have exhausted their traffic limits.
// @Summary      Delete depleted clients
// @Description  Remove all clients who have exhausted their traffic limits
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "Inbound ID"
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /inbounds/delDepletedClients/{id} [post]
func (a *InboundController) delDepletedClients(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}
	err = a.inboundService.DelDepletedClients(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.delDepletedClientsSuccess"), nil)
}

// onlines retrieves the list of currently online clients.
// @Summary      Get online clients
// @Description  Get list of currently online clients
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /inbounds/onlines [post]
func (a *InboundController) onlines(c *gin.Context) {
	jsonObj(c, a.inboundService.GetOnlineClients(), nil)
}

// lastOnline retrieves the last online timestamps for clients.
// @Summary      Get last online times
// @Description  Get last online timestamps for all clients
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /inbounds/lastOnline [post]
func (a *InboundController) lastOnline(c *gin.Context) {
	data, err := a.inboundService.GetClientsLastOnline()
	jsonObj(c, data, err)
}

// updateClientTraffic updates the traffic statistics for a client by email.
// @Summary      Update client traffic
// @Description  Manually update traffic statistics for a specific client
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        email  path  string  true  "Client email"
// @Param        traffic  body  object  true  "Traffic data"  example({"upload": 1024, "download": 2048})
// @Success      200  {object}  entity.Msg
// @Failure      400  {object}  entity.Msg
// @Router       /inbounds/updateClientTraffic/{email} [post]
func (a *InboundController) updateClientTraffic(c *gin.Context) {
	email := c.Param("email")

	// Define the request structure for traffic update
	type TrafficUpdateRequest struct {
		Upload   int64 `json:"upload"`
		Download int64 `json:"download"`
	}

	var request TrafficUpdateRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}

	err = a.inboundService.UpdateClientTrafficByEmail(email, request.Upload, request.Download)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientUpdateSuccess"), nil)
}

// delInboundClientByEmail deletes a client from an inbound by email address.
// @Summary      Delete client by email
// @Description  Remove a client from an inbound by email address
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id     path      int     true  "Inbound ID"
// @Param        email  path      string  true  "Client email address"
// @Success      200    {object}  entity.Msg
// @Failure      400    {object}  entity.Msg
// @Router       /inbounds/{id}/delClientByEmail/{email} [post]
func (a *InboundController) delInboundClientByEmail(c *gin.Context) {
	inboundId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "Invalid inbound ID", err)
		return
	}

	email := c.Param("email")
	needRestart, err := a.inboundService.DelInboundClientByEmail(inboundId, email)
	if err != nil {
		jsonMsg(c, "Failed to delete client by email", err)
		return
	}

	jsonMsg(c, "Client deleted successfully", nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}
