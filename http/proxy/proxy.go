// Pacakge proxy provides a reverse proxy for talking to Apple AxM APIs.
// Based on the standard Go reverse proxy.
package proxy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/micromdm/nanoaxm/client"

	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanolib/log/ctxlog"
)

// New creates a new NanoAxM reverse proxy. This proxy will dispatch
// requests using transport (which should be a NanoAxM RoundTripper which
// handles the OAuth 2 component). AxM names are assumed to already be
// in the context, likely using [NewNameMiddleware].
func New(transport http.RoundTripper, apiURL string, logger log.Logger) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport:    transport,
		Director:     newDirector(apiURL, logger.With("function", "director")),
		ErrorHandler: newErrorHandler(logger.With("msg", "proxy error")),
	}
}

// newErrorHandler creates a new function for ReverseProxy.ErrorHandler.
func newErrorHandler(logger log.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(rw http.ResponseWriter, req *http.Request, err error) {
		// use the same error as the standrad reverse proxy
		rw.WriteHeader(http.StatusBadGateway)

		logger := ctxlog.Logger(req.Context(), logger)

		var httpErr *client.HTTPError
		if errors.As(err, &httpErr) {
			logger.Info(
				"err", "HTTP error",
				"status", httpErr.Status,
				"body", string(httpErr.Body),
			)
			// write the same body content to try and give some clue of what
			// happened to the proxy user
			rw.Write([]byte(httpErr.Body))
			return
		}

		logger.Info("err", err)
	}
}

// newDirector creates a new [*httputil.ReverseProxy] director which replaces
// the target URL and host parameters in the request. The baseURL param
// should point to the target URL.
func newDirector(baseURL string, logger log.Logger) func(*http.Request) {
	if logger == nil {
		panic("nil logger")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	return func(req *http.Request) {
		name := client.GetName(req.Context())
		if name == "" {
			ctxlog.Logger(req.Context(), logger).Info("err", "missing AxM name")
			// this will probably lead to a very broken proxy.
			// but we can't really do anything about it here.
			return
		}

		// perform our actual request modifications (i.e. swapping in the
		// correct AxM URL components based on the context)
		req.URL.Scheme = base.Scheme
		req.URL.Host = base.Host
		req.Host = base.Host
	}
}

// newCopiedRequest makes a copy of r with a new copy of r.URL and returns it.
func newCopiedRequest(r *http.Request) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2.URL = new(url.URL)
	*r2.URL = *r.URL
	return r2
}

// NewNameMiddleware treats the first "/"-separated URL path as the AxM name,
// removes it from the URL path, and injects it as a context value.
//
// For example if the request URL path is "hello/world/" then "hello" is the
// AxM name and is set in the request context and "/world/" is then set in the
// HTTP path request passed onto h.
//
// Note the very beginning of the URL path is used as the AxM name. This
// necessitates stripping the URL prefix before using this handler. Note also
// that AxM names with a "/" or "%2F" may cause issues as we naively
// search and cut by "/" in the path.
func NewNameMiddleware(h http.Handler, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r2 := newCopiedRequest(r)

		name, endpoint, found := CutIncl(r.URL.Path, "/")
		if found {
			r2.URL.Path = endpoint
		}

		logger := ctxlog.Logger(r.Context(), logger)

		if name == "" {
			logger.Info("msg", "extracting AxM name", "err", "AxM name not found in path")
			http.NotFound(w, r)
			return
		}

		// try to perform the same extraction on the RawPath as we did for Path
		if r.URL.RawPath != "" {
			if _, endpoint, found = CutIncl(r.URL.RawPath, "/"); found {
				r2.URL.RawPath = endpoint
			}
		}

		logger.Debug("msg", "proxy serve", "name", name, "endpoint", endpoint)

		h.ServeHTTP(w, r2.WithContext(client.WithName(r2.Context(), name)))
	}
}

// CutIncl is like strings.Cut but keeps sep in after.
func CutIncl(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i:], true
	}
	return s, "", false
}
