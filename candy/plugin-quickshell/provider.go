package quickshell

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/opencharly/plugin-quickshell/candy/plugin-quickshell/params"
	"github.com/opencharly/sdk"
	"github.com/opencharly/sdk/kit"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// quickshellEnv is the plugin-side decode of the CheckEnv the host ships as
// Operation.Env. Mode gates the box-context skip; the venue work travels over the
// executor reverse channel, not this snapshot.
type quickshellEnv struct {
	Mode string `json:"mode"` // "live" | "box"
}

type provider struct{ pb.UnimplementedProviderServer }

// Invoke runs one `quickshell:` operation: decode the Op and the typed plugin input,
// skip under `charly check box` (no running shell in a disposable container), dial
// back the host's live executor, dispatch, then self-evaluate the shared matchers.
func (p provider) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	var op spec.Op
	if len(req.GetParamsJson()) > 0 {
		if err := json.Unmarshal(req.GetParamsJson(), &op); err != nil {
			return sdk.ResultJSON("fail", "quickshell: decode op: "+err.Error())
		}
	}
	var in params.QuickshellInput
	kit.DecodeInput(op.PluginInput, &in)
	var env quickshellEnv
	if len(req.GetEnvJson()) > 0 {
		_ = json.Unmarshal(req.GetEnvJson(), &env)
	}

	if env.Mode == "box" {
		return sdk.ResultJSON("skip", fmt.Sprintf(
			"quickshell: %s requires a running desktop shell (skip under charly check box)", in.Method))
	}

	// EXEC-based: a missing broker is a HARD FAIL, never a silent skip — the verb
	// cannot do its job without the venue.
	exec, err := sdk.ExecutorFromInvoke(req.GetExecutorBrokerId())
	if err != nil {
		return sdk.ResultJSON("fail", fmt.Sprintf(
			"quickshell: %s has no host executor attached — it needs the live venue (%v)", in.Method, err))
	}

	out, runErr := dispatch(ctx, exec, &in)
	return sdk.VerbVerdict("quickshell", in.Method, out, runErr, &op, false)
}
