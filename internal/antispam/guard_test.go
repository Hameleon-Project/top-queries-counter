package antispam

import (
	"strconv"
	"testing"
)

func TestGuard_UserQueryCooldown(t *testing.T) {
	g := New(5, 100, 100)
	ts := int64(1_000_000)
	if !g.Allow("u1", "iphone", "1.1.1.1", ts) {
		t.Fatal("first event must pass")
	}
	if g.Allow("u1", "iphone", "1.1.1.1", ts+1) {
		t.Fatal("duplicate within cooldown must be rejected")
	}
	if !g.Allow("u1", "iphone", "1.1.1.1", ts+10) {
		t.Fatal("after cooldown must pass")
	}
}

func TestGuard_IPRateLimit(t *testing.T) {
	g := New(1, 3, 10_000)
	ts := int64(1_000_000)
	for i := 0; i < 3; i++ {
		if !g.Allow("user"+strconv.Itoa(i), "q", "9.9.9.9", ts) {
			t.Fatalf("event %d should pass", i)
		}
	}
	if g.Allow("ux", "q2", "9.9.9.9", ts) {
		t.Fatal("4th event from same IP in one minute must be rejected")
	}
}

func TestGuard_QuerySpike(t *testing.T) {
	g := New(1, 10_000, 2)
	ts := int64(2_000_000)
	if !g.Allow("a", "spike", "1.1.1.1", ts) {
		t.Fatal("first")
	}
	if !g.Allow("b", "spike", "2.2.2.2", ts) {
		t.Fatal("second")
	}
	if g.Allow("c", "spike", "3.3.3.3", ts) {
		t.Fatal("query spike must be rejected")
	}
}
