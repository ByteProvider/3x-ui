package controller

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/logger"
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
// @Description  Retrieve traffic statistics for a specific client by email address
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        email  path      string  true  "Client email address"
// @Success      200    {object}  entity.Msg
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
// @Summary      Get client traffic by ID
// @Description  Retrieve traffic statistics for clients in a specific inbound by ID
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Inbound ID"
// @Success      200  {object}  entity.Msg
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
// @Description  Create a new inbound configuration
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
// @Description  Delete an inbound configuration by its ID
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
// @Summary      Get client IPs
// @Description  Retrieve the IP addresses associated with a client by email
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
// @Summary      Clear client IPs
// @Description  Clear the IP addresses for a client by email
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
// @Summary      Add inbound client
// @Description  Add a new client to an existing inbound
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        data  body      model.Inbound  true  "Inbound client data"
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

// AddClientWithLinkRequest defines the request structure for adding a client with only essential fields
type AddClientWithLinkRequest struct {
	Id    int    `json:"id" form:"id" example:"1"`       // Inbound ID
	Email string `json:"email" form:"email" example:"user@example.com"` // Client email address
}

// AddClientWithLinkResponse defines the response structure with generated link and UUID
type AddClientWithLinkResponse struct {
	Link  string `json:"link" example:"vless://uuid@host:port?type=tcp#email"`  // Generated config link
	UUID  string `json:"uuid" example:"9cf47c17-6512-40ec-87e0-e59801366929"`   // Client UUID or password
	Email string `json:"email" example:"user@example.com"`                       // Client email
}

// addInboundClientWithLink adds a new client to an existing inbound and returns the config link.
// @Summary      Add inbound client with link
// @Description  Add a new client to an existing inbound with only email and inbound id, using default values for other parameters
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        data  body      AddClientWithLinkRequest  true  "Inbound ID and client email"
// @Success      200   {object}  entity.Msg{obj=AddClientWithLinkResponse}
// @Failure      400   {object}  entity.Msg
// @Router       /inbounds/addClientWithLink [post]
func (a *InboundController) addInboundClientWithLink(c *gin.Context) {
	request := &AddClientWithLinkRequest{}
	err := c.ShouldBind(request)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.inboundUpdateSuccess"), err)
		return
	}

	// Get the inbound to determine the protocol
	inbound, err := a.inboundService.GetInbound(request.Id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.obtain"), err)
		return
	}

	// Generate default values for the client
	clientId := uuid.New().String()
	clientPassword := random.Seq(10) // For trojan and shadowsocks
	subId := random.Seq(16)

	// Build the settings JSON based on the protocol with default values
	var settingsJSON string
	switch inbound.Protocol {
	case model.VMESS:
		settingsJSON = fmt.Sprintf(`{
			"clients": [{
				"id": "%s",
				"security": "auto",
				"email": "%s",
				"limitIp": 0,
				"totalGB": 0,
				"expiryTime": 0,
				"enable": true,
				"tgId": "",
				"subId": "%s",
				"comment": "",
				"reset": 0
			}]
		}`, clientId, request.Email, subId)
	case model.VLESS:
		settingsJSON = fmt.Sprintf(`{
			"clients": [{
				"id": "%s",
				"flow": "",
				"email": "%s",
				"limitIp": 0,
				"totalGB": 0,
				"expiryTime": 0,
				"enable": true,
				"tgId": "",
				"subId": "%s",
				"comment": "",
				"reset": 0
			}]
		}`, clientId, request.Email, subId)
	case model.Trojan:
		settingsJSON = fmt.Sprintf(`{
			"clients": [{
				"password": "%s",
				"email": "%s",
				"limitIp": 0,
				"totalGB": 0,
				"expiryTime": 0,
				"enable": true,
				"tgId": "",
				"subId": "%s",
				"comment": "",
				"reset": 0
			}]
		}`, clientPassword, request.Email, subId)
	case model.Shadowsocks:
		settingsJSON = fmt.Sprintf(`{
			"clients": [{
				"password": "%s",
				"email": "%s",
				"limitIp": 0,
				"totalGB": 0,
				"expiryTime": 0,
				"enable": true,
				"tgId": "",
				"subId": "%s",
				"comment": "",
				"reset": 0
			}]
		}`, clientPassword, request.Email, subId)
	default:
		jsonMsg(c, "Unsupported protocol", fmt.Errorf("protocol %s not supported", inbound.Protocol))
		return
	}

	// Create the inbound data with settings
	data := &model.Inbound{
		Id:       request.Id,
		Settings: settingsJSON,
	}

	needRestart, err := a.inboundService.AddInboundClient(data)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "somethingWentWrong"), err)
		return
	}

	// Refresh the inbound to get the updated client list
	inbound, err = a.inboundService.GetInbound(request.Id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.obtain"), err)
		return
	}

	// Determine the UUID to return based on protocol
	responseUUID := clientId
	if inbound.Protocol == model.Trojan || inbound.Protocol == model.Shadowsocks {
		responseUUID = clientPassword
	}

	// Get server address from request host
	host := c.Request.Host
	// Remove port if present
	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}

	// Generate the config link using the getLink function from util.go
	link := getLink(inbound, host, request.Email)
	
	// Log if link generation failed
	if link == "" {
		logger.Warning("Failed to generate link for client: ", request.Email, " protocol: ", inbound.Protocol, " host: ", host)
	}

	// Prepare response object
	response := map[string]string{
		"link":  link,
		"uuid":  responseUUID,
		"email": request.Email,
	}

	jsonMsgObj(c, I18nWeb(c, "pages.inbounds.toasts.inboundClientAddSuccess"), response, nil)
	if needRestart {
		a.xrayService.SetToNeedRestart()
	}
}

// delInboundClient deletes a client from an inbound by inbound ID and client ID.
// @Summary      Delete inbound client
// @Description  Delete a client from an inbound by inbound ID and client ID
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
// @Description  Update a client's configuration in an inbound
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        clientId  path      string         true  "Client ID"
// @Param        inbound   body      model.Inbound  true  "Updated client data"
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
// @Description  Reset the traffic counter for a specific client in an inbound
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id     path      int     true  "Inbound ID"
// @Param        email  path      string  true  "Client email address"
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
// @Description  Reset all traffic counters across all inbounds
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
// @Description  Import an inbound configuration from provided JSON data
// @Tags         inbounds
// @Accept       json
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
// @Description  Delete clients in an inbound who have exhausted their traffic limits
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
// @Description  Retrieve the list of currently online clients
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
// @Summary      Get last online clients
// @Description  Retrieve the last online timestamps for clients
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
// @Description  Update the traffic statistics for a client by email
// @Tags         inbounds
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        email    path      string  true  "Client email address"
// @Param        traffic  body      object  true  "Traffic data (upload, download)"
// @Success      200      {object}  entity.Msg
// @Failure      400      {object}  entity.Msg
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
// @Description  Delete a client from an inbound by email address
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
