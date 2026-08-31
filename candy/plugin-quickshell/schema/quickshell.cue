// This plugin's OWN CUE schema — the typed plugin_input for the `quickshell`
// live-deployment check verb, served over the Describe channel. It is the SINGLE
// SOURCE for this plugin's params, used two ways (the contract core `spec` and the
// reference plugin-http use):
//
//  1. GENERATE the Go param struct — `cue exp gengotypes` emits ../params/cue_types_gen.go,
//     so the provider decodes plugin_input into a TYPED struct, never a hand-parsed map.
//  2. VALIDATE authored input AT RUNTIME — the host splices this source onto the base
//     and validates every authored `quickshell` step's plugin_input against #QuickshellInput.
//
// SELF-CONTAINED: it references NO base def, so it compiles standalone AND splices onto
// the base. The shared step matchers (exit_status/stdout/stderr) stay on core #Op.
#QuickshellInput: {
	// method — what to do with the shell's IPC.
	//
	//   call  the generic primitive: invoke `function` on `target` with `args`
	//   ping  the liveness convenience — `call` against the shell's own ping
	//
	// The set is deliberately small. Quickshell's IPC surface is `call` plus the
	// listen/wait/prop streaming forms; a check verb wants request/response, and
	// everything a desktop exposes is reachable through `call`.
	method: "call" | "ping"
	// config — the Quickshell config DIRECTORY (`qs ipc -p <dir>`), i.e. the folder
	// holding shell.qml. REQUIRED: `qs` matches instances by config path, and a caller
	// from outside the session has no ambient one. Omarchy's is /usr/share/omarchy/shell.
	config: string & !=""
	// target — the IpcHandler target name. Defaults to "shell" for `ping`.
	target?: string
	// function — the IpcHandler function to invoke. Required for `call`.
	function?: string
	// args — positional arguments passed after the function name. A JSON payload is
	// just a string argument, e.g. '{"menu":"root"}'.
	args?: [...string]
	// wayland_display — the WAYLAND_DISPLAY to talk to. Empty means "discover it from
	// XDG_RUNTIME_DIR", which is what a step running over SSH needs: qs matches
	// instances by display and an SSH session has none.
	wayland_display?: string @go(WaylandDisplay)
	// runtime_dir — the XDG_RUNTIME_DIR the SHELL registered under, when that differs
	// from the one a check step inherits. Empty means "use the step's own".
	//
	// qs finds instances through $XDG_RUNTIME_DIR, so the two have to agree. They do
	// not in every venue: a nested-compositor pod runs the compositor and its shell
	// under a dedicated runtime dir (omarchy-cstream uses /tmp/cstream-rt) while an
	// exec into that container inherits a different one. Measured there: with the
	// step's own dir, `qs list --all` reports "No running instances" and every call
	// fails; with the shell's, the same call answers "ok".
	//
	// It pairs with wayland_display, and both are usually needed together — qs filters
	// instances by DISPLAY as well as by runtime dir, so getting one right and the
	// other wrong still fails, with a different message each way.
	runtime_dir?: string @go(RuntimeDir)
}
