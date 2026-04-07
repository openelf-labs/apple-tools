//go:build darwin

package contacts

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/openelf-labs/apple-tools/core"
)

//go:embed scripts/*.js
var scripts embed.FS

func mustLoad(name string) []byte {
	data, err := scripts.ReadFile("scripts/" + name)
	if err != nil {
		panic(fmt.Sprintf("contacts: embedded script %q not found: %v", name, err))
	}
	return data
}

var (
	scriptSearch      = mustLoad("search.js")
	scriptGetDetails  = mustLoad("get_details.js")
	scriptFindByPhone = mustLoad("find_by_phone.js")
)

// phonePattern matches a reasonable phone number: optional +, then digits, spaces, dashes, parens.
var phonePattern = regexp.MustCompile(`^\+?[\d\s\-().]{3,20}$`)

// Register adds all Contacts tools to the registry.
func Register(r core.Registry) {
	r.Add(core.Tool{
		Name: "contacts_search",
		Description: "Search contacts in macOS Contacts app by name. Omit query to list all contacts. Returns matching contacts with name, phone numbers, emails, organization, and job title.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {
					"type": "string",
					"description": "Name or partial name to search for. Omit or empty to list all contacts."
				},
				"limit": {
					"type": "integer",
					"description": "Maximum number of results (default 25)",
					"default": 25
				}
			},
			"additionalProperties": false
		}`),
		Handler: handleSearch,
	})

	r.Add(core.Tool{
		Name: "contacts_details",
		Description: "Get detailed information for a contact by exact name. Returns full contact details including phones, emails, addresses, organization, job title, birthday, and notes.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"name": {
					"type": "string",
					"description": "Exact name of the contact to look up"
				}
			},
			"required": ["name"],
			"additionalProperties": false
		}`),
		Handler: handleGetDetails,
	})

	r.Add(core.Tool{
		Name: "contacts_find_by_phone",
		Description: "Find a contact by phone number. Normalizes the number and tries multiple variants (with/without country code) to find a match.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"phoneNumber": {
					"type": "string",
					"description": "Phone number to search for (any common format)"
				}
			},
			"required": ["phoneNumber"],
			"additionalProperties": false
		}`),
		Handler: handleFindByPhone,
	})
}

func handleSearch(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	params.Query = strings.TrimSpace(params.Query)
	// Empty query = list all contacts (up to limit)
	params.Limit = core.ClampLimit(params.Limit, 25, 200)

	result, err := core.RunJXA(ctx, scriptSearch, params)
	if err != nil {
		return "", classifyContactsError(err)
	}
	return string(result), nil
}

func handleGetDetails(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	params.Name = strings.TrimSpace(params.Name)
	if params.Name == "" {
		return "", fmt.Errorf("%w: name must not be empty", core.ErrInvalidInput)
	}

	result, err := core.RunJXA(ctx, scriptGetDetails, params)
	if err != nil {
		return "", classifyContactsError(err)
	}
	return string(result), nil
}

func handleFindByPhone(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		PhoneNumber string `json:"phoneNumber"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrInvalidInput, err)
	}

	params.PhoneNumber = strings.TrimSpace(params.PhoneNumber)
	if params.PhoneNumber == "" {
		return "", fmt.Errorf("%w: phoneNumber must not be empty", core.ErrInvalidInput)
	}
	if !phonePattern.MatchString(params.PhoneNumber) {
		return "", fmt.Errorf("%w: phoneNumber has an unreasonable format", core.ErrInvalidInput)
	}

	result, err := core.RunJXA(ctx, scriptFindByPhone, params)
	if err != nil {
		return "", classifyContactsError(err)
	}
	return string(result), nil
}

// classifyContactsError wraps errors with Contacts-specific context.
func classifyContactsError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, core.ErrPermissionDenied) {
		return core.NewPermissionError("Contacts", "Contacts")
	}
	return err
}
