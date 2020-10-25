package routing

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"goplugins/core/framework/color"
	"goplugins/core/framework/log"
	"io"
	"io/ioutil"
	stdLog "log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

type (
	// Mux is the top-level routing framework instance.
	Mux struct {
		common
		StdLogger        *stdLog.Logger
		colorer          *color.Color
		premiddleware    []MiddlewareFunc
		middleware       []MiddlewareFunc
		maxParam         *int
		router           *Router
		routers          map[string]*Router
		notFoundHandler  HandlerFunc
		pool             sync.Pool
		Server           *http.Server
		TLSServer        *http.Server
		Listener         net.Listener
		TLSListener      net.Listener
		AutoTLSManager   autocert.Manager
		DisableHTTP2     bool
		Debug            bool
		HidePort         bool
		HTTPErrorHandler HTTPErrorHandler
		Binder           Binder
		Validator        Validator
		Renderer         Renderer
		Logger           Logger
		IPExtractor      IPExtractor
	}

	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// HTTPErrorHandler is a centralized HTTP error handler.
	HTTPErrorHandler func(error, Context)

	// Validator is the interface that wraps the Validate function.
	Validator interface {
		Validate(i interface{}) error
	}

	// HandlerFunc defines a function to serve HTTP requests.
	HandlerFunc func(Context) error

	// Renderer is the interface that wraps the Render function.
	Renderer interface {
		Render(io.Writer, string, interface{}, Context) error
	}

	// Map defines a generic map of type `map[string]interface{}`.
	Map map[string]interface{}

	// Common struct for Mux & Group.
	common struct{}
)

const (
	charsetUTF8 = "charset=UTF-8"
	// PROPFIND Method can be used on collection and property resources.
	PROPFIND = "PROPFIND"
	// REPORT Method can be used to get information about a resource, see rfc 3253
	REPORT = "REPORT"
)

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextXML                          = "text/xml"
	MIMETextXMLCharsetUTF8               = MIMETextXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

// Headers
const (
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXXSSProtection                  = "X-XSS-Protection"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderXCSRFToken                      = "X-CSRF-Token"
	HeaderReferrerPolicy                  = "Referrer-Policy"
)

// New creates an instance of Mux.
func New() (m *Mux) {
	m = &Mux{
		Server:    new(http.Server),
		TLSServer: new(http.Server),
		AutoTLSManager: autocert.Manager{
			Prompt: autocert.AcceptTOS,
		},
		Logger:   log.New("router"),
		colorer:  color.New(),
		maxParam: new(int),
	}
	m.Server.Handler = m
	m.TLSServer.Handler = m
	m.HTTPErrorHandler = m.DefaultHTTPErrorHandler
	m.Binder = &DefaultBinder{}
	m.pool.New = func() interface{} {
		return m.NewContext(nil, nil)
	}

	m.router = NewRouter(m)
	m.routers = map[string]*Router{}

	return
}

// NewContext returns a Context instance.
func (m *Mux) NewContext(r *http.Request, w http.ResponseWriter) Context {
	return &context{
		request:  r,
		response: NewResponse(w, m),
		store:    make(Map),
		mux:      m,
		pvalues:  make([]string, *m.maxParam),
		handler:  NotFoundHandler,
	}
}

// Router returns the default router.
func (m *Mux) Router() *Router {
	return m.router
}

// Routers returns the map of host => router.
func (m *Mux) Routers() map[string]*Router {
	return m.routers
}

// Pre adds middleware to the chain which is run before router.
func (m *Mux) Pre(middleware ...MiddlewareFunc) {
	m.premiddleware = append(m.premiddleware, middleware...)
}

