package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mhsanaei/3x-ui/v2/config"
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/logger"
	"github.com/mhsanaei/3x-ui/v2/util/common"
	"github.com/mhsanaei/3x-ui/v2/util/random"
	"github.com/mhsanaei/3x-ui/v2/web/entity"
	"github.com/mhsanaei/3x-ui/v2/xray"

	"github.com/gin-gonic/gin"
)

// getRemoteIp extracts the real IP address from the request headers or remote address.
func getRemoteIp(c *gin.Context) string {
	value := c.GetHeader("X-Real-IP")
	if value != "" {
		return value
	}
	value = c.GetHeader("X-Forwarded-For")
	if value != "" {
		ips := strings.Split(value, ",")
		return ips[0]
	}
	addr := c.Request.RemoteAddr
	ip, _, _ := net.SplitHostPort(addr)
	return ip
}

// jsonMsg sends a JSON response with a message and error status.
func jsonMsg(c *gin.Context, msg string, err error) {
	jsonMsgObj(c, msg, nil, err)
}

// jsonObj sends a JSON response with an object and error status.
func jsonObj(c *gin.Context, obj any, err error) {
	jsonMsgObj(c, "", obj, err)
}

// jsonMsgObj sends a JSON response with a message, object, and error status.
func jsonMsgObj(c *gin.Context, msg string, obj any, err error) {
	m := entity.Msg{
		Obj: obj,
	}
	if err == nil {
		m.Success = true
		if msg != "" {
			m.Msg = msg
		}
	} else {
		m.Success = false
		m.Msg = msg + " (" + err.Error() + ")"
		logger.Warning(msg+" "+I18nWeb(c, "fail")+": ", err)
	}
	c.JSON(http.StatusOK, m)
}

// pureJsonMsg sends a pure JSON message response with custom status code.
func pureJsonMsg(c *gin.Context, statusCode int, success bool, msg string) {
	c.JSON(statusCode, entity.Msg{
		Success: success,
		Msg:     msg,
	})
}

// html renders an HTML template with the provided data and title.
func html(c *gin.Context, name string, title string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["title"] = title
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.GetHeader("X-Real-IP")
	}
	if host == "" {
		var err error
		host, _, err = net.SplitHostPort(c.Request.Host)
		if err != nil {
			host = c.Request.Host
		}
	}
	data["host"] = host
	data["request_uri"] = c.Request.RequestURI
	data["base_path"] = c.GetString("base_path")
	c.HTML(http.StatusOK, name, getContext(data))
}

// getContext adds version and other context data to the provided gin.H.
func getContext(h gin.H) gin.H {
	a := gin.H{
		"cur_ver": config.GetVersion(),
	}
	for key, value := range h {
		a[key] = value
	}
	return a
}

// isAjax checks if the request is an AJAX request.
func isAjax(c *gin.Context) bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

// getLink generates a subscription link for the given inbound, address, and email
func getLink(inbound *model.Inbound, address, email string) string {
	switch inbound.Protocol {
	case "vmess":
		return genVmessLink(inbound, address, email)
	case "vless":
		return genVlessLink(inbound, address, email)
	case "trojan":
		return genTrojanLink(inbound, address, email)
	case "shadowsocks":
		return genShadowsocksLink(inbound, address, email)
	}
	return ""
}

