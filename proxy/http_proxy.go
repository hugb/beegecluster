package proxy

import (
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/hugb/beegecontroller/config"
)

type Proxy struct {
	*http.Transport
}

type HttpApiFunc func(w http.ResponseWriter, r *http.Request) error

func NewProxyServer() {
	route, err := createRouter()
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

func createRouter() (*mux.Router, error) {
	/*p := &Proxy{
		Transport: &http.Transport{ResponseHeaderTimeout: time.Duration(5) * time.Second},
	}*/
	r := mux.NewRouter()
	m := map[string]map[string]HttpApiFunc{
		"GET":    {},
		"POST":   {},
		"DELETE": {},
	}
	for method, routes := range m {
		for route, fct := range routes {
			localFct := fct
			localRoute := route
			localMethod := method

			f := makeHttpHandler(localFct)

			if localRoute == "" {
				r.Methods(localMethod).HandlerFunc(f)
			} else {
				r.Path(localRoute).Methods(localMethod).HandlerFunc(f)
				r.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(f)
			}
		}
	}

	return r, nil
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
//
func httpError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

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
	} else {
		//http.StatusInternalServerError
	}

	http.Error(w, err.Error(), statusCode)
}

// 代理到后端web服务器
func (this *Proxy) httpProxy(host string, responseWriter http.ResponseWriter, request *http.Request) {
	handler := NewRequestHandler(request, responseWriter)
	// 仅支持http1.0和1.1
	if !isProtocolSupported(request) {
		handler.HandleUnsupportedProtocol()
		return
	}
	// 负载均衡健康检查
	if isLoadBalancerHeartbeat(request) {
		handler.HandleHeartbeat()
		return
	}
	if host == "" {
		handler.HandleMissingRoute()
		return
	}
	if isTcpUpgrade(request) {
		handler.HandleTcpRequest(host)
		return
	}
	// websocket代理支持
	if isWebSocketUpgrade(request) {
		handler.HandleWebSocketRequest(host)
		return
	}
	response, err := handler.HandleHttpRequest(this.Transport, host)
	if err != nil {
		handler.HandleBadGateway(err)
		return
	}
	handler.WriteResponse(response)
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
