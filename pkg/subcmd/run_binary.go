package subcmd

import (
    "bytes"
    "context"
    "os/exec"
    "time"
)

// RunBinary executes an external binary with the given arguments.
// It returns the combined stdout+stderr output and any error.
// The command is canceled if it exceeds the timeout duration.
func RunBinary(ctx context.Context, binary string, args []string, timeout time.Duration) (string, error) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    cmd := exec.CommandContext(ctx, binary, args...)
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    err := cmd.Run()
    return out.String(), err
}