// genVmessLink generates a VMess protocol link for the given inbound and client
func genVmessLink(inbound *model.Inbound, address, email string) string {
	if inbound.Protocol != model.VMESS {
		return ""
	}
	obj := map[string]any{
		"v":    "2",
		"add":  address,
		"port": inbound.Port,
		"type": "none",
	}
	var stream map[string]any
	json.Unmarshal([]byte(inbound.StreamSettings), &stream)
	network, _ := stream["network"].(string)
	obj["net"] = network
	switch network {
	case "tcp":
		tcp, _ := stream["tcpSettings"].(map[string]any)
		header, _ := tcp["header"].(map[string]any)
		typeStr, _ := header["type"].(string)
		obj["type"] = typeStr
		if typeStr == "http" {
			request := header["request"].(map[string]any)
			requestPath, _ := request["path"].([]any)
			obj["path"] = requestPath[0].(string)
			headers, _ := request["headers"].(map[string]any)
			obj["host"] = searchHost(headers)
		}
	case "kcp":
		kcp, _ := stream["kcpSettings"].(map[string]any)
		header, _ := kcp["header"].(map[string]any)
		obj["type"], _ = header["type"].(string)
		obj["path"], _ = kcp["seed"].(string)
	case "ws":
		ws, _ := stream["wsSettings"].(map[string]any)
		obj["path"] = ws["path"].(string)
		if host, ok := ws["host"].(string); ok && len(host) > 0 {
			obj["host"] = host
		} else {
			headers, _ := ws["headers"].(map[string]any)
			obj["host"] = searchHost(headers)
		}
	case "grpc":
		grpc, _ := stream["grpcSettings"].(map[string]any)
		obj["path"] = grpc["serviceName"].(string)
		obj["authority"] = grpc["authority"].(string)
		if grpc["multiMode"].(bool) {
			obj["type"] = "multi"
		}
	case "httpupgrade":
		httpupgrade, _ := stream["httpupgradeSettings"].(map[string]any)
		obj["path"] = httpupgrade["path"].(string)
		if host, ok := httpupgrade["host"].(string); ok && len(host) > 0 {
			obj["host"] = host
		} else {
			headers, _ := httpupgrade["headers"].(map[string]any)
			obj["host"] = searchHost(headers)
		}
	case "xhttp":
		xhttp, _ := stream["xhttpSettings"].(map[string]any)
		obj["path"] = xhttp["path"].(string)
		if host, ok := xhttp["host"].(string); ok && len(host) > 0 {
			obj["host"] = host
		} else {
			headers, _ := xhttp["headers"].(map[string]any)
			obj["host"] = searchHost(headers)
		}
		obj["mode"] = xhttp["mode"].(string)
	}
	security, _ := stream["security"].(string)
	obj["tls"] = security
	if security == "tls" {
		tlsSetting, _ := stream["tlsSettings"].(map[string]any)
		alpns, _ := tlsSetting["alpn"].([]any)
		if len(alpns) > 0 {
			var alpn []string
			for _, a := range alpns {
				alpn = append(alpn, a.(string))
			}
			obj["alpn"] = strings.Join(alpn, ",")
		}
		if sniValue, ok := searchKey(tlsSetting, "serverName"); ok {
			obj["sni"], _ = sniValue.(string)
		}

		tlsSettings, _ := searchKey(tlsSetting, "settings")
		if tlsSetting != nil {
			if fpValue, ok := searchKey(tlsSettings, "fingerprint"); ok {
				obj["fp"], _ = fpValue.(string)
			}
			if insecure, ok := searchKey(tlsSettings, "allowInsecure"); ok {
				obj["allowInsecure"], _ = insecure.(bool)
			}
		}
	}

	// Get clients from inbound settings
	var settings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &settings)
	clientsAny, _ := settings["clients"].([]any)
	var clients []map[string]any
	for _, clientAny := range clientsAny {
		clientMap, _ := clientAny.(map[string]any)
		clients = append(clients, clientMap)
	}

	clientIndex := -1
	for i, client := range clients {
		if clientEmail, ok := client["email"].(string); ok && clientEmail == email {
			clientIndex = i
			break
		}
	}
	if clientIndex == -1 {
		return ""
	}
	obj["id"], _ = clients[clientIndex]["id"].(string)
	obj["scy"], _ = clients[clientIndex]["security"].(string)

	externalProxies, _ := stream["externalProxy"].([]any)

	if len(externalProxies) > 0 {
		links := ""
		for index, externalProxy := range externalProxies {
			ep, _ := externalProxy.(map[string]any)
			newSecurity, _ := ep["forceTls"].(string)
			newObj := map[string]any{}
			for key, value := range obj {
				if !(newSecurity == "none" && (key == "alpn" || key == "sni" || key == "fp" || key == "allowInsecure")) {
					newObj[key] = value
				}
			}
			remarkStr, _ := ep["remark"].(string)
			newObj["ps"] = genRemark(inbound, email, remarkStr, inbound.ClientStats, false)
			newObj["add"] = ep["dest"].(string)
			newObj["port"] = int(ep["port"].(float64))

			if newSecurity != "same" {
				newObj["tls"] = newSecurity
			}
			if index > 0 {
				links += "\n"
			}
			jsonStr, _ := json.MarshalIndent(newObj, "", "  ")
			links += "vmess://" + base64.StdEncoding.EncodeToString(jsonStr)
		}
		return links
	}

	obj["ps"] = genRemark(inbound, email, "", inbound.ClientStats, false)

	jsonStr, _ := json.MarshalIndent(obj, "", "  ")
	return "vmess://" + base64.StdEncoding.EncodeToString(jsonStr)
}

