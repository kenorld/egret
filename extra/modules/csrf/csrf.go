package csrf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"html/template"
	"io"
	"math"
	"net/url"

	"github.com/kenorld/egret"
)

// allowMethods are HTTP methods that do NOT require a token
var allowedMethods = map[string]bool{
	"GET":     true,
	"HEAD":    true,
	"OPTIONS": true,
	"TRACE":   true,
}

func RandomString(length int) (string, error) {
	buffer := make([]byte, int(math.Ceil(float64(length)/2)))
	if _, err := io.ReadFull(rand.Reader, buffer); err != nil {
		return "", nil
	}
	str := hex.EncodeToString(buffer)
	return str[:length], nil
}

func RefreshToken(c *egret.Context) {
	token, err := RandomString(64)
	if err != nil {
		panic(err)
	}
	c.Session["csrf_token"] = token
}

// CsrfHandler enables CSRF request token creation and verification.
//
// Usage:
//  1) Add `csrf.CsrfHandler` to the app's filters (it must come after the egret.SessionHandler).
//  2) Add CSRF fields to a form with the template tag `{{ csrftoken . }}`. The filter adds a function closure to the `RenderArgs` that can pull out the secret and make the token as-needed, caching the value in the request. Ajax support provided through the `X-CSRFToken` header.
func CsrfHandler(c *egret.Context, fc []egret.Handler) {
	token, foundToken := c.Session["csrf_token"]

	if !foundToken {
		RefreshToken(c)
	}

	referer, refErr := url.Parse(c.Request.Header.Get("Referer"))
	isSameOrigin := sameOrigin(c.Request.URL, referer)

	// If the Request method isn't in the white listed methods
	if !allowedMethods[c.Request.Method] && !IsExempt(c) {
		// Token wasn't present at all
		if !foundToken {
			c.Result = c.Forbidden("EJECT CSRF: Session token missing.")
			return
		}

		// Referer header is invalid
		if refErr != nil {
			c.Result = c.Forbidden("EJECT CSRF: HTTP Referer malformed.")
			return
		}

		// Same origin
		if !isSameOrigin {
			c.Result = c.Forbidden("EJECT CSRF: Same origin mismatch.")
			return
		}

		var requestToken string
		// First check for token in post data
		if c.Request.Method == "POST" {
			requestToken = c.Request.FormValue("csrftoken")
		}

		// Then check for token in custom headers, as with AJAX
		if requestToken == "" {
			requestToken = c.Request.Header.Get("X-CSRFToken")
		}

		if requestToken == "" || !compareToken(requestToken, token) {
			c.Result = c.Forbidden("EJECT CSRF: Invalid token.")
			return
		}
	}

	// Only add token to RenderArgs if the request is: not AJAX, not missing referer header, and is same origin.
	if c.Request.Header.Get("X-CSRFToken") == "" && isSameOrigin {
		c.Set("_csrftoken", token)
	}
}

func compareToken(requestToken, token string) bool {
	// ConstantTimeCompare will panic if the []byte aren't the same length
	if len(requestToken) != len(token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(requestToken), []byte(token)) == 1
}

// Validates same origin policy
func sameOrigin(u1, u2 *url.URL) bool {
	return u1.Scheme == u2.Scheme && u1.Host == u2.Host
}

func init() {
	egret.SharedTemplateFunc["csrftoken"] = func(renderArgs map[string]interface{}) template.HTML {
		if tokenFunc, ok := renderArgs["_csrftoken"]; !ok {
			panic("EJECT CSRF: _csrftoken missing from RenderArgs.")
		} else {
			return template.HTML(tokenFunc.(func() string)())
		}
	}
}
