package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hugb/beegecluster/registry"
	"github.com/hugb/beegecluster/utils"
)

func (this *Proxy) getImagesJSON(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	host := utils.GetHostFromQueryParam(r)
	if host == "" {
		imagesBytes, err := json.Marshal(registry.RegistryServer.GetAllImages())
		if err != nil {
			fmt.Fprintf(w, "images json encode error: %s", err)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(imagesBytes)
		}
	} else {
		this.httpProxy(host, w, r)
	}
	return nil
}

func (this *Proxy) getImagesByName(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}

	this.httpProxy(registry.RegistryServer.GetHostByImageId(vars["name"]), w, r)

	return nil
}
