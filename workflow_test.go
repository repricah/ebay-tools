package ebaytools

import (
	"os"
	"strings"
	"testing"
)

func TestCIWorkflowUsesSelfHostedRunner(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(".github/workflows/ci.yml")
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}

	content := string(data)
	for _, needle := range []string{
		`name: CI`,
		`pull_request:`,
		`push:`,
		`branches: ["main"]`,
		`workflow_dispatch:`,
		`runs-on: [self-hosted, gha-runner]`,
		`actions/checkout@v4`,
		`actions/setup-go@v5`,
		`go test ./... -count=1`,
	} {
		if !strings.Contains(content, needle) {
			t.Fatalf("workflow missing %q", needle)
		}
	}
}
