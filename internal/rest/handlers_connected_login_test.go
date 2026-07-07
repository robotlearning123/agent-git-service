package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ngaut/agent-git-service/internal/service"
)

func TestConnectedCallbackRequiresStateCookieBeforeCodeExchange(t *testing.T) {
	t.Parallel()

	deps := &Deps{Svc: &service.Service{}}
	req := httptest.NewRequest(http.MethodGet, "/auth/connected/callback?code=oauth-code&state=csrf-state", nil)
	rec := httptest.NewRecorder()

	deps.ConnectedCallback(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !strings.Contains(body, "invalid or missing state") {
		t.Fatalf("expected missing state error, got %q", body)
	}
}
