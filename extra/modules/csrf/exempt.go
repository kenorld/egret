package csrf

import (
	"fmt"
	"strings"

	"github.com/kenorld/egret"
)

var (
	exemptPath   = make(map[string]bool)
	exemptAction = make(map[string]bool)
)

func MarkExempt(route string) {
	if strings.HasPrefix(route, "/") {
		// e.g. "/controller/action"
		exemptPath[strings.ToLower(route)] = true
	} else if routeParts := strings.Split(route, "."); len(routeParts) == 2 {
		// e.g. "ContextName.ActionName"
		exemptAction[route] = true
	} else {
		err := fmt.Sprintf("csrf.MarkExempt() received invalid argument \"%v\". Either provide a path prefixed with \"/\" or controller action in the form of \"ContextName.ActionName\".", route)
		panic(err)
	}
}

func IsExempt(c *egret.Context) bool {
	if _, ok := exemptPath[strings.ToLower(c.Request.Request.URL.Path)]; ok {
		return true
	} else if _, ok := exemptAction[c.Action]; ok {
		return true
	}

	return false
}
