package transform_test

import (
	"testing"

	"github.com/ngaut/agent-git-service/internal/db"
	"github.com/ngaut/agent-git-service/internal/rest/transform"
)

// TestTokenListItem_OmitsSecret pins the security contract for token listing:
// the cleartext token value is returned only from the mint endpoints (Token),
// never from list responses (TokenListItem). Regression guard for a leak where
// GET /api/v3/user/tokens returned every token's usable secret in plaintext.
func TestTokenListItem_OmitsSecret(t *testing.T) {
	tok := db.Token{ID: 7, Name: "ci-token", Value: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}

	created := transform.Token(tok)
	if created["token"] != tok.Value {
		t.Fatalf("Token() must return the cleartext value on mint; got %v", created["token"])
	}

	listed := transform.TokenListItem(tok)
	if _, present := listed["token"]; present {
		t.Fatalf("TokenListItem() must not expose the secret; got token=%v", listed["token"])
	}
	// Non-secret metadata must still be present so clients can identify/revoke.
	if listed["id"] != tok.ID {
		t.Errorf("TokenListItem() missing id: got %v", listed["id"])
	}
	if listed["name"] != tok.Name {
		t.Errorf("TokenListItem() missing name: got %v", listed["name"])
	}
	if _, ok := listed["created_at"]; !ok {
		t.Errorf("TokenListItem() missing created_at")
	}
}