// Use adds middleware to the chain which is run after router.
func (m *Mux) Use(middleware ...MiddlewareFunc) {
	m.middleware = append(m.middleware, middleware...)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the
// router with optional route-level middleware.
func (m *Mux) CONNECT(path string, h HandlerFunc, mf ...MiddlewareFunc) *Route {
	return m.Add(http.MethodConnect, path, h, mf...)
}

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
func (m *Mux) DELETE(path string, h HandlerFunc, mf ...MiddlewareFunc) *Route {
	return m.Add(http.MethodDelete, path, h, mf...)
}

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (m *Mux) GET(path string, h HandlerFunc, mf ...MiddlewareFunc) *Route {
	return m.Add(http.MethodGet, path, h, mf...)
}

// HEAD registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware.
func (m *Mux) HEAD(path string, h HandlerFunc, mf ...MiddlewareFunc) *Route {
	return m.Add(http.MethodHead, path, h, mf...)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
func (m *Mux) OPTIONS(path string, h HandlerFunc, mf ...MiddlewareFunc) *Route {
	return m.Add(http.MethodOptions, path, h, mf...)
}

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
func (m *Mux) PATCH(path string, h HandlerFunc, mf ...MiddlewareFunc) *Route {
	return m.Add(http.MethodPatch, path, h, mf...)
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
func (m *Mux) POST(path string, h HandlerFunc, mf ...MiddlewareFunc) *Route {
	return m.Add(http.MethodPost, path, h, mf...)
}

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
func (m *Mux) PUT(path string, h HandlerFunc, mf ...MiddlewareFunc) *Route {
	return m.Add(http.MethodPut, path, h, mf...)
}

// TRACE registers a new TRACE route for a path with matching handler in the
// router with optional route-level middleware.
func (m *Mux) TRACE(path string, h HandlerFunc, mf ...MiddlewareFunc) *Route {
	return m.Add(http.MethodTrace, path, h, mf...)
}

// Any registers a new route for all HTTP methods and path with matching handler
// in the router with optional route-level middleware.
func (m *Mux) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route {
	routes := make([]*Route, len(methods))
	for i, met := range methods {
		routes[i] = m.Add(met, path, handler, middleware...)
	}
	return routes
}

// Match registers a new route for multiple HTTP methods and path with matching
// handler in the router with optional route-level middleware.
func (m *Mux) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route {
	routes := make([]*Route, len(methods))
	for i, met := range methods {
		routes[i] = m.Add(met, path, handler, middleware...)
	}
	return routes
}

// Static registers a new route with path prefix to serve static files from the
// provided root directory.
func (m *Mux) Static(prefix, root string) *Route {
	if root == "" {
		root = "." // For security we want to restrict to CWD.
	}
	return m.static(prefix, root, m.GET)
}

func (common) static(prefix, root string, get func(string, HandlerFunc, ...MiddlewareFunc) *Route) *Route {
	h := func(c Context) error {
		p, err := url.PathUnescape(c.Param("*"))
		if err != nil {
			return err
		}

		name := filepath.Join(root, path.Clean("/"+p)) // "/"+ for security
		fi, err := os.Stat(name)
		if err != nil {
			// The access path does not exist
			return NotFoundHandler(c)
		}

		// If the request is for a directory and does not end with "/"
		p = c.Request().URL.Path // path must not be empty.
		if fi.IsDir() && p[len(p)-1] != '/' {
			// Redirect to ends with "/"
			return c.Redirect(http.StatusMovedPermanently, p+"/")
		}
		return c.File(name)
	}
	if prefix == "/" {
		return get(prefix+"*", h)
	}
	return get(prefix+"/*", h)
}

func (common) file(path, file string, get func(string, HandlerFunc, ...MiddlewareFunc) *Route,
	m ...MiddlewareFunc) *Route {
	return get(path, func(c Context) error {
		return c.File(file)
	}, m...)
}

// File registers a new route with path to serve a static file with optional route-level middleware.
func (m *Mux) File(path, file string, mf ...MiddlewareFunc) *Route {
	return m.file(path, file, m.GET, mf...)
}

func (m *Mux) add(host, method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	name := handlerName(handler)
	router := m.findRouter(host)
	router.Add(method, path, func(c Context) error {
		h := applyMiddleware(handler, middleware...)
		return h(c)
	})
	r := &Route{
		Method: method,
		Path:   path,
		Name:   name,
	}
	m.router.routes[method+path] = r
	return r
}

// Add registers a new route for an HTTP method and path with matching handler
// in the router with optional route-level middleware.
func (m *Mux) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	return m.add("", method, path, handler, middleware...)
}

// Host creates a new router group for the provided host and optional host-level middleware.
func (m *Mux) Host(name string, mf ...MiddlewareFunc) (g *Group) {
	m.routers[name] = NewRouter(m)
	g = &Group{host: name, mux: m}
	g.Use(mf...)
	return
}

// Group creates a new router group with prefix and optional group-level middleware.
func (m *Mux) Group(prefix string, mf ...MiddlewareFunc) (g *Group) {
	g = &Group{prefix: prefix, mux: m}
	g.Use(mf...)
	return
}

// URI generates a URI from handler.
func (m *Mux) URI(handler HandlerFunc, params ...interface{}) string {
	name := handlerName(handler)
	return m.Reverse(name, params...)
}

// URL is an alias for `URI` function.
func (m *Mux) URL(h HandlerFunc, params ...interface{}) string {
	return m.URI(h, params...)
}

// Reverse generates an URL from route name and provided parameters.
func (m *Mux) Reverse(name string, params ...interface{}) string {
	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0
	for _, r := range m.router.routes {
		if r.Name == name {
			for i, l := 0, len(r.Path); i < l; i++ {
				if r.Path[i] == ':' && n < ln {
					for ; i < l && r.Path[i] != '/'; i++ {
					}
					uri.WriteString(fmt.Sprintf("%v", params[n]))
					n++
				}
				if i < l {
					uri.WriteByte(r.Path[i])
				}
			}
			break
		}
	}
	return uri.String()
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (m *Mux) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// Acquire context
	c := m.pool.Get().(*context)
	c.Reset(req, res)
	h := NotFoundHandler

	if m.premiddleware == nil {
		m.findRouter(req.Host).Find(req.Method, req.URL.EscapedPath(), c)
		h = c.Handler()
		h = applyMiddleware(h, m.middleware...)
	} else {
		h = func(c Context) error {
			m.findRouter(req.Host).Find(req.Method, req.URL.EscapedPath(), c)
			h := c.Handler()
			h = applyMiddleware(h, m.middleware...)
			return h(c)
		}
		h = applyMiddleware(h, m.premiddleware...)
	}

	// Execute chain
	if err := h(c); err != nil {
		m.HTTPErrorHandler(err, c)
	}

	// Release context
	m.pool.Put(c)
}

// Start starts an HTTP server.
func (m *Mux) Start(address string) error {
	m.Server.Addr = address
	return m.StartServer(m.Server)
}

// StartTLS starts an HTTPS server.
// If `certFile` or `keyFile` is `string` the values are treated as file paths.
// If `certFile` or `keyFile` is `[]byte` the values are treated as the certificate or key as-is.
func (m *Mux) StartTLS(address string, certFile, keyFile interface{}) (err error) {
	var cert []byte
	if cert, err = filepathOrContent(certFile); err != nil {
		return
	}

	var key []byte
	if key, err = filepathOrContent(keyFile); err != nil {
		return
	}

	s := m.TLSServer
	s.TLSConfig = new(tls.Config)
	s.TLSConfig.Certificates = make([]tls.Certificate, 1)
	if s.TLSConfig.Certificates[0], err = tls.X509KeyPair(cert, key); err != nil {
		return
	}

	return m.startTLS(address)
}

func filepathOrContent(fileOrContent interface{}) (content []byte, err error) {
	switch v := fileOrContent.(type) {
	case string:
		return ioutil.ReadFile(v)
	case []byte:
		return v, nil
	default:
		return nil, ErrInvalidCertOrKeyType
	}
}

// StartAutoTLS starts an HTTPS server using certificates automatically installed from https://letsencrypt.org.
func (m *Mux) StartAutoTLS(address string) error {
	s := m.TLSServer
	s.TLSConfig = new(tls.Config)
	s.TLSConfig.GetCertificate = m.AutoTLSManager.GetCertificate
	s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, acme.ALPNProto)
	return m.startTLS(address)
}

func (m *Mux) startTLS(address string) error {
	s := m.TLSServer
	s.Addr = address
	if !m.DisableHTTP2 {
		s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, "h2")
	}
	return m.StartServer(m.TLSServer)
}