// genVlessLink generates a VLESS protocol link for the given inbound and client
func genVlessLink(inbound *model.Inbound, address, email string) string {
	if inbound.Protocol != model.VLESS {
		return ""
	}
	var stream map[string]any
	json.Unmarshal([]byte(inbound.StreamSettings), &stream)

	// Get clients from inbound settings
	var settings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &settings)
	clientsAny, _ := settings["clients"].([]any)
	var clients []map[string]any
	for _, clientAny := range clientsAny {
		clientMap, _ := clientAny.(map[string]any)
		clients = append(clients, clientMap)
	}

	clientIndex := -1
	for i, client := range clients {
		if clientEmail, ok := client["email"].(string); ok && clientEmail == email {
			clientIndex = i
			break
		}
	}
	if clientIndex == -1 {
		return ""
	}

	uuid, _ := clients[clientIndex]["id"].(string)
	port := inbound.Port
	streamNetwork, _ := stream["network"].(string)
	params := make(map[string]string)
	params["type"] = streamNetwork

	// Add encryption parameter for VLESS from inbound settings
	if encryption, ok := settings["encryption"].(string); ok {
		params["encryption"] = encryption
	}

	switch streamNetwork {
	case "tcp":
		tcp, _ := stream["tcpSettings"].(map[string]any)
		header, _ := tcp["header"].(map[string]any)
		typeStr, _ := header["type"].(string)
		if typeStr == "http" {
			request := header["request"].(map[string]any)
			requestPath, _ := request["path"].([]any)
			params["path"] = requestPath[0].(string)
			headers, _ := request["headers"].(map[string]any)
			params["host"] = searchHost(headers)
			params["headerType"] = "http"
		}
	case "kcp":
		kcp, _ := stream["kcpSettings"].(map[string]any)
		header, _ := kcp["header"].(map[string]any)
		headerType, _ := header["type"].(string)
		params["headerType"] = headerType
		seed, _ := kcp["seed"].(string)
		params["seed"] = seed
	case "ws":
		ws, _ := stream["wsSettings"].(map[string]any)
		path, _ := ws["path"].(string)
		params["path"] = path
		if host, ok := ws["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		} else {
			headers, _ := ws["headers"].(map[string]any)
			params["host"] = searchHost(headers)
		}
	case "grpc":
		grpc, _ := stream["grpcSettings"].(map[string]any)
		serviceName, _ := grpc["serviceName"].(string)
		params["serviceName"] = serviceName
		if authority, ok := grpc["authority"].(string); ok {
			params["authority"] = authority
		}
		if multiMode, ok := grpc["multiMode"].(bool); ok && multiMode {
			params["mode"] = "multi"
		}
	case "httpupgrade":
		httpupgrade, _ := stream["httpupgradeSettings"].(map[string]any)
		path, _ := httpupgrade["path"].(string)
		params["path"] = path
		if host, ok := httpupgrade["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		} else {
			headers, _ := httpupgrade["headers"].(map[string]any)
			params["host"] = searchHost(headers)
		}
	case "xhttp":
		xhttp, _ := stream["xhttpSettings"].(map[string]any)
		path, _ := xhttp["path"].(string)
		params["path"] = path
		if host, ok := xhttp["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		} else {
			headers, _ := xhttp["headers"].(map[string]any)
			params["host"] = searchHost(headers)
		}
		mode, _ := xhttp["mode"].(string)
		params["mode"] = mode
	}
	security, _ := stream["security"].(string)
	if security == "tls" {
		params["security"] = "tls"
		tlsSetting, _ := stream["tlsSettings"].(map[string]any)
		alpns, _ := tlsSetting["alpn"].([]any)
		var alpn []string
		for _, a := range alpns {
			alpn = append(alpn, a.(string))
		}
		if len(alpn) > 0 {
			params["alpn"] = strings.Join(alpn, ",")
		}
		if sniValue, ok := searchKey(tlsSetting, "serverName"); ok {
			if sni, ok := sniValue.(string); ok {
				params["sni"] = sni
			}
		}

		tlsSettings, _ := searchKey(tlsSetting, "settings")
		if tlsSettings != nil {
			if fpValue, ok := searchKey(tlsSettings, "fingerprint"); ok {
				if fp, ok := fpValue.(string); ok {
					params["fp"] = fp
				}
			}
			if insecure, ok := searchKey(tlsSettings, "allowInsecure"); ok {
				if insecure.(bool) {
					params["allowInsecure"] = "1"
				}
			}
		}

		if streamNetwork == "tcp" {
			if flow, ok := clients[clientIndex]["flow"].(string); ok && len(flow) > 0 {
				params["flow"] = flow
			}
		}
	}

	if security == "reality" {
		params["security"] = "reality"
		realitySetting, _ := stream["realitySettings"].(map[string]any)
		realitySettings, _ := searchKey(realitySetting, "settings")
		if realitySetting != nil {
			if sniValue, ok := searchKey(realitySetting, "serverNames"); ok {
				sNames, _ := sniValue.([]any)
				if len(sNames) > 0 {
					params["sni"] = sNames[random.Num(len(sNames))].(string)
				}
			}
			if pbkValue, ok := searchKey(realitySettings, "publicKey"); ok {
				if pbk, ok := pbkValue.(string); ok {
					params["pbk"] = pbk
				}
			}
			if sidValue, ok := searchKey(realitySetting, "shortIds"); ok {
				shortIds, _ := sidValue.([]any)
				if len(shortIds) > 0 {
					params["sid"] = shortIds[random.Num(len(shortIds))].(string)
				}
			}
			if fpValue, ok := searchKey(realitySettings, "fingerprint"); ok {
				if fp, ok := fpValue.(string); ok && len(fp) > 0 {
					params["fp"] = fp
				}
			}
			if pqvValue, ok := searchKey(realitySettings, "mldsa65Verify"); ok {
				if pqv, ok := pqvValue.(string); ok && len(pqv) > 0 {
					params["pqv"] = pqv
				}
			}
			params["spx"] = "/" + random.Seq(15)
		}

		if streamNetwork == "tcp" {
			if flow, ok := clients[clientIndex]["flow"].(string); ok && len(flow) > 0 {
				params["flow"] = flow
			}
		}
	}

	if security != "tls" && security != "reality" {
		params["security"] = "none"
	}

	externalProxies, _ := stream["externalProxy"].([]any)

	if len(externalProxies) > 0 {
		links := ""
		for index, externalProxy := range externalProxies {
			ep, _ := externalProxy.(map[string]any)
			newSecurity, _ := ep["forceTls"].(string)
			dest, _ := ep["dest"].(string)
			port := int(ep["port"].(float64))
			link := fmt.Sprintf("vless://%s@%s:%d", uuid, dest, port)

			if newSecurity != "same" {
				params["security"] = newSecurity
			} else {
				params["security"] = security
			}
			url, _ := url.Parse(link)
			q := url.Query()

			for k, v := range params {
				if !(newSecurity == "none" && (k == "alpn" || k == "sni" || k == "fp" || k == "allowInsecure")) {
					q.Add(k, v)
				}
			}

			// Set the new query values on the URL
			url.RawQuery = q.Encode()

			remarkStr, _ := ep["remark"].(string)
			url.Fragment = genRemark(inbound, email, remarkStr, inbound.ClientStats, false)

			if index > 0 {
				links += "\n"
			}
			links += url.String()
		}
		return links
	}

	link := fmt.Sprintf("vless://%s@%s:%d", uuid, address, port)
	url, _ := url.Parse(link)
	q := url.Query()

	for k, v := range params {
		q.Add(k, v)
	}

	// Set the new query values on the URL
	url.RawQuery = q.Encode()

	url.Fragment = genRemark(inbound, email, "", inbound.ClientStats, false)
	return url.String()
}

