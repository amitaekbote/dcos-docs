package cmd

import "testing"

// TestComponentCheckGetHealthURL validates that parameters Role and ForceTLS
// return the expected URL based on adminrouter / 3dt configuration.
func TestComponentCheckGetHealthURL(t *testing.T) {
	c := &ComponentCheck{
		Name: "TEST",
	}

	for _, item := range []struct {
		role     string
		forceTLS bool
		expected string
	}{
		{
			role:     "master",
			forceTLS: false,
			expected: "http://127.0.0.1:1050/",
		},
		{
			role:     "master",
			forceTLS: true,
			expected: "https://127.0.0.1:443/",
		},
		{
			role:     "agent",
			forceTLS: false,
			expected: "http://127.0.0.1:61001/",
		},
		{
			role:     "agent",
			forceTLS: true,
			expected: "https://127.0.0.1:61002/",
		},
		{
			role:     "agent_public",
			forceTLS: false,
			expected: "http://127.0.0.1:61001/",
		},
		{
			role:     "agent_public",
			forceTLS: true,
			expected: "https://127.0.0.1:61002/",
		},
	} {
		mockCLICfg := &CLIConfigFlags{
			NodeIPStr: "127.0.0.1",
			Role:      item.role,
			ForceTLS:  item.forceTLS,
		}

		url, err := c.getHealthURL("/", mockCLICfg)
		if err != nil {
			t.Fatalf("Error running getHealthURL: %s", err)
		}

		if url.String() != item.expected {
			t.Fatalf("Expect %s. Got %s", item.expected, url.String())
		}
	}
}
