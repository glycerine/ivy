package control

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBDDFeatureFilesCoverRequiredFlows(t *testing.T) {
	required := map[string][]string{
		"signup.feature": {
			"Feature: Sign-up",
			"Scenario: New user signs up and verifies email",
			"Scenario: Duplicate sign-up does not create duplicate product users",
		},
		"login.feature": {
			"Feature: Login",
			"Scenario: Existing user logs in and sees authorized projects",
			"Scenario: Wrong password does not create an app session",
		},
		"account_recovery.feature": {
			"Feature: Account recovery",
			"Scenario: User resets password from a recovery email",
			"Scenario: Unknown email receives the same neutral response",
		},
		"billing.feature": {
			"Feature: Billing information",
			"Scenario: Account owner adds billing information",
			"Scenario: Billing admin can update billing but receives no project access",
		},
		"alpha_dashboard.feature": {
			"Feature: Alpha tester dashboard",
			"Scenario: Admin invites an alpha tester and grants project access",
			"Scenario: Invited alpha tester accepts invite and opens project",
		},
	}
	for name, snippets := range required {
		path := filepath.Join("features", name)
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(body)
		for _, snippet := range snippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s missing %q", path, snippet)
			}
		}
	}
}