// genTrojanLink generates a Trojan protocol link for the given inbound and client
func genTrojanLink(inbound *model.Inbound, address, email string) string {
	if inbound.Protocol != model.Trojan {
		return ""
	}
	var stream map[string]any
	json.Unmarshal([]byte(inbound.StreamSettings), &stream)

	// Get clients from inbound settings
	var settings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &settings)
	clientsAny, _ := settings["clients"].([]any)
	var clients []map[string]any
	for _, clientAny := range clientsAny {
		clientMap, _ := clientAny.(map[string]any)
		clients = append(clients, clientMap)
	}

	clientIndex := -1
	for i, client := range clients {
		if clientEmail, ok := client["email"].(string); ok && clientEmail == email {
			clientIndex = i
			break
		}
	}
	if clientIndex == -1 {
		return ""
	}

	password, _ := clients[clientIndex]["password"].(string)
	port := inbound.Port
	streamNetwork, _ := stream["network"].(string)
	params := make(map[string]string)
	params["type"] = streamNetwork

	switch streamNetwork {
	case "tcp":
		tcp, _ := stream["tcpSettings"].(map[string]any)
		header, _ := tcp["header"].(map[string]any)
		typeStr, _ := header["type"].(string)
		if typeStr == "http" {
			request := header["request"].(map[string]any)
			requestPath, _ := request["path"].([]any)
			params["path"] = requestPath[0].(string)
			headers, _ := request["headers"].(map[string]any)
			params["host"] = searchHost(headers)
			params["headerType"] = "http"
		}
	case "kcp":
		kcp, _ := stream["kcpSettings"].(map[string]any)
		header, _ := kcp["header"].(map[string]any)
		headerType, _ := header["type"].(string)
		params["headerType"] = headerType
		seed, _ := kcp["seed"].(string)
		params["seed"] = seed
	case "ws":
		ws, _ := stream["wsSettings"].(map[string]any)
		path, _ := ws["path"].(string)
		params["path"] = path
		if host, ok := ws["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		} else {
			headers, _ := ws["headers"].(map[string]any)
			params["host"] = searchHost(headers)
		}
	case "grpc":
		grpc, _ := stream["grpcSettings"].(map[string]any)
		serviceName, _ := grpc["serviceName"].(string)
		params["serviceName"] = serviceName
		if authority, ok := grpc["authority"].(string); ok {
			params["authority"] = authority
		}
		if multiMode, ok := grpc["multiMode"].(bool); ok && multiMode {
			params["mode"] = "multi"
		}
	case "httpupgrade":
		httpupgrade, _ := stream["httpupgradeSettings"].(map[string]any)
		path, _ := httpupgrade["path"].(string)
		params["path"] = path
		if host, ok := httpupgrade["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		} else {
			headers, _ := httpupgrade["headers"].(map[string]any)
			params["host"] = searchHost(headers)
		}
	case "xhttp":
		xhttp, _ := stream["xhttpSettings"].(map[string]any)
		path, _ := xhttp["path"].(string)
		params["path"] = path
		if host, ok := xhttp["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		} else {
			headers, _ := xhttp["headers"].(map[string]any)
			params["host"] = searchHost(headers)
		}
		mode, _ := xhttp["mode"].(string)
		params["mode"] = mode
	}
	security, _ := stream["security"].(string)
	if security == "tls" {
		params["security"] = "tls"
		tlsSetting, _ := stream["tlsSettings"].(map[string]any)
		alpns, _ := tlsSetting["alpn"].([]any)
		var alpn []string
		for _, a := range alpns {
			alpn = append(alpn, a.(string))
		}
		if len(alpn) > 0 {
			params["alpn"] = strings.Join(alpn, ",")
		}
		if sniValue, ok := searchKey(tlsSetting, "serverName"); ok {
			if sni, ok := sniValue.(string); ok {
				params["sni"] = sni
			}
		}

		tlsSettings, _ := searchKey(tlsSetting, "settings")
		if tlsSettings != nil {
			if fpValue, ok := searchKey(tlsSettings, "fingerprint"); ok {
				if fp, ok := fpValue.(string); ok {
					params["fp"] = fp
				}
			}
			if insecure, ok := searchKey(tlsSettings, "allowInsecure"); ok {
				if insecure.(bool) {
					params["allowInsecure"] = "1"
				}
			}
		}
	}

	if security == "reality" {
		params["security"] = "reality"
		realitySetting, _ := stream["realitySettings"].(map[string]any)
		realitySettings, _ := searchKey(realitySetting, "settings")
		if realitySetting != nil {
			if sniValue, ok := searchKey(realitySetting, "serverNames"); ok {
				sNames, _ := sniValue.([]any)
				if len(sNames) > 0 {
					params["sni"] = sNames[random.Num(len(sNames))].(string)
				}
			}
			if pbkValue, ok := searchKey(realitySettings, "publicKey"); ok {
				if pbk, ok := pbkValue.(string); ok {
					params["pbk"] = pbk
				}
			}
			if sidValue, ok := searchKey(realitySetting, "shortIds"); ok {
				shortIds, _ := sidValue.([]any)
				if len(shortIds) > 0 {
					params["sid"] = shortIds[random.Num(len(shortIds))].(string)
				}
			}
			if fpValue, ok := searchKey(realitySettings, "fingerprint"); ok {
				if fp, ok := fpValue.(string); ok && len(fp) > 0 {
					params["fp"] = fp
				}
			}
			if pqvValue, ok := searchKey(realitySettings, "mldsa65Verify"); ok {
				if pqv, ok := pqvValue.(string); ok && len(pqv) > 0 {
					params["pqv"] = pqv
				}
			}
			params["spx"] = "/" + random.Seq(15)
		}

		if streamNetwork == "tcp" {
			if flow, ok := clients[clientIndex]["flow"].(string); ok && len(flow) > 0 {
				params["flow"] = flow
			}
		}
	}

	if security != "tls" && security != "reality" {
		params["security"] = "none"
	}

	externalProxies, _ := stream["externalProxy"].([]any)

	if len(externalProxies) > 0 {
		links := ""
		for index, externalProxy := range externalProxies {
			ep, _ := externalProxy.(map[string]any)
			newSecurity, _ := ep["forceTls"].(string)
			dest, _ := ep["dest"].(string)
			port := int(ep["port"].(float64))
			link := fmt.Sprintf("trojan://%s@%s:%d", password, dest, port)

			if newSecurity != "same" {
				params["security"] = newSecurity
			} else {
				params["security"] = security
			}
			url, _ := url.Parse(link)
			q := url.Query()

			for k, v := range params {
				if !(newSecurity == "none" && (k == "alpn" || k == "sni" || k == "fp" || k == "allowInsecure")) {
					q.Add(k, v)
				}
			}

			// Set the new query values on the URL
			url.RawQuery = q.Encode()

			remarkStr, _ := ep["remark"].(string)
			url.Fragment = genRemark(inbound, email, remarkStr, inbound.ClientStats, false)

			if index > 0 {
				links += "\n"
			}
			links += url.String()
		}
		return links
	}

	link := fmt.Sprintf("trojan://%s@%s:%d", password, address, port)

	url, _ := url.Parse(link)
	q := url.Query()

	for k, v := range params {
		q.Add(k, v)
	}

	// Set the new query values on the URL
	url.RawQuery = q.Encode()

	url.Fragment = genRemark(inbound, email, "", inbound.ClientStats, false)
	return url.String()
}

