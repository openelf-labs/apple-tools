# Contributing to apple-tools

## Requirements

- macOS 13.0+ (Ventura or later)
- Go 1.22+
- No additional dependencies required

## Getting Started

```bash
git clone https://github.com/openelf-labs/apple-tools.git
cd apple-tools
go test ./...                                          # Run all tests
go run ./cmd/apple-tools-demo list                     # List all tools
go run ./cmd/apple-tools-demo call apple_system_battery # Test a tool
```

You do NOT need OpenELF installed. apple-tools is fully standalone.

## Adding a New Tool

Each tool requires 3 files (Go handler + JXA script + test):

### 1. JXA Script (`category/scripts/tool_name.js`)

```javascript
ObjC.import("Foundation");

// Read parameters from environment variable (NEVER concatenate into script)
var env = $.NSProcessInfo.processInfo.environment.objectForKey("APPLE_TOOLS_PARAMS");
var params = env ? JSON.parse(env.js) : {};

// Your implementation here
var result = { /* ... */ };

// Always output JSON
JSON.stringify(result);
```

**Security rules:**
- MUST read parameters via `APPLE_TOOLS_PARAMS` environment variable
- MUST NOT concatenate user input into script strings
- MUST output valid JSON via `JSON.stringify()`
- MUST handle missing/null parameters gracefully

### 2. Go Handler (`category/category.go`)

```go
//go:build darwin

package category

import (
    "context"
    "embed"
    "encoding/json"
    "fmt"

    "github.com/openelf-labs/apple-tools/core"
)

//go:embed scripts/*.js
var scripts embed.FS

func Register(r core.Registry) {
    r.Add(core.Tool{
        Name:        "apple_category_action",
        Description: "Clear description for the LLM. Include what it does and what it returns.",
        Parameters:  json.RawMessage(`{
            "type":"object",
            "properties":{
                "param1":{"type":"string","description":"What this param does"}
            },
            "required":["param1"]
        }`),
        Handler: handleAction,
    })
}

func handleAction(ctx context.Context, input json.RawMessage) (string, error) {
    var params struct {
        Param1 string `json:"param1"`
    }
    if err := json.Unmarshal(input, &params); err != nil {
        return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
    }
    // Validate
    if params.Param1 == "" {
        return "", fmt.Errorf("%w: param1 is required", core.ErrInvalidInput)
    }
    // Execute
    script, _ := scripts.ReadFile("scripts/action.js")
    result, err := core.RunJXA(ctx, script, params)
    if err != nil {
        return "", err
    }
    // Format for LLM
    return string(result), nil
}
```

### 3. Test (`category/category_test.go`)

```go
//go:build darwin

package category

import (
    "encoding/json"
    "testing"

    "github.com/openelf-labs/apple-tools/testutil"
)

func TestRegister(t *testing.T) {
    reg := &testutil.MockRegistry{}
    Register(reg)
    if len(reg.Tools) != 1 {
        t.Fatalf("expected 1 tool, got %d", len(reg.Tools))
    }
    if !json.Valid(reg.Tools[0].Parameters) {
        t.Error("invalid JSON schema")
    }
}

func TestValidation(t *testing.T) {
    reg := &testutil.MockRegistry{}
    Register(reg)
    _, err := testutil.CallTool(t, reg, "apple_category_action", map[string]any{"param1": ""})
    if err == nil {
        t.Error("expected validation error")
    }
}
```

### 4. Update apple.go

Add your category to the `RegisterAll` function and import.

### 5. Update Config

Add an enable flag to `config.go`'s `Config` struct.

### 6. Update README

Add your tool to the Data Access Declaration table in README.md.

## Naming Conventions

- Tool names: `apple_{category}_{action}` (e.g., `apple_calendar_list`)
- JXA scripts: `{action}.js` (e.g., `list_events.js`)
- Go packages: lowercase category name

## Testing

```bash
go test ./...                    # Unit tests (parameter validation, error classification)
go test ./... -short             # Skip slow integration tests
go test ./category/ -v           # Verbose output for one module
go vet ./...                     # Static analysis
```

Integration tests (marked with `if testing.Short() { t.Skip() }`) actually call osascript and interact with macOS apps. They are safe (read-only operations) but may trigger permission prompts on first run.

## Code Review Checklist

- [ ] JXA script reads params from env var, not string concatenation
- [ ] Go handler validates all required fields
- [ ] Tool description is clear enough for an LLM to use correctly
- [ ] JSON Schema in Parameters is valid
- [ ] Tests cover: registration, parameter validation, at least one integration test
- [ ] Data Access Declaration updated in README.md
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
