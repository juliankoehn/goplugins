package session

import (
	"bufio"
	"bytes"
	"goplugins/core/framework/session/memstore"
	"log"
	"net"
	"net/http"
	"time"
)

type (
	// Session holds the configuration settings for your sessions.
	Session struct {
		// IdleTimeout controls the maximum length of time a session can be inactive
		// before it expires. For example, some applications may wish to set this so
		// there is a timeout after 20 minutes of inactivity. By default IdleTimeout
		// is not set and there is no inactivity timeout.
		IdleTimeout time.Duration

		// Lifetime controls the maximum length of time that a session is valid for
		// before it expires. The lifetime is an 'absolute expiry' which is set when
		// the session is first created and does not change. The default value is 24
		// hours.
		Lifetime time.Duration

		// Store controls the session store where the session data is persisted.
		Store Store

		// Cookie contains the configuration settings for session cookies.
		Cookie Cookie

		// Codec controls the encoder/decoder used to transform session data to a
		// byte slice for use by the session store. By default session data is
		// encoded/decoded using encoding/gob.
		Codec Codec

		// ErrorFunc allows you to control behavior when an error is encountered by
		// the LoadAndSave middleware. The default behavior is for a HTTP 500
		// "Internal Server Error" message to be sent to the client and the error
		// logged using Go's standard logger. If a custom ErrorFunc is set, then
		// control will be passed to this instead. A typical use would be to provide
		// a function which logs the error and returns a customized HTML error page.
		ErrorFunc func(http.ResponseWriter, *http.Request, error)

		// contextKey is the key used to set and retrieve the session data from a
		// context.Context. It's automatically generated to ensure uniqueness.
		contextKey contextKey
	}

	// Cookie contains the configuration settings for session cookies.
	Cookie struct {
		// Name sets the name of the session cookie. It should not contain
		// whitespace, commas, colons, semicolons, backslashes, the equals sign or
		// control characters as per RFC6265. The default cookie name is "session".
		// If your application uses two different sessions, you must make sure that
		// the cookie name for each is unique.
		Name string

		// Domain sets the 'Domain' attribute on the session cookie. By default
		// it will be set to the domain name that the cookie was issued from.
		Domain string

		// HTTPOnly sets the 'HttpOnly' attribute on the session cookie. The
		// default value is true.
		HTTPOnly bool

		// Path sets the 'Path' attribute on the session cookie. The default value
		// is "/". Passing the empty string "" will result in it being set to the
		// path that the cookie was issued from.
		Path string

		// Persist sets whether the session cookie should be persistent or not
		// (i.e. whether it should be retained after a user closes their browser).
		// The default value is true, which means that the session cookie will not
		// be destroyed when the user closes their browser and the appropriate
		// 'Expires' and 'MaxAge' values will be added to the session cookie.
		Persist bool

		// SameSite controls the value of the 'SameSite' attribute on the session
		// cookie. By default this is set to 'SameSite=Lax'. If you want no SameSite
		// attribute or value in the session cookie then you should set this to 0.
		SameSite http.SameSite

		// Secure sets the 'Secure' attribute on the session cookie. The default
		// value is false. It's recommended that you set this to true and serve all
		// requests over HTTPS in production environments.
		// See https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Session_Management_Cheat_Sheet.md#transport-layer-security.
		Secure bool
	}

	// Store is the interface for session stores.
	Store interface {
		// Delete should remove the session token and corresponding data from the
		// session store. If the token does not exist then Delete should be a no-op
		// and return nil (not an error).
		Delete(token string) (err error)

		// Find should return the data for a session token from the store. If the
		// session token is not found or is expired, the found return value should
		// be false (and the err return value should be nil). Similarly, tampered
		// or malformed tokens should result in a found return value of false and a
		// nil err value. The err return value should be used for system errors only.
		Find(token string) (b []byte, found bool, err error)

		// Commit should add the session token and data to the store, with the given
		// expiry time. If the session token already exists, then the data and
		// expiry time should be overwritten.
		Commit(token string, b []byte, expiry time.Time) (err error)
	}
)

// New returns a new session manager with the default options. It is safe for
// concurrent use.
func New() *Session {
	s := &Session{
		IdleTimeout: 0,
		Lifetime:    24 * time.Hour,
		Store:       memstore.New(),
		Codec:       GobCodec{},
		ErrorFunc:   defaultErrorFunc,
		contextKey:  generateContextKey(),
		Cookie: Cookie{
			Name:     "session",
			Domain:   "",
			HTTPOnly: true,
			Path:     "/",
			Persist:  true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		},
	}
	return s
}

// LoadAndSave provides middleware which automatically loads and saves session
// data for the current request, and communicates the session token to and from
// the client in a cookie.
func (s *Session) LoadAndSave(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		cookie, err := r.Cookie(s.Cookie.Name)
		if err == nil {
			token = cookie.Value
		}

		ctx, err := s.Load(r.Context(), token)
		if err != nil {
			s.ErrorFunc(w, r, err)
			return
		}

		sr := r.WithContext(ctx)
		bw := &bufferedResponseWriter{ResponseWriter: w}
		next.ServeHTTP(bw, sr)

		if sr.MultipartForm != nil {
			sr.MultipartForm.RemoveAll()
		}

		switch s.Status(ctx) {
		case Modified:
			token, expiry, err := s.Commit(ctx)
			if err != nil {
				s.ErrorFunc(w, r, err)
				return
			}
			s.writeSessionCookie(w, token, expiry)
		case Destroyed:
			s.writeSessionCookie(w, "", time.Time{})
		}

		if bw.code != 0 {
			w.WriteHeader(bw.code)
		}
		w.Write(bw.buf.Bytes())
	})
}

func (s *Session) writeSessionCookie(w http.ResponseWriter, token string, expiry time.Time) {
	cookie := &http.Cookie{
		Name:     s.Cookie.Name,
		Value:    token,
		Path:     s.Cookie.Path,
		Domain:   s.Cookie.Domain,
		Secure:   s.Cookie.Secure,
		HttpOnly: s.Cookie.HTTPOnly,
		SameSite: s.Cookie.SameSite,
	}

	if expiry.IsZero() {
		cookie.Expires = time.Unix(1, 0)
		cookie.MaxAge = -1
	} else if s.Cookie.Persist {
		cookie.Expires = time.Unix(expiry.Unix()+1, 0)        // Round up to the nearest second.
		cookie.MaxAge = int(time.Until(expiry).Seconds() + 1) // Round up to the nearest second.
	}

	w.Header().Add("Set-Cookie", cookie.String())
	addHeaderIfMissing(w, "Cache-Control", `no-cache="Set-Cookie"`)
	addHeaderIfMissing(w, "Vary", "Cookie")

}

func addHeaderIfMissing(w http.ResponseWriter, key, value string) {
	for _, h := range w.Header()[key] {
		if h == value {
			return
		}
	}
	w.Header().Add(key, value)
}

func defaultErrorFunc(w http.ResponseWriter, r *http.Request, err error) {
	log.Output(2, err.Error())
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

type bufferedResponseWriter struct {
	http.ResponseWriter
	buf         bytes.Buffer
	code        int
	wroteHeader bool
}

func (bw *bufferedResponseWriter) Write(b []byte) (int, error) {
	return bw.buf.Write(b)
}

func (bw *bufferedResponseWriter) WriteHeader(code int) {
	if !bw.wroteHeader {
		bw.code = code
		bw.wroteHeader = true
	}
}

func (bw *bufferedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := bw.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

func (bw *bufferedResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := bw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}
