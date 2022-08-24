package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/go-kit/log"
	"github.com/observatorium/obsctl/pkg/config"
)

const prefixHeader = "X-Forwarded-Prefix"

// NewProxyServer returns an HTTP reverse proxy server, based on current tenant and API context.
// It also adds a /api/<resource>/v1/<tenant>/ path prefix to every request sent to it.
// For example http://localhost:8080/api/v1/stores becomes https://myobsapi.com/api/metrics/v1/example-tenant/api/v1/stores.
// This makes UIs like Thanos Querier fully functional.
func NewProxyServer(ctx context.Context, logger log.Logger, resource, listenAddr string) (*http.Server, error) {
	cfg, err := config.Read(logger)
	if err != nil {
		return nil, fmt.Errorf("getting reading config: %w", err)
	}

	t, err := cfg.Transport(ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("getting current transport: %w", err)
	}

	apiURL, err := url.Parse(cfg.APIs[cfg.Current.API].URL)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid URL", cfg.APIs[cfg.Current.API].URL)
	}

	// url.Parse might pass a URL with only path, so need to check here for scheme and host.
	// As per docs: https://pkg.go.dev/net/url#Parse.
	if apiURL.Host == "" || apiURL.Scheme == "" {
		return nil, fmt.Errorf("%s is not a valid URL (scheme: %s,host: %s)", apiURL, apiURL.Scheme, apiURL.Host)
	}

	p := httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Scheme = apiURL.Scheme
			// Set the Host at both request and request.URL objects.
			request.Host = apiURL.Host
			request.URL.Host = apiURL.Host
			// Derive path from the paths of configured URL and request URL.
			request.URL.Path, request.URL.RawPath = joinURLPath(apiURL, request.URL, resource, cfg.APIs[cfg.Current.API].Contexts[cfg.Current.Tenant].Tenant)
			request.Header.Add(prefixHeader, "/")
		},
		Transport: t,
	}

	return &http.Server{
		Addr:    listenAddr,
		Handler: &p,
	}, nil
}

func singleJoiningSlash(a, b string) string {
	bslash := strings.HasPrefix(b, "/")

	if bslash {
		return a + b
	} else {
		return a + "/" + b
	}
}

// Modification of
// https://go.dev/src/net/http/httputil/reverseproxy.go#L116
func joinURLPath(a, b *url.URL, resource, tenant string) (string, string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(path.Join(a.Path, "api/"+resource+"/v1/"+tenant), b.Path), ""
	}
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	apath = path.Join(apath, "api/"+resource+"/v1/"+tenant)
	a.Path = path.Join(a.Path, "api/"+resource+"/v1/"+tenant)

	bslash := strings.HasPrefix(bpath, "/")
	if bslash {
		return a.Path + b.Path, apath + bpath
	} else {
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
}
