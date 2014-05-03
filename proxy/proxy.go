package proxy

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/hugb/beegecontroller/config"
)

type Proxy struct {
	*http.Transport
}

type HttpApiFunc func(w http.ResponseWriter, r *http.Request) error

func NewProxyServer() {
	proxy := &Proxy{
		Transport: &http.Transport{
			ResponseHeaderTimeout: time.Duration(5) * time.Second,
		},
	}
	route, err := proxy.createRouter()
	if err != nil {
		panic(err)
	}

	ln, err := net.Listen("tcp", config.CS.ServiceAddress)
	if err != nil {
		panic(err)
	}

	httpSrv := http.Server{Addr: config.CS.ServiceAddress, Handler: route}
	if err = httpSrv.Serve(ln); err != nil {
		panic(err)
	}
}

func (this *Proxy) createRouter() (*mux.Router, error) {
	router := mux.NewRouter()
	routerMap := map[string]map[string]HttpApiFunc{
		"GET": {
			"test": this.test,
		},
		"POST":   {},
		"DELETE": {},
	}
	for method, routes := range routerMap {
		for route, fct := range routes {
			localFct := fct
			localRoute := route
			localMethod := method

			f := makeHttpHandler(localFct)

			if localRoute == "" {
				router.Methods(localMethod).HandlerFunc(f)
			} else {
				router.Path(localRoute).Methods(localMethod).HandlerFunc(f)
				router.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(f)
			}
		}
	}

	return router, nil
}

func makeHttpHandler(handlerFunc HttpApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo:验证版本兼容性

		// todo:处理所有api的公共业务逻辑

		if err := handlerFunc(w, r); err != nil {
			httpError(w, err)
		}
	}
}

// 根据错误生成不同的http错误响应
func httpError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	if strings.Contains(err.Error(), "No such") {
		statusCode = http.StatusNotFound
	} else if strings.Contains(err.Error(), "Bad parameter") {
		statusCode = http.StatusBadRequest
	} else if strings.Contains(err.Error(), "Conflict") {
		statusCode = http.StatusConflict
	} else if strings.Contains(err.Error(), "Impossible") {
		statusCode = http.StatusNotAcceptable
	} else if strings.Contains(err.Error(), "Wrong login/password") {
		statusCode = http.StatusUnauthorized
	} else if strings.Contains(err.Error(), "hasn't been activated") {
		statusCode = http.StatusForbidden
	}

	http.Error(w, err.Error(), statusCode)
}

// 代理到后端web服务器
func (this *Proxy) httpProxy(host string, w http.ResponseWriter, r *http.Request) {
	handler := requestHandler{
		request:  r,
		response: w,
	}
	// 仅支持http1.0和1.1
	if !isProtocolSupported(r) {
		handler.unsupportedProtocol()
		return
	}
	// 负载均衡健康检查
	if isLoadBalancerHeartbeat(r) {
		handler.heartbeat()
		return
	}
	if host == "" {
		handler.missingRoute(host)
		return
	}
	if isTcpUpgrade(r) {
		handler.tcpRequest(host)
		return
	}
	// websocket代理支持
	if isWebSocketUpgrade(r) {
		handler.webSocketRequest(host)
		return
	}
	if response, err := handler.httpRequest(this.Transport, host); err != nil {
		handler.badGateway(err)
	} else {
		handler.writeResponse(response)
	}
}

func isProtocolSupported(request *http.Request) bool {
	return request.ProtoMajor == 1 && (request.ProtoMinor == 0 || request.ProtoMinor == 1)
}

func isLoadBalancerHeartbeat(request *http.Request) bool {
	return request.UserAgent() == "HTTP-Monitor/1.1"
}

func isTcpUpgrade(request *http.Request) bool {
	return upgradeHeader(request) == "tcp"
}

func isWebSocketUpgrade(request *http.Request) bool {
	return strings.ToLower(upgradeHeader(request)) == "websocket"
}

func upgradeHeader(request *http.Request) string {
	if strings.ToLower(request.Header.Get("Connection")) == "upgrade" {
		return request.Header.Get("Upgrade")
	} else {
		return ""
	}
}
