//go:build darwin

package core

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestRunJXA_SimpleOutput(t *testing.T) {
	result, err := RunJXA(context.Background(), []byte(`JSON.stringify({ok:true})`), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out struct{ OK bool `json:"ok"` }
	if err := json.Unmarshal(result, &out); err != nil || !out.OK {
		t.Errorf("unexpected output: %s", result)
	}
}

func TestRunJXA_ParamsPassthrough(t *testing.T) {
	script := []byte(`
		ObjC.import("Foundation");
		var env = $.NSProcessInfo.processInfo.environment.objectForKey("APPLE_TOOLS_PARAMS").js;
		var p = JSON.parse(env);
		JSON.stringify({got: p.val});
	`)
	result, err := RunJXA(context.Background(), script, map[string]string{"val": "test123"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var out struct{ Got string `json:"got"` }
	json.Unmarshal(result, &out)
	if out.Got != "test123" {
		t.Errorf("expected test123, got %s", out.Got)
	}
}

func TestRunJXA_Timeout(t *testing.T) {
	script := []byte(`ObjC.import("Foundation");$.NSThread.sleepForTimeInterval(10);JSON.stringify({});`)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := RunJXA(ctx, script, nil)
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout, got: %v", err)
	}
}

func TestRunCommand_Basic(t *testing.T) {
	out, err := RunCommand(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if string(out) != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", out)
	}
}

func TestRunCommand_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := RunCommand(ctx, "sleep", "10")
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout, got: %v", err)
	}
}
