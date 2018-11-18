package csrf

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kenorld/egret"
)

func TestExemptPath(t *testing.T) {
	MarkExempt("/Context/Action")

	resp := httptest.NewRecorder()
	postRequest, _ := http.NewRequest("POST", "http://www.example.com/Context/Action", nil)
	c := egret.NewContext(egret.NewRequest(postRequest), egret.NewResponse(resp))
	c.Session = make(egret.Session)

	testHandlers[0](c, testHandlers)

	if c.Response.Status == 403 {
		t.Fatal("post to csrf exempt action should pass")
	}
}

func TestExemptPathCaseInsensitive(t *testing.T) {
	MarkExempt("/Context/Action")

	resp := httptest.NewRecorder()
	postRequest, _ := http.NewRequest("POST", "http://www.example.com/controller/action", nil)
	c := egret.NewContext(egret.NewRequest(postRequest), egret.NewResponse(resp))
	c.Session = make(egret.Session)

	testHandlers[0](c, testHandlers)

	if c.Response.Status == 403 {
		t.Fatal("post to csrf exempt action should pass")
	}
}

func TestExemptAction(t *testing.T) {
	MarkExempt("Context.Action")

	resp := httptest.NewRecorder()
	postRequest, _ := http.NewRequest("POST", "http://www.example.com/Context/Action", nil)
	c := egret.NewContext(egret.NewRequest(postRequest), egret.NewResponse(resp))
	c.Session = make(egret.Session)
	c.Action = "Context.Action"

	testHandlers[0](c, testHandlers)

	if c.Response.Status == 403 {
		t.Fatal("post to csrf exempt action should pass")
	}
}
