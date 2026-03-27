//go:build darwin

package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	DefaultTimeout = 30 * time.Second
	ParamsEnvKey   = "APPLE_TOOLS_PARAMS"
)

// RunJXA executes a JXA script via osascript with params passed as an env var.
func RunJXA(ctx context.Context, script []byte, params any) (json.RawMessage, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to encode parameters: %v", ErrInvalidInput, err)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "osascript", "-l", "JavaScript", "-e", string(script))
	cmd.Env = append(os.Environ(), ParamsEnvKey+"="+string(paramsJSON))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if ctx.Err() != nil {
			if cmd.Process != nil {
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			}
			return nil, fmt.Errorf("%w: osascript killed after timeout", ErrTimeout)
		}

		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}

		classified := ClassifyError(exitCode, stderr.String())
		if classified != nil {
			return nil, classified
		}
		return nil, fmt.Errorf("osascript failed: %w", err)
	}

	output := bytes.TrimSpace(stdout.Bytes())
	if len(output) == 0 {
		return json.RawMessage("null"), nil
	}

	return json.RawMessage(output), nil
}

// RunCommand executes a CLI command and returns stdout.
func RunCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() != nil {
			if cmd.Process != nil {
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			}
			return nil, fmt.Errorf("%w: %s killed after timeout", ErrTimeout, name)
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s failed: %s", name, msg)
	}

	return stdout.Bytes(), nil
}
