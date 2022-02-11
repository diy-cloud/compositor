package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/snowmerak/compositor/config"
	"github.com/snowmerak/compositor/vm/multipass"
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

func AddProxyServer(id, name, dst string) error {
	url, err := url.Parse(dst)
	if err != nil {
		return err
	}
	proxyServer := httputil.NewSingleHostReverseProxy(url)
	ProxyMap.Lock()
	ProxyWorks.Lock()
	ProxyRealName.Lock()
	ProxyMap.m[id] = proxyServer
	ProxyWorks.m[id] = new(int)
	ProxyRealName.m[id] = name
	ProxyMap.Unlock()
	ProxyWorks.Unlock()
	ProxyRealName.Unlock()
	return nil
}

func GetProxyServer(id string) (*httputil.ReverseProxy, bool) {
	ProxyMap.Lock()
	defer ProxyMap.Unlock()
	server, ok := ProxyMap.m[id]
	return server, ok
}

func RemoveProxyServer(id string) error {
	ProxyMap.Lock()
	ProxyWorks.Lock()
	ProxyRealName.Lock()
	works := ProxyWorks.m[id]
	delete(ProxyMap.m, id)
	delete(ProxyWorks.m, id)
	name := ProxyRealName.m[id]
	delete(ProxyRealName.m, id)
	ProxyMap.Unlock()
	ProxyWorks.Unlock()
	ProxyRealName.Unlock()
	for *works > 0 {
		runtime.Gosched()
	}
	instance := multipass.New()
	if err := os.RemoveAll(filepath.Join(config.HomePath, name)); err != nil {
		return err
	}
	return instance.Delete(name)
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