// genShadowsocksLink generates a Shadowsocks protocol link for the given inbound and client
func genShadowsocksLink(inbound *model.Inbound, address, email string) string {
	if inbound.Protocol != model.Shadowsocks {
		return ""
	}
	var stream map[string]any
	json.Unmarshal([]byte(inbound.StreamSettings), &stream)

	var settings map[string]any
	json.Unmarshal([]byte(inbound.Settings), &settings)
	inboundPassword, _ := settings["password"].(string)
	method, _ := settings["method"].(string)

	clientsAny, _ := settings["clients"].([]any)
	var clients []map[string]any
	for _, clientAny := range clientsAny {
		clientMap, _ := clientAny.(map[string]any)
		clients = append(clients, clientMap)
	}

	clientIndex := -1
	for i, client := range clients {
		if clientEmail, ok := client["email"].(string); ok && clientEmail == email {
			clientIndex = i
			break
		}
	}
	if clientIndex == -1 {
		return ""
	}

	streamNetwork, _ := stream["network"].(string)
	params := make(map[string]string)
	params["type"] = streamNetwork

	switch streamNetwork {
	case "tcp":
		tcp, _ := stream["tcpSettings"].(map[string]any)
		header, _ := tcp["header"].(map[string]any)
		typeStr, _ := header["type"].(string)
		if typeStr == "http" {
			request := header["request"].(map[string]any)
			requestPath, _ := request["path"].([]any)
			params["path"] = requestPath[0].(string)
			headers, _ := request["headers"].(map[string]any)
			params["host"] = searchHost(headers)
			params["headerType"] = "http"
		}
	case "kcp":
		kcp, _ := stream["kcpSettings"].(map[string]any)
		header, _ := kcp["header"].(map[string]any)
		headerType, _ := header["type"].(string)
		params["headerType"] = headerType
		seed, _ := kcp["seed"].(string)
		params["seed"] = seed
	case "ws":
		ws, _ := stream["wsSettings"].(map[string]any)
		path, _ := ws["path"].(string)
		params["path"] = path
		if host, ok := ws["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		} else {
			headers, _ := ws["headers"].(map[string]any)
			params["host"] = searchHost(headers)
		}
	case "grpc":
		grpc, _ := stream["grpcSettings"].(map[string]any)
		serviceName, _ := grpc["serviceName"].(string)
		params["serviceName"] = serviceName
		if authority, ok := grpc["authority"].(string); ok {
			params["authority"] = authority
		}
		if multiMode, ok := grpc["multiMode"].(bool); ok && multiMode {
			params["mode"] = "multi"
		}
	case "httpupgrade":
		httpupgrade, _ := stream["httpupgradeSettings"].(map[string]any)
		path, _ := httpupgrade["path"].(string)
		params["path"] = path
		if host, ok := httpupgrade["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		} else {
			headers, _ := httpupgrade["headers"].(map[string]any)
			params["host"] = searchHost(headers)
		}
	case "xhttp":
		xhttp, _ := stream["xhttpSettings"].(map[string]any)
		path, _ := xhttp["path"].(string)
		params["path"] = path
		if host, ok := xhttp["host"].(string); ok && len(host) > 0 {
			params["host"] = host
		} else {
			headers, _ := xhttp["headers"].(map[string]any)
			params["host"] = searchHost(headers)
		}
		mode, _ := xhttp["mode"].(string)
		params["mode"] = mode
	}

	security, _ := stream["security"].(string)
	if security == "tls" {
		params["security"] = "tls"
		tlsSetting, _ := stream["tlsSettings"].(map[string]any)
		alpns, _ := tlsSetting["alpn"].([]any)
		var alpn []string
		for _, a := range alpns {
			alpn = append(alpn, a.(string))
		}
		if len(alpn) > 0 {
			params["alpn"] = strings.Join(alpn, ",")
		}
		if sniValue, ok := searchKey(tlsSetting, "serverName"); ok {
			if sni, ok := sniValue.(string); ok {
				params["sni"] = sni
			}
		}

		tlsSettings, _ := searchKey(tlsSetting, "settings")
		if tlsSettings != nil {
			if fpValue, ok := searchKey(tlsSettings, "fingerprint"); ok {
				if fp, ok := fpValue.(string); ok {
					params["fp"] = fp
				}
			}
			if insecure, ok := searchKey(tlsSettings, "allowInsecure"); ok {
				if insecure.(bool) {
					params["allowInsecure"] = "1"
				}
			}
		}
	}

	clientPassword, _ := clients[clientIndex]["password"].(string)
	encPart := fmt.Sprintf("%s:%s", method, clientPassword)
	if method[0] == '2' {
		encPart = fmt.Sprintf("%s:%s:%s", method, inboundPassword, clientPassword)
	}

	externalProxies, _ := stream["externalProxy"].([]any)

	if len(externalProxies) > 0 {
		links := ""
		for index, externalProxy := range externalProxies {
			ep, _ := externalProxy.(map[string]any)
			newSecurity, _ := ep["forceTls"].(string)
			dest, _ := ep["dest"].(string)
			port := int(ep["port"].(float64))
			link := fmt.Sprintf("ss://%s@%s:%d", base64.StdEncoding.EncodeToString([]byte(encPart)), dest, port)

			if newSecurity != "same" {
				params["security"] = newSecurity
			} else {
				params["security"] = security
			}
			url, _ := url.Parse(link)
			q := url.Query()

			for k, v := range params {
				if !(newSecurity == "none" && (k == "alpn" || k == "sni" || k == "fp" || k == "allowInsecure")) {
					q.Add(k, v)
				}
			}

			// Set the new query values on the URL
			url.RawQuery = q.Encode()

			remarkStr, _ := ep["remark"].(string)
			url.Fragment = genRemark(inbound, email, remarkStr, inbound.ClientStats, false)

			if index > 0 {
				links += "\n"
			}
			links += url.String()
		}
		return links
	}

	link := fmt.Sprintf("ss://%s@%s:%d", base64.StdEncoding.EncodeToString([]byte(encPart)), address, inbound.Port)
	url, _ := url.Parse(link)
	q := url.Query()

	for k, v := range params {
		q.Add(k, v)
	}

	// Set the new query values on the URL
	url.RawQuery = q.Encode()

	url.Fragment = genRemark(inbound, email, "", inbound.ClientStats, false)
	return url.String()
}

