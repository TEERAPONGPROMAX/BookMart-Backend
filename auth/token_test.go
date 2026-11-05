package auth

import "testing"

func TestIssuerAcceptsSignedTokenAndRejectsTampering(t *testing.T) {
	issuer := NewIssuer("this-test-secret-is-at-least-thirty-two-bytes-long")
	token, err := issuer.Issue(42, "reader@example.com", "member")
	if err != nil {
		t.Fatal(err)
	}
	id, role, email, err := issuer.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if id != 42 || role != "member" || email != "reader@example.com" {
		t.Fatalf("unexpected claims: %d %s %s", id, role, email)
	}
	forged := token[:len(token)-1] + "x"
	if _, _, _, err := issuer.Parse(forged); err == nil {
		t.Fatal("tampered token was accepted")
	}
}
