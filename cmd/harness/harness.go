// The Harness for a Egret program.
//
// It has a couple responsibilities:
// 1. Parse the user program, generating a main.go file that registers
//    controller classes and starts the user's server.
// 2. Build and run the user program.  Show compile errors.
// 3. Monitor the user source and re-build / restart the program when necessary.

package harness

import (
	"crypto/tls"
	"fmt"
	"go/build"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/kenorld/egret"
	"github.com/kenorld/egret/cmd/utils"
)

var (
	watcher    *egret.Watcher
	doNotWatch = []string{"views"}

	lastRequestHadError int32
)

// Harness reverse proxies requests to the application server.
// It builds / runs / rebuilds / restarts the server when code is changed.
type Harness struct {
	app        *App
	serverHost string
	port       int
	proxy      *httputil.ReverseProxy
}

func renderError(w http.ResponseWriter, r *http.Request, err error) {
	req, resp := egret.NewRequest(r), egret.NewResponse(w)
	c := egret.NewContext(req, resp)
	c.RenderError(err)
	c.ExecuteRender()
}

// ServeHTTP handles all requests.
// It checks for changes to app, rebuilds if necessary, and forwards the request.
func (h *Harness) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Don't rebuild the app for favicon requests.
	if lastRequestHadError > 0 && r.URL.Path == "/favicon.ico" {
		return
	}

	// Flush any change events and rebuild app if necessary.
	// Render an error page if the rebuild / restart failed.
	err := watcher.Notify()
	if err != nil {
		atomic.CompareAndSwapInt32(&lastRequestHadError, 0, 1)
		renderError(w, r, err)
		return
	}
	atomic.CompareAndSwapInt32(&lastRequestHadError, 1, 0)

	// Reverse proxy the request.
	// (Need special code for websockets, courtesy of bradfitz)
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		proxyWebsocket(w, r, h.serverHost)
	} else {
		h.proxy.ServeHTTP(w, r)
	}
}

// NewHarness method returns a reverse proxy that forwards requests
// to the given port.
func NewHarness() *Harness {
	// Get a template loader to render errors.
	// Prefer the app's views/errors directory, and fall back to the stock error pages.
	// egret.MainTemplateLoader = egret.NewTemplateLoader(
	// 	[]string{filepath.Join(egret.EgretPath, "core", "views")})
	// egret.MainTemplateLoader.Refresh()

	addr := egret.HttpAddr
	port := egret.Config.GetIntDefault("harness.port", 0)
	scheme := "http"
	if egret.HttpTLSEnabled {
		scheme = "https"
	}

	// If the server is running on the wildcard address, use "localhost"
	if addr == "" {
		addr = "localhost"
	}

	if port == 0 {
		port = getFreePort()
	}

	serverURL, _ := url.ParseRequestURI(fmt.Sprintf(scheme+"://%s:%d", addr, port))

	harness := &Harness{
		port:       port,
		serverHost: serverURL.String()[len(scheme+"://"):],
		proxy:      httputil.NewSingleHostReverseProxy(serverURL),
	}

	if egret.HttpTLSEnabled {
		harness.proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return harness
}

// Refresh method rebuilds the Egret application and run it on the given port.
func (h *Harness) Refresh() (err *egret.Error) {
	if h.app != nil {
		h.app.Kill()
	}

	utils.Logger.Info("Rebuilding...")
	h.app, err = Build()
	if err != nil {
		return
	}

	h.app.Port = h.port
	if err2 := h.app.Cmd().Start(); err2 != nil {
		return &egret.Error{
			Name:    "failed_start_up",
			Title:   "App failed to start up",
			Summary: err2.Error(),
		}
	}

	return
}

// WatchDir method returns false to file matches with doNotWatch
// otheriwse true
func (h *Harness) WatchDir(info os.FileInfo) bool {
	return !egret.ContainsString(doNotWatch, info.Name())
}

// WatchFile method returns true given filename HasSuffix of ".go"
// otheriwse false
func (h *Harness) WatchFile(filename string) bool {
	return strings.HasSuffix(filename, ".go")
}

// Run the harness, which listens for requests and proxies them to the app
// server, which it runs and rebuilds as necessary.
func (h *Harness) Run() {
	var paths []string
	if egret.Config.GetBoolDefault("watch.gopath", false) {
		gopaths := filepath.SplitList(build.Default.GOPATH)
		paths = append(paths, gopaths...)
	}
	paths = append(paths, egret.CodePaths...)
	watcher = egret.NewWatcher()
	watcher.Listen(h, paths...)

	go func() {
		addr := fmt.Sprintf("%s:%d", egret.HttpAddr, egret.HttpPort)
		utils.Logger.Info("Listening on address: " + addr)

		var err error
		if egret.HttpTLSEnabled {
			err = http.ListenAndServeTLS(addr, egret.HttpTLSCert, egret.HttpTLSKey, h)
		} else {
			err = http.ListenAndServe(addr, h)
		}
		if err != nil {
			utils.Logger.Fatal("Failed to start reverse proxy", zap.Error(err))
		}
	}()

	// Kill the app on signal.
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	if h.app != nil {
		h.app.Kill()
	}
	os.Exit(1)
}

// Find an unused port
func getFreePort() (port int) {
	conn, err := net.Listen("tcp", ":0")
	if err != nil {
		utils.Logger.Fatal("Get free port error", zap.Error(err))
	}

	port = conn.Addr().(*net.TCPAddr).Port
	err = conn.Close()
	if err != nil {
		utils.Logger.Fatal("Get free port error", zap.Error(err))
	}
	return port
}

// proxyWebsocket copies data between websocket client and server until one side
// closes the connection.  (ReverseProxy doesn't work with websocket requests.)
func proxyWebsocket(w http.ResponseWriter, r *http.Request, host string) {
	var (
		d   net.Conn
		err error
	)
	if egret.HttpTLSEnabled {
		// since this proxy isn't used in production,
		// it's OK to set InsecureSkipVerify to true
		// no need to add another configuration option.
		d, err = tls.Dial("tcp", host, &tls.Config{InsecureSkipVerify: true})
	} else {
		d, err = net.Dial("tcp", host)
	}
	if err != nil {
		http.Error(w, "Error contacting backend server.", 500)
		utils.Logger.Error("Error dialing websocket backend", zap.String("host", host), zap.Error(err))
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Not a hijacker?", 500)
		return
	}
	nc, _, err := hj.Hijack()
	if err != nil {
		utils.Logger.Error("Hijack error", zap.Error(err))
		return
	}
	defer nc.Close()
	defer d.Close()

	err = r.Write(d)
	if err != nil {
		utils.Logger.Error("Error copying request to target", zap.Error(err))
		return
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}
	go cp(d, nc)
	go cp(nc, d)
	<-errc
}
