package quickshell

import (
	"context"
	"fmt"
	"strings"

	"github.com/opencharly/plugin-quickshell/candy/plugin-quickshell/params"
	"github.com/opencharly/sdk"
)

// ipcFailures are the messages Quickshell's `qs ipc` writes to STDOUT while exiting
// ZERO. They are the whole reason this verb exists rather than a `command:` step.
//
// Read the shape carefully: a caller that only checks the exit status treats
// "Target not found." as a SUCCESSFUL call. A bed asserting "the menu opened" would
// pass against a shell with no menu plugin loaded, against a typo in the target name,
// and against an IPC surface that changed under it. The step would be green and
// assert nothing — the silent-pass shape this org has been bitten by repeatedly.
//
// The strings are Quickshell's own, not any desktop's: they appear in the quickshell
// binary itself, which is what makes this check desktop-agnostic.
var ipcFailures = []string{
	"Target not found.",
	"Function not found.",
	"Too few arguments provided",
	"Too many arguments provided",
}

// dispatch runs one quickshell method against the venue.
func dispatch(ctx context.Context, ex *sdk.Executor, in *params.QuickshellInput) (string, error) {
	target, function, err := resolveCall(in)
	if err != nil {
		return "", err
	}
	return ipcCall(ctx, ex, in, target, function)
}

// resolveCall turns the input into the (target, function) pair to invoke, applying the
// `ping` convenience. Separated from the execution so it is testable without a venue.
func resolveCall(in *params.QuickshellInput) (target, function string, err error) {
	switch in.Method {
	case "ping":
		// The liveness convenience. "shell" is Quickshell's conventional root target
		// and `ping` its conventional health function; both stay overridable, because
		// a desktop is free to name them otherwise.
		target, function = in.Target, in.Function
		if target == "" {
			target = "shell"
		}
		if function == "" {
			function = "ping"
		}
		return target, function, nil
	case "call":
		if in.Target == "" {
			return "", "", fmt.Errorf("quickshell: call needs a target (the IpcHandler name)")
		}
		if in.Function == "" {
			return "", "", fmt.Errorf("quickshell: call needs a function (the IpcHandler function to invoke)")
		}
		return in.Target, in.Function, nil
	default:
		return "", "", fmt.Errorf("quickshell: unknown method %q (want call or ping)", in.Method)
	}
}

// ipcCommand builds the shell line that performs one IPC call.
//
// WAYLAND_DISPLAY is recovered from XDG_RUNTIME_DIR when the caller has none, because
// `qs` matches instances BY DISPLAY and a step arriving over SSH has no ambient
// display. Without this the verb fails on every guest it is most useful for.
func ipcCommand(in *params.QuickshellInput, target, function string) string {
	var b strings.Builder
	if in.WaylandDisplay != "" {
		fmt.Fprintf(&b, "export WAYLAND_DISPLAY=%s; ", shellQuote(in.WaylandDisplay))
	} else {
		b.WriteString(`if [ -z "${WAYLAND_DISPLAY:-}" ]; then ` +
			`s=$(ls -t "${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"/wayland-[0-9]* 2>/dev/null | grep -v '\.lock$' | head -n1); ` +
			`[ -n "$s" ] && export WAYLAND_DISPLAY="${s##*/}"; fi; `)
	}
	fmt.Fprintf(&b, "qs ipc -n -p %s call -- %s %s", shellQuote(in.Config), shellQuote(target), shellQuote(function))
	for _, a := range in.Args {
		fmt.Fprintf(&b, " %s", shellQuote(a))
	}
	return b.String()
}

// venueCapture is the seam the venue call goes through. A var, not a direct method
// call, so a test can substitute a canned qs response and prove the exit-zero guard is
// actually WIRED IN — not merely present as a function. Testing the classifier alone
// would pass with the call site deleted, which is the same "coverage that cannot fail"
// trap this plugin exists to close.
var venueCapture = func(ctx context.Context, ex *sdk.Executor, cmd string) (string, error) {
	return ex.VenueCapture(ctx, cmd)
}

func ipcCall(ctx context.Context, ex *sdk.Executor, in *params.QuickshellInput, target, function string) (string, error) {
	out, err := venueCapture(ctx, ex, ipcCommand(in, target, function))
	if err != nil {
		return "", fmt.Errorf("quickshell: %s.%s: the shell did not answer (is it running, and is config %q right?): %w",
			target, function, in.Config, err)
	}
	if msg, bad := classifyIPCOutput(out); bad {
		return "", fmt.Errorf("quickshell: %s.%s: %s", target, function, msg)
	}
	return out, nil
}

// classifyIPCOutput turns an exit-ZERO qs response into a verdict. It is the guard the
// whole verb exists for: these messages arrive on stdout with a success exit status.
func classifyIPCOutput(out string) (msg string, failed bool) {
	trimmed := strings.TrimSpace(out)
	for _, f := range ipcFailures {
		if strings.HasPrefix(trimmed, f) {
			return trimmed + " (qs reported this on STDOUT and still exited 0 — a command: step " +
				"would have passed)", true
		}
	}
	return "", false
}

// shellQuote single-quotes an argument for the venue shell.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
