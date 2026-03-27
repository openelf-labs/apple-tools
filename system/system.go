//go:build darwin

package system

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openelf-labs/apple-tools/core"
)

func Register(r core.Registry) {
	r.Add(core.Tool{
		Name:        "apple_system_battery",
		Description: "Get battery status. Returns JSON {source, percentage, charging, available, raw}.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler:     handleBattery,
	})

	r.Add(core.Tool{
		Name:        "apple_system_disk",
		Description: "Get disk usage for a path. Returns JSON {path, filesystem, size, used, available, capacity}.",
		Parameters: json.RawMessage(`{
			"type":"object",
			"properties":{
				"path":{"type":"string","description":"Filesystem path to check (default: /)"}
			}
		}`),
		Handler: handleDisk,
	})

	r.Add(core.Tool{
		Name:        "apple_system_network",
		Description: "Get network status. Returns JSON {wifi, interfaces: [{name, ip}], dns}.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		Handler:     handleNetwork,
	})
}

func handleBattery(ctx context.Context, input json.RawMessage) (string, error) {
	out, err := core.RunCommand(ctx, "pmset", "-g", "batt")
	if err != nil {
		return "", fmt.Errorf("failed to get battery info: %w", err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return core.MustFormatJSON(map[string]any{"available": false}), nil
	}

	result := map[string]any{"raw": raw, "available": true}

	if strings.Contains(raw, "AC Power") {
		result["source"] = "ac"
	} else if strings.Contains(raw, "Battery Power") {
		result["source"] = "battery"
	}

	for _, line := range strings.Split(raw, "\n") {
		if idx := strings.Index(line, "%"); idx > 0 {
			start := idx - 1
			for start >= 0 && line[start] >= '0' && line[start] <= '9' {
				start--
			}
			if start < idx-1 {
				result["percentage"] = line[start+1:idx] + "%"
			}
			result["charging"] = strings.Contains(line, "charging") && !strings.Contains(line, "discharging")
			break
		}
	}

	return core.MustFormatJSON(result), nil
}

func handleDisk(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal(input, &params)
	if params.Path == "" {
		params.Path = "/"
	}
	if core.ContainsTraversal(params.Path) {
		return "", fmt.Errorf("%w: path must not contain '..'", core.ErrInvalidInput)
	}

	out, err := core.RunCommand(ctx, "df", "-h", params.Path)
	if err != nil {
		return "", fmt.Errorf("failed to get disk info: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return core.MustFormatJSON(map[string]any{"path": params.Path, "raw": string(out)}), nil
	}

	fields := strings.Fields(lines[1])
	if len(fields) >= 5 {
		return core.MustFormatJSON(map[string]any{
			"path":       params.Path,
			"filesystem": fields[0],
			"size":       fields[1],
			"used":       fields[2],
			"available":  fields[3],
			"capacity":   fields[4],
		}), nil
	}

	return core.MustFormatJSON(map[string]any{"path": params.Path, "raw": lines[1]}), nil
}

func handleNetwork(ctx context.Context, input json.RawMessage) (string, error) {
	result := map[string]any{}

	wifiOut, err := core.RunCommand(ctx, "networksetup", "-getairportnetwork", "en0")
	if err == nil {
		line := strings.TrimSpace(string(wifiOut))
		if parts := strings.SplitN(line, ": ", 2); len(parts) == 2 {
			result["wifi"] = parts[1]
		} else {
			result["wifi"] = line
		}
	}

	ifOut, err := core.RunCommand(ctx, "ifconfig", "-l")
	if err == nil {
		var interfaces []map[string]string
		for _, iface := range strings.Fields(strings.TrimSpace(string(ifOut))) {
			if iface == "lo0" || strings.HasPrefix(iface, "utun") || strings.HasPrefix(iface, "awdl") ||
				strings.HasPrefix(iface, "llw") || strings.HasPrefix(iface, "bridge") ||
				strings.HasPrefix(iface, "ap") || strings.HasPrefix(iface, "gif") ||
				strings.HasPrefix(iface, "stf") || strings.HasPrefix(iface, "anpi") {
				continue
			}
			ipOut, err := core.RunCommand(ctx, "ipconfig", "getifaddr", iface)
			if err == nil {
				if ip := strings.TrimSpace(string(ipOut)); ip != "" {
					interfaces = append(interfaces, map[string]string{"name": iface, "ip": ip})
				}
			}
		}
		if len(interfaces) > 0 {
			result["interfaces"] = interfaces
		}
	}

	dnsOut, err := core.RunCommand(ctx, "networksetup", "-getdnsservers", "Wi-Fi")
	if err == nil {
		dns := strings.TrimSpace(string(dnsOut))
		if !strings.Contains(dns, "aren't any") {
			result["dns"] = strings.Split(dns, "\n")
		}
	}

	return core.MustFormatJSON(result), nil
}
