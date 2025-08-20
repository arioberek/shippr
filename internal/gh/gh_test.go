package gh

import "testing"

func TestSlug(t *testing.T) {
	if Slug("org", "repo") != "org/repo" {
		t.Fatalf("unexpected slug")
	}
}

func TestHumanStrategy(t *testing.T) {
	if HumanStrategy("--rebase") != "rebase" {
		t.Fatalf("rebase not mapped")
	}
	if HumanStrategy("--merge") != "merge" {
		t.Fatalf("merge not mapped")
	}
	if HumanStrategy("--squash") != "squash" {
		t.Fatalf("default squash not mapped")
	}
	if HumanStrategy("") != "squash" {
		t.Fatalf("empty should squash")
	}
}
