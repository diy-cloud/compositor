package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

var ProxyMap = struct {
	sync.Mutex
	m map[string]*httputil.ReverseProxy
}{
	m: make(map[string]*httputil.ReverseProxy),
}

var ProxyRealName = struct {
	sync.Mutex
	m map[string]string
}{
	m: make(map[string]string),
}

var ProxyWorks = struct {
	sync.Mutex
	m map[string]*int
}{
	m: make(map[string]*int),
}

var ProxyPorts = struct {
	sync.Mutex
	m map[string]int
}{
	m: make(map[string]int),
}

func AddProxyServer(routeName, containerName string, port int) error {
	url, err := url.Parse(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		return err
	}
	proxyServer := httputil.NewSingleHostReverseProxy(url)
	ProxyMap.Lock()
	ProxyWorks.Lock()
	ProxyRealName.Lock()
	ProxyPorts.Lock()
	ProxyMap.m[routeName] = proxyServer
	ProxyWorks.m[routeName] = new(int)
	ProxyRealName.m[routeName] = containerName
	ProxyPorts.m[routeName] = port
	ProxyMap.Unlock()
	ProxyWorks.Unlock()
	ProxyRealName.Unlock()
	ProxyPorts.Unlock()
	return nil
}

func GetProxyServer(name string) (*httputil.ReverseProxy, bool) {
	ProxyMap.Lock()
	defer ProxyMap.Unlock()
	server, ok := ProxyMap.m[name]
	return server, ok
}

func RemoveProxyServer(id string) (string, *int, int, error) {
	ProxyMap.Lock()
	ProxyWorks.Lock()
	ProxyRealName.Lock()
	ProxyPorts.Lock()
	works := ProxyWorks.m[id]
	delete(ProxyMap.m, id)
	delete(ProxyWorks.m, id)
	name := ProxyRealName.m[id]
	delete(ProxyRealName.m, id)
	port := ProxyPorts.m[id]
	ProxyMap.Unlock()
	ProxyWorks.Unlock()
	ProxyRealName.Unlock()
	ProxyPorts.Unlock()
	return name, works, port, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")[1:]
	if len(path) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := path[0]
	r.URL.Path = "/" + strings.Join(path[1:], "/")
	ProxyMap.Lock()
	proxyServer, ok := ProxyMap.m[id]
	ProxyMap.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ProxyWorks.Lock()
	workCount := ProxyWorks.m[id]
	ProxyWorks.Unlock()
	*workCount++
	proxyServer.ServeHTTP(w, r)
	*workCount--
}