// genRemark generates a remark string for subscription links
func genRemark(inbound *model.Inbound, email string, extra string, clientStats []xray.ClientTraffic, showInfo bool) string {
	// For simplified version without remarkModel, just return the inbound remark + email
	separationChar := " "

	var remark []string
	if len(inbound.Remark) > 0 {
		remark = append(remark, inbound.Remark)
	}
	if len(email) > 0 {
		remark = append(remark, email)
	}
	if len(extra) > 0 {
		remark = append(remark, extra)
	}

	if showInfo {
		statsExist := false
		var stats xray.ClientTraffic
		for _, clientStat := range clientStats {
			if clientStat.Email == email {
				stats = clientStat
				statsExist = true
				break
			}
		}

		// Get remained days
		if statsExist {
			if !stats.Enable {
				return fmt.Sprintf("⛔️N/A%s%s", separationChar, strings.Join(remark, separationChar))
			}
			if vol := stats.Total - (stats.Up + stats.Down); vol > 0 {
				remark = append(remark, fmt.Sprintf("%s%s", common.FormatTraffic(vol), "📊"))
			}
			now := time.Now().Unix()
			switch exp := stats.ExpiryTime / 1000; {
			case exp > 0:
				remainingSeconds := exp - now
				days := remainingSeconds / 86400
				hours := (remainingSeconds % 86400) / 3600
				minutes := (remainingSeconds % 3600) / 60
				if days > 0 {
					if hours > 0 {
						remark = append(remark, fmt.Sprintf("%dD,%dH⏳", days, hours))
					} else {
						remark = append(remark, fmt.Sprintf("%dD⏳", days))
					}
				} else if hours > 0 {
					remark = append(remark, fmt.Sprintf("%dH⏳", hours))
				} else {
					remark = append(remark, fmt.Sprintf("%dM⏳", minutes))
				}
			case exp < 0:
				days := exp / -86400
				hours := (exp % -86400) / 3600
				minutes := (exp % -3600) / 60
				if days > 0 {
					if hours > 0 {
						remark = append(remark, fmt.Sprintf("%dD,%dH⏳", days, hours))
					} else {
						remark = append(remark, fmt.Sprintf("%dD⏳", days))
					}
				} else if hours > 0 {
					remark = append(remark, fmt.Sprintf("%dH⏳", hours))
				} else {
					remark = append(remark, fmt.Sprintf("%dM⏳", minutes))
				}
			}
		}
	}
	return strings.Join(remark, separationChar)
}

// searchKey recursively searches for a key in a nested map or array structure
func searchKey(data any, key string) (any, bool) {
	switch val := data.(type) {
	case map[string]any:
		for k, v := range val {
			if k == key {
				return v, true
			}
			if result, ok := searchKey(v, key); ok {
				return result, true
			}
		}
	case []any:
		for _, v := range val {
			if result, ok := searchKey(v, key); ok {
				return result, true
			}
		}
	}
	return nil, false
}

// searchHost searches for the host header in request headers
func searchHost(headers any) string {
	data, _ := headers.(map[string]any)
	for k, v := range data {
		if strings.EqualFold(k, "host") {
			switch v.(type) {
			case []any:
				hosts, _ := v.([]any)
				if len(hosts) > 0 {
					return hosts[0].(string)
				} else {
					return ""
				}
			case any:
				return v.(string)
			}
		}
	}

	return ""
}
