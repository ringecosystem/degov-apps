package middleware

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/ringecosystem/degov-apps/services"
	"github.com/ringecosystem/degov-apps/types"
)

const (
	DegovDaocodeKey ContextKey = "degov_daocode"
)

type DegovMiddleware struct {
	daoService *services.DaoService
	daoCache   *cache.Cache
}

func NewDegovMiddleware() *DegovMiddleware {
	// Create cache with 2 minutes TTL and 5 minutes cleanup interval
	c := cache.New(2*time.Minute, 5*time.Minute)

	return &DegovMiddleware{
		daoService: services.NewDaoService(),
		daoCache:   c,
	}
}

// Middleware returns a standard middleware function
func (m *DegovMiddleware) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return m.HTTPMiddleware(next)
	}
}

func (m *DegovMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originHeader := r.Header.Get("Origin")
		refererHeader := r.Header.Get("Referer")
		siteHeader := r.Header.Get("x-degov-site")
		daocodeHeader := r.Header.Get("x-degov-daocode")

		var daoCode string

		// if daocodeHeader no empty, use it
		if daocodeHeader != "" {
			daoCode = daocodeHeader
		} else {
			daoCode = m.findDaoCodeByURL(siteHeader, originHeader, refererHeader)
		}

		if daoCode != "" {
			ctx := context.WithValue(r.Context(), DegovDaocodeKey, daoCode)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// try to find the DAO code based on the origin or referer headers
func (m *DegovMiddleware) findDaoCodeByURL(customSite, origin, referer string) string {
	// use origin if available, otherwise use referer
	targetURL := customSite
	if targetURL == "" {
		targetURL = origin
	}
	if targetURL == "" {
		targetURL = referer
	}

	if targetURL == "" {
		return ""
	}

	targetHost := m.extractHost(targetURL)
	if targetHost == "" {
		return ""
	}

	// get DAOs from cache, if not found, fetch from database and cache
	daos := m.getDaosFromCache()
	if daos == nil {
		return ""
	}

	// match DAO by endpoint
	for _, dao := range daos {
		if dao.Endpoint != "" {
			if strings.EqualFold(dao.Endpoint, targetHost) {
				return dao.Code
			}
		}
	}

	return ""
}

// extractHost extracts the host from a given URL string
func (m *DegovMiddleware) extractHost(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	// if url does not contain a protocol, add default https
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return parsedURL.Host
}

// from cache, if not found, fetch from database and cache
func (m *DegovMiddleware) getDaosFromCache() []DaoEndpoint {
	// try to get from cache
	if cached, found := m.daoCache.Get("daos"); found {
		if daos, ok := cached.([]DaoEndpoint); ok {
			return daos
		}
	}

	// Fetch DAOs from the database
	gqlDaos, err := m.daoService.GetDaos(types.BasicInput[*string]{})
	if err != nil {
		return nil
	}

	// Simplified DAO information structure
	var daos []DaoEndpoint
	for _, gqlDao := range gqlDaos {
		dao := DaoEndpoint{
			Code:     gqlDao.Code,
			Endpoint: m.extractHost(gqlDao.Endpoint),
		}
		daos = append(daos, dao)
	}

	// store in cache with 2 minutes TTL
	m.daoCache.Set("daos", daos, 2*time.Minute)

	return daos
}

// DaoEndpoint represents a simplified DAO information structure
type DaoEndpoint struct {
	Code     string
	Endpoint string
}
