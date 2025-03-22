package common

import (
	"context"
	"errors"
	"fmt"
	"grove/config"
	"log"
	"os/exec"
	"strings"
	"time"
)

type ProcessExecutionError struct {
	Cmd         string
	ExitCode    int
	Stdout      string
	Stderr      string
	Description string
}

type FlagPair struct {
	Key   string
	Value interface{}
}

func (e *ProcessExecutionError) Error() string {
	return fmt.Sprintf("Command '%s' failed. %s Exit code: %d\nstderr: %s\nstdout: %s",
		e.Cmd, e.Description, e.ExitCode, e.Stderr, e.Stdout)
}

func Execute(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("no command specified")
	}

	// Tạo command với context có timeout
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	// Thực thi command, lấy output và error
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Nếu lỗi do timeout, return lỗi cụ thể
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("command '%s' timed out", strings.Join(args, " "))
		}

		// Nếu lỗi khác thì tạo ProcessExecutionError
		exitCode := 1
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		}
		return "", &ProcessExecutionError{
			Cmd:         strings.Join(args, " "),
			ExitCode:    exitCode,
			Stderr:      string(output),
			Stdout:      "",
			Description: "Execute command failed",
		}
	}

	return string(output), nil
}

func ExecuteWithTimeout(args []string, kwargs map[string]interface{}) (string, error) {
	// Lấy giá trị timeout từ kwargs, ko có thì dùng mặc định
	timeout, ok := kwargs["timeout"].(int)
	if !ok {
		timeout = config.DefaultTimeout
	}

	logOutputOnError, _ := kwargs["log_output_on_error"].(bool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	rootPermission, ok := kwargs["run_as_root"].(bool)
	if ok && rootPermission {
		args = append([]string{"sudo"}, args...)
	}

	// Chạy command với timeout
	result, err := Execute(ctx, args)
	if err != nil {
		var exitError *ProcessExecutionError
		if errors.As(err, &exitError) {
			// Log chi tiết lỗi
			if logOutputOnError {
				log.Printf("Command '%s' failed. %s Exit code: %d\nstderr: %s\nstdout: %s", exitError.Cmd,
					exitError.Description, exitError.ExitCode, exitError.Stderr, exitError.Stdout)
			}
			return "", err
		}

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			msg := fmt.Sprintf("Timeout after %v seconds running: %v %v.",
				timeout, args, kwargs)
			log.Println(msg)
			return "", fmt.Errorf("command '%s' timed out", strings.Join(args, " "))
		}
	}

	return result, nil
}

func buildCommandOptions(options []FlagPair) []string {
	var result []string
	for _, item := range options {
		if valid, ok := item.Value.(bool); ok && valid {
			if value, oke := item.Value.(string); oke {
				result = append(result, "-"+value)
			}
		}
	}
	return result
}

func ExecuteShellCmd(cmd string, options []string, args []string, kwargs map[string]interface{}) (string, error) {
	var execArgs []FlagPair

	runAsRoot, ok := kwargs["run_as_root"].(bool)
	if !ok {
		runAsRoot = false
	}
	if runAsRoot {
		execArgs = append(execArgs, FlagPair{"run_as_root", true})
		execArgs = append(execArgs, FlagPair{"root_helper", "sudo"})
	}

	timeout, ok := kwargs["timeout"].(int)
	if !ok {
		timeout = config.DefaultTimeout
	}
	execArgs = append(execArgs, FlagPair{"timeout", timeout})
	cmdFlags := buildCommandOptions(execArgs)
	cmdArgs := append([]string{cmd}, cmdFlags...)
	cmdArgs = append([]string{cmd}, args...)
	stdout, _ := ExecuteWithTimeout(cmdArgs, kwargs)
	return stdout, nil
}
