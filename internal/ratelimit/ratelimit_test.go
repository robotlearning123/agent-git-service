package ratelimit

import (
	"net/http/httptest"
	"testing"
)

func TestSubjectForRequest_TreatsEmbeddedActorsAsAuthenticated(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/api/v3/user", nil)
	req = req.WithContext(WithActor(req.Context(), "embedded:meshx:subject-1"))

	subject := SubjectForRequest(req)
	if !subject.Authenticated {
		t.Fatal("expected embedded actor to be treated as authenticated")
	}
	if got := subject.Actor; got != "embedded:meshx:subject-1" {
		t.Fatalf("actor: got %q want %q", got, "embedded:meshx:subject-1")
	}
}

func TestSubjectForRequestIgnoresSpoofableForwardedHeaders(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/api/v3/user", nil)
	req.RemoteAddr = "198.51.100.10:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 203.0.113.2")
	req.Header.Set("X-Real-IP", "203.0.113.3")

	subject := SubjectForRequest(req)
	if got := subject.Actor; got != "ip:198.51.100.10" {
		t.Fatalf("actor: got %q want %q", got, "ip:198.51.100.10")
	}
}
