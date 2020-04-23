package httpmux

import (
	"aicode"
	"fmt"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
)

// HTTP methods
const (
	CONNECT  = "CONNECT"
	DELETE   = "DELETE"
	GET      = "GET"
	HEAD     = "HEAD"
	OPTIONS  = "OPTIONS"
	PATCH    = "PATCH"
	POST     = "POST"
	PROPFIND = "PROPFIND"
	PUT      = "PUT"
	TRACE    = "TRACE"
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

const (
	charsetUTF8 = "charset=UTF-8"
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
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)

type server struct {
	router *httprouter.Router
	pool   sync.Pool
}

func New() *server {
	s := new(server)
	s.router = httprouter.New()
	s.pool.New = func() interface{} {
		return s.NewContext(nil, nil)
	}
	return s
}

type Handle func(c Context) aicode.HTTPError

func (r *server) NewContext(req *http.Request, rsp http.ResponseWriter) Context {
	return &context{
		request:  req,
		response: NewResponse(rsp, r),
		store:    make(map[string]interface{}),
	}
}
func (r *server) warpFunc(h Handle) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(rsp http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		c := r.pool.Get().(*context)
		c.Reset(req, rsp)
		defer func() {
			if rc := recover(); rc != nil {
				c.JSON(http.StatusOK, aicode.NewHTTPError(aicode.ComInnerError.Code(), fmt.Sprint(rc)))
			}
		}()
		err := h(c)
		if err != nil {
			c.JSON(http.StatusOK, err)
		}

	}
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *server) GET(path string, handle Handle) {
	r.router.Handle("GET", path, r.warpFunc(handle))
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *server) HEAD(path string, handle Handle) {
	r.router.Handle("HEAD", path, r.warpFunc(handle))
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *server) OPTIONS(path string, handle Handle) {
	r.router.Handle("OPTIONS", path, r.warpFunc(handle))
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *server) POST(path string, handle Handle) {
	r.router.Handle("POST", path, r.warpFunc(handle))
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *server) PUT(path string, handle Handle) {
	r.router.Handle("PUT", path, r.warpFunc(handle))
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *server) PATCH(path string, handle Handle) {
	r.router.Handle("PATCH", path, r.warpFunc(handle))
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *server) DELETE(path string, handle Handle) {
	r.router.Handle("DELETE", path, r.warpFunc(handle))
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *server) Handle(method, path string, handle Handle) {
	r.router.Handle(method, path, r.warpFunc(handle))
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}
