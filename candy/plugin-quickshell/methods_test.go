package quickshell

import (
	"context"
	"strings"
	"testing"

	"github.com/opencharly/sdk"

	"github.com/opencharly/plugin-quickshell/candy/plugin-quickshell/params"
)

// THE property this verb exists for.
//
// Quickshell's `qs ipc` reports IPC-level failures on STDOUT and exits ZERO. A
// `command:` step asserting "the menu opened" therefore PASSES against a shell with no
// menu plugin loaded, against a typo in the target name, and against an IPC surface
// that changed underneath it — green, asserting nothing.
//
// These four strings live in the quickshell binary itself, which is also what makes
// this verb desktop-agnostic: they are Quickshell's vocabulary, not omarchy's.
func TestClassifyIPCOutput_ExitZeroFailuresAreCaught(t *testing.T) {
	for _, out := range []string{
		"Target not found.",
		"Function not found.",
		"Too few arguments provided (expected 2)",
		"Too many arguments provided",
	} {
		msg, failed := classifyIPCOutput(out)
		if !failed {
			t.Errorf("qs said %q and the verb reported SUCCESS — a command: step would too", out)
			continue
		}
		if !strings.Contains(msg, "exited 0") {
			t.Errorf("failure message does not explain WHY this needs a verb rather than command:\n  %s", msg)
		}
	}
}

// A real response must not be mistaken for a failure. The guard is prefix-based, so a
// payload that merely mentions one of the phrases stays a success.
func TestClassifyIPCOutput_RealResponsesPass(t *testing.T) {
	for _, out := range []string{
		"",
		"pong",
		`{"plugins":["omarchy.bar","omarchy.menu"]}`,
		"the user asked: why is Target not found. a failure?",
	} {
		if _, failed := classifyIPCOutput(out); failed {
			t.Errorf("a legitimate response was rejected as an IPC failure: %q", out)
		}
	}
}

// `ping` is a convenience over `call`, and its defaults must be overridable — a desktop
// is free to name its root target and health function otherwise.
func TestResolveCall(t *testing.T) {
	for _, tc := range []struct{ name, method, inTarget, inFn, wantTarget, wantFn string }{
		{"ping defaults", "ping", "", "", "shell", "ping"},
		{"ping target override", "ping", "bar", "", "bar", "ping"},
		{"ping function override", "ping", "", "health", "shell", "health"},
		{"call explicit", "call", "omarchy.menu", "summon", "omarchy.menu", "summon"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, fn, err := resolveCall(&params.QuickshellInput{
				Method: tc.method, Target: tc.inTarget, Function: tc.inFn,
			})
			if err != nil {
				t.Fatalf("resolveCall: %v", err)
			}
			if target != tc.wantTarget || fn != tc.wantFn {
				t.Errorf("got %s.%s, want %s.%s", target, fn, tc.wantTarget, tc.wantFn)
			}
		})
	}
}

// `call` without a target or function is an authoring error, and must be reported as
// one rather than sent to qs as an empty argument — which qs answers with "Target not
// found." on stdout, exit 0, i.e. the silent pass this verb is built to prevent.
func TestResolveCall_RejectsIncompleteCall(t *testing.T) {
	if _, _, err := resolveCall(&params.QuickshellInput{Method: "call", Function: "summon"}); err == nil {
		t.Error("call with no target was accepted")
	}
	if _, _, err := resolveCall(&params.QuickshellInput{Method: "call", Target: "shell"}); err == nil {
		t.Error("call with no function was accepted")
	}
}

// The step runs over SSH, where there is no ambient WAYLAND_DISPLAY — and `qs` matches
// instances BY DISPLAY. Without recovery from XDG_RUNTIME_DIR the verb fails on exactly
// the guests it is most useful for.
func TestIPCCommand_RecoversTheWaylandDisplay(t *testing.T) {
	got := ipcCommand(&params.QuickshellInput{Config: "/usr/share/omarchy/shell"}, "shell", "ping")
	if !strings.Contains(got, "XDG_RUNTIME_DIR") || !strings.Contains(got, "wayland-") {
		t.Errorf("no display recovery; the verb would fail over SSH:\n  %s", got)
	}
	if !strings.Contains(got, "qs ipc -n -p '/usr/share/omarchy/shell' call -- 'shell' 'ping'") {
		t.Errorf("unexpected ipc invocation:\n  %s", got)
	}
}

// An explicit display wins, and no discovery is emitted for it.
func TestIPCCommand_ExplicitDisplay(t *testing.T) {
	got := ipcCommand(&params.QuickshellInput{Config: "/c", WaylandDisplay: "wayland-1"}, "shell", "ping")
	if !strings.Contains(got, "export WAYLAND_DISPLAY='wayland-1'") {
		t.Errorf("explicit display not exported:\n  %s", got)
	}
	if strings.Contains(got, "XDG_RUNTIME_DIR") {
		t.Errorf("discovery emitted despite an explicit display:\n  %s", got)
	}
}

// Arguments are quoted, so a JSON payload survives the venue shell intact.
func TestIPCCommand_QuotesJSONPayload(t *testing.T) {
	got := ipcCommand(&params.QuickshellInput{
		Config: "/c", Args: []string{`{"menu":"root"}`},
	}, "omarchy.menu", "summon")
	if !strings.Contains(got, `'{"menu":"root"}'`) {
		t.Errorf("JSON payload not quoted for the shell:\n  %s", got)
	}
}

// The scalar shorthand (`quickshell: ping`) only works if the capability declares which
// field the scalar lands in. Without it every shorthand step fails to PARSE — the whole
// project load, not just that step.
func TestProvidedCapability_DeclaresMethodPrimary(t *testing.T) {
	caps := providedCapabilities()
	if len(caps) != 1 || caps[0].Word != "quickshell" {
		t.Fatalf("unexpected capabilities: %+v", caps)
	}
	if caps[0].Primary != "method" {
		t.Errorf("Primary = %q, want \"method\" — without it `quickshell: ping` cannot parse", caps[0].Primary)
	}
}

// The classifier being CORRECT is not the same as it being CALLED. This drives the real
// ipcCall path with a canned qs response and asserts the exit-zero failure is caught —
// it fails if the guard is ever removed from the call site, which a test of
// classifyIPCOutput alone would not.
func TestIPCCall_ExitZeroFailureIsRejectedOnTheRealPath(t *testing.T) {
	orig := venueCapture
	t.Cleanup(func() { venueCapture = orig })
	venueCapture = func(_ context.Context, _ *sdk.Executor, _ string) (string, error) {
		return "Target not found.\n", nil // qs's real shape: failure on stdout, exit 0
	}

	_, err := ipcCall(context.Background(), nil, &params.QuickshellInput{Config: "/c"}, "omarchy.menu", "summon")
	if err == nil {
		t.Fatal("ipcCall returned SUCCESS for \"Target not found.\" — the guard is not wired into the call path")
	}
	if !strings.Contains(err.Error(), "Target not found.") {
		t.Errorf("error does not carry qs's own message:\n  %v", err)
	}
}

// And a real response still succeeds through the same path.
func TestIPCCall_RealResponsePasses(t *testing.T) {
	orig := venueCapture
	t.Cleanup(func() { venueCapture = orig })
	venueCapture = func(_ context.Context, _ *sdk.Executor, _ string) (string, error) {
		return "pong", nil
	}
	out, err := ipcCall(context.Background(), nil, &params.QuickshellInput{Config: "/c"}, "shell", "ping")
	if err != nil {
		t.Fatalf("a valid response was rejected: %v", err)
	}
	if out != "pong" {
		t.Errorf("output = %q, want pong", out)
	}
}
