// Command serve is the OUT-OF-PROCESS entrypoint for the quickshell verb plugin: a
// thin shim serving the importable provider over go-plugin gRPC via sdk.Serve.
package main

import (
	quickshell "github.com/opencharly/plugin-quickshell/candy/plugin-quickshell"
	"github.com/opencharly/sdk"
)

func main() { sdk.Serve(quickshell.NewProvider(), quickshell.NewMeta()) }
