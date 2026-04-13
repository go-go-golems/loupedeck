package device

import "testing"

func TestResolveProfileReturnsKnownDevice(t *testing.T) {
	profile, err := resolveProfile("0004")
	if err != nil {
		t.Fatalf("resolveProfile returned error: %v", err)
	}
	if profile.Name != "Loupedeck Live" {
		t.Fatalf("expected Loupedeck Live profile, got %q", profile.Name)
	}
	if len(profile.Displays) != 3 {
		t.Fatalf("expected 3 displays, got %d", len(profile.Displays))
	}
}

func TestResolveProfileRejectsUnknownProduct(t *testing.T) {
	if _, err := resolveProfile("ffff"); err == nil {
		t.Fatalf("expected unknown product to return error")
	}
}

func TestApplyProfilePopulatesDisplaysAndModel(t *testing.T) {
	profile, err := resolveProfile("0007")
	if err != nil {
		t.Fatalf("resolveProfile returned error: %v", err)
	}

	l := &Loupedeck{displays: map[string]*Display{}}
	l.applyProfile(profile)

	if l.Model != "Loupedeck CT v2" {
		t.Fatalf("expected model to be populated from profile, got %q", l.Model)
	}
	for _, name := range []string{"left", "main", "right", "all", "dial"} {
		if l.GetDisplay(name) == nil {
			t.Fatalf("expected display %q to be configured", name)
		}
	}
}
