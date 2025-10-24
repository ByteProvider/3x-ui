package controller

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/sub"
	"github.com/mhsanaei/3x-ui/v2/util/random"
	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/session"

	"github.com/gin-gonic/gin"
)

// InboundController handles HTTP requests related to Xray inbounds management.
type InboundController struct {
	inboundService service.InboundService
	xrayService    service.XrayService
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
	g.POST("/createClient", a.createClientWithConfig)
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

// createClientWithConfig creates a new client and returns the configuration link.
// @Summary      Create client with config link
// @Description  Create a new client and automatically generate configuration link (vless://, vmess://, etc.)
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      object  true  "Client creation request"  example({"id": 1, "email": "user@example.com", "totalGB": 10737418240, "expiryTime": 1735689600000, "enable": true, "limitIp": 2, "subId": "randomSubId"})
// @Success      200      {object}  entity.Msg{obj=object}  "Returns client UUID, email, and config link"
// @Failure      400      {object}  entity.Msg
// @Router       /inbounds/createClient [post]
func (a *InboundController) createClientWithConfig(c *gin.Context) {
	// Parse request
	var req struct {
		Id         int    `json:"id"`
		Email      string `json:"email"`
		TotalGB    int64  `json:"totalGB"`
		ExpiryTime int64  `json:"expiryTime"`
		Enable     bool   `json:"enable"`
		LimitIP    int    `json:"limitIp"`
		TgID       string `json:"tgId"`
		SubID      string `json:"subId"`
		Flow       string `json:"flow"`
		Comment    string `json:"comment"`
		Reset      int    `json:"reset"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		jsonMsg(c, "Invalid request data", err)
		return
	}

	if req.Email == "" {
		jsonMsg(c, "Email is required", fmt.Errorf("email cannot be empty"))
		return
	}

	// Get the inbound
	inbound, err := a.inboundService.GetInbound(req.Id)
	if err != nil {
		jsonMsg(c, "Failed to get inbound", err)
		return
	}

	// Generate client ID based on protocol
	var clientID, password, clientData string
	nowTs := time.Now().Unix() * 1000

	switch inbound.Protocol {
	case model.VMESS:
		clientID = uuid.New().String()
		clientData = fmt.Sprintf(`{
			"id": "%s",
			"security": "auto",
			"email": "%s",
			"limitIp": %d,
			"totalGB": %d,
			"expiryTime": %d,
			"enable": %t,
			"tgId": "%s",
			"subId": "%s",
			"comment": "%s",
			"reset": %d,
			"created_at": %d,
			"updated_at": %d
		}`, clientID, req.Email, req.LimitIP, req.TotalGB, req.ExpiryTime, req.Enable, req.TgID, req.SubID, req.Comment, req.Reset, nowTs, nowTs)

	case model.VLESS:
		clientID = uuid.New().String()
		flow := req.Flow
		if flow == "" {
			flow = ""
		}
		clientData = fmt.Sprintf(`{
			"id": "%s",
			"flow": "%s",
			"email": "%s",
			"limitIp": %d,
			"totalGB": %d,
			"expiryTime": %d,
			"enable": %t,
			"tgId": "%s",
			"subId": "%s",
			"comment": "%s",
			"reset": %d,
			"created_at": %d,
			"updated_at": %d
		}`, clientID, flow, req.Email, req.LimitIP, req.TotalGB, req.ExpiryTime, req.Enable, req.TgID, req.SubID, req.Comment, req.Reset, nowTs, nowTs)

	case model.Trojan:
		password = random.Seq(10)
		clientData = fmt.Sprintf(`{
			"password": "%s",
			"email": "%s",
			"limitIp": %d,
			"totalGB": %d,
			"expiryTime": %d,
			"enable": %t,
			"tgId": "%s",
			"subId": "%s",
			"comment": "%s",
			"reset": %d,
			"created_at": %d,
			"updated_at": %d
		}`, password, req.Email, req.LimitIP, req.TotalGB, req.ExpiryTime, req.Enable, req.TgID, req.SubID, req.Comment, req.Reset, nowTs, nowTs)

	case model.Shadowsocks:
		password = random.Seq(32)
		clientData = fmt.Sprintf(`{
			"password": "%s",
			"email": "%s",
			"limitIp": %d,
			"totalGB": %d,
			"expiryTime": %d,
			"enable": %t,
			"tgId": "%s",
			"subId": "%s",
			"comment": "%s",
			"reset": %d,
			"created_at": %d,
			"updated_at": %d
		}`, password, req.Email, req.LimitIP, req.TotalGB, req.ExpiryTime, req.Enable, req.TgID, req.SubID, req.Comment, req.Reset, nowTs, nowTs)

	default:
		jsonMsg(c, "Unsupported protocol", fmt.Errorf("protocol %s not supported", inbound.Protocol))
		return
	}

	// Parse and update inbound settings
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
		jsonMsg(c, "Failed to parse inbound settings", err)
		return
	}

	// Parse the client data
	var newClient interface{}
	if err := json.Unmarshal([]byte(clientData), &newClient); err != nil {
		jsonMsg(c, "Failed to parse client data", err)
		return
	}

	// Add client to settings
	clients, ok := settings["clients"].([]interface{})
	if !ok {
		clients = []interface{}{}
	}
	clients = append(clients, newClient)
	settings["clients"] = clients

	// Marshal back to string
	newSettings, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		jsonMsg(c, "Failed to marshal settings", err)
		return
	}

	inbound.Settings = string(newSettings)

	// Add client to inbound
	needRestart, err := a.inboundService.AddInboundClient(inbound)
	if err != nil {
		jsonMsg(c, "Failed to add client", err)
		return
	}

	// Generate configuration link
	var settingService service.SettingService
	host, _ := settingService.GetSubDomain()
	if host == "" {
		host = c.Request.Host
	}

	subService := sub.NewSubService(false, "")

	// Reload inbound to get updated client list
	inbound, err = a.inboundService.GetInbound(req.Id)
	if err != nil {
		jsonMsg(c, "Failed to reload inbound", err)
		return
	}

	configLink := subService.GetLinkWithAddress(inbound, req.Email, host)

	// Prepare response
	response := map[string]interface{}{
		"success": true,
		"email":   req.Email,
		"link":    configLink,
	}

	if clientID != "" {
		response["uuid"] = clientID
	}
	if password != "" {
		response["password"] = password
	}

	jsonObj(c, response, nil)
	
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
