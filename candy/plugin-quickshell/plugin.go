// Package quickshell is the charly plugin serving the `quickshell` live-deployment
// check verb (an importable root package + its own go.mod). It drives the IPC channel
// of a Quickshell desktop shell — the control surface a desktop's panels, menus,
// notifications and overlays are summoned and dismissed through — by invoking
// `qs ipc call` inside the running deployment. The host go-builds this binary and
// serves it OUT-OF-PROCESS over go-plugin gRPC via the charly plugin SDK, so the
// `quickshell:` verb dispatches through the provider registry exactly like a built-in.
//
// EXEC-based external verb, like wl: the host attaches its live DeployExecutor over
// the E3b reverse channel and this plugin dials back through the SDK
// (sdk.ExecutorFromInvoke) to RunCapture the venue's `qs` binary. It owns NO podman /
// SSH machinery.
//
// DESKTOP-AGNOSTIC BY CONSTRUCTION. The verb speaks Quickshell's IPC, not any one
// desktop's vocabulary: `config` names the shell directory and `target`/`function`
// name whatever that shell's QML exposes. Omarchy is simply one caller.
package quickshell

import (
	"embed"

	"github.com/opencharly/sdk"
	pb "github.com/opencharly/spec/proto"
)

//go:embed schema/*.cue
var schemaFS embed.FS

// NewProvider returns the quickshell provider.
func NewProvider() pb.ProviderServer { return &provider{} }

// NewMeta advertises verb:quickshell (plugin_input #QuickshellInput) + the plugin's
// self-contained CUE schema. Primary names the field a SCALAR shorthand lands in, so
// `quickshell: ping` desugars to `quickshell: {method: ping}`.
func NewMeta() pb.PluginMetaServer {
	return sdk.NewMeta("2026.243.1500", providedCapabilities(), schemaFS)
}

// providedCapabilities is the SINGLE source for what this plugin advertises, so the
// manifest and anything checking it cannot drift apart.
func providedCapabilities() []sdk.ProvidedCapability {
	return []sdk.ProvidedCapability{{
		Class:    "verb",
		Word:     "quickshell",
		InputDef: "#QuickshellInput",
		Primary:  "method",
	}}
}