// StartServer starts a custom http server.
func (m *Mux) StartServer(s *http.Server) (err error) {
	// Setup
	m.colorer.SetOutput(m.Logger.Output())
	s.ErrorLog = m.StdLogger
	s.Handler = m
	if m.Debug {
		m.Logger.SetLevel(log.DEBUG)
	}

	if s.TLSConfig == nil {
		if m.Listener == nil {
			m.Listener, err = newListener(s.Addr)
			if err != nil {
				return err
			}
		}
		if !m.HidePort {
			m.colorer.Printf("⇨ http server started on %s\n", m.colorer.Green(m.Listener.Addr()))
		}
		return s.Serve(m.Listener)
	}

	if m.TLSListener == nil {
		l, err := newListener(s.Addr)
		if err != nil {
			return err
		}
		m.TLSListener = tls.NewListener(l, s.TLSConfig)
	}

	return s.Serve(m.TLSListener)
}

// WrapHandler wraps `http.Handler` into `echo.HandlerFunc`.
func WrapHandler(h http.Handler) HandlerFunc {
	return func(c Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

// WrapMiddleware wraps `func(http.Handler) http.Handler` into `echo.MiddlewareFunc`
func WrapMiddleware(m func(http.Handler) http.Handler) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) (err error) {
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.SetRequest(r)
				c.SetResponse(NewResponse(w, c.Mux()))
				err = next(c)
			})).ServeHTTP(c.Response(), c.Request())
			return
		}
	}
}

func (m *Mux) findRouter(host string) *Router {
	if len(m.routers) > 0 {
		if r, ok := m.routers[host]; ok {
			return r
		}
	}
	return m.router
}

func handlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	if c, err = ln.AcceptTCP(); err != nil {
		return
	} else if err = c.(*net.TCPConn).SetKeepAlive(true); err != nil {
		return
	}
	// Ignore error from setting the KeepAlivePeriod as some systems, such as
	// OpenBSD, do not support setting TCP_USER_TIMEOUT on IPPROTO_TCP
	_ = c.(*net.TCPConn).SetKeepAlivePeriod(3 * time.Minute)
	return
}

func newListener(address string) (*tcpKeepAliveListener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &tcpKeepAliveListener{l.(*net.TCPListener)}, nil
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
