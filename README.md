# plugin-quickshell

Quickshell desktop-shell automation for OpenCharly — the `quickshell:` check verb.

The verb drives the IPC channel of a [Quickshell](https://quickshell.outfoxxed.me)
desktop shell — the surface a desktop's panels, menus, notifications and overlays
are summoned and dismissed through. It is served **out-of-process** — charly's
loader fetches this repo, host-builds the provider binary, and serves it over
go-plugin gRPC.

**Why it is a verb and not a `command:` step.** Quickshell's `qs ipc` reports
IPC-level failures — `Target not found.`, `Function not found.`, argument-count
errors — on **stdout** while exiting **zero**. A `command:` step asserting "the
menu opened" therefore passes against a shell with no menu plugin loaded, against
a typo in the target name, and against an IPC surface that changed underneath it:
green, asserting nothing. This verb reads those responses and fails on them.

It is **desktop-agnostic by construction**: it speaks Quickshell's IPC, not any
one desktop's vocabulary — `config:` names the shell directory and
`target:`/`function:` name whatever that shell's QML exposes. Omarchy is one
caller; the failure strings the verb keys on live in the `quickshell` binary
itself.

## What it provides

| Capability | Surface |
|---|---|
| `verb:quickshell` | the `quickshell:` check verb — `ping` (liveness) and `call` (invoke a target's function) |

## How to use it

Compose the plugin candy in a Quickshell-desktop bed:

```yaml
- '@github.com/opencharly/plugin-quickshell/candy/plugin-quickshell:<tag>'
```

Then author the verb in a plan:

```yaml
- check: the shell answers a ping
  context: [runtime]
  eventually: 60s
  quickshell:
    method: ping
    config: /usr/share/omarchy/shell

- check: the core shell plugins are loaded
  context: [runtime]
  quickshell:
    method: call
    config: /usr/share/omarchy/shell
    target: shell
    function: listPlugins
  stdout:
    - contains: omarchy.bar
    - contains: omarchy.menu
```

`config:` is required because `qs` matches instances **by config path**, and a
step arriving from outside the session has no ambient one. On a summon step,
assert the ANSWER — a shell can accept an IPC call and render nothing.

## Layout

- `candy/plugin-quickshell/` — the plugin module: `methods.go` (the `ping`/`call`
  surface), `provider.go` / `plugin.go`, `schema/quickshell.cue` (the
  self-contained input schema), `params/cue_types_gen.go`, `cmd/serve/main.go`.
- `charly.yml` — the root project manifest (`discover: candy` + the
  `quickshell-skill` skill entity).
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.

## Related

- Owning skill: `/charly-check:quickshell` — the `quickshell:` check verb,
  authored in this candy's `skill:` entity.
- `/charly-check:wl` — the compositor-observation verb that pairs with it.
- `/charly-internals:plugin` — the out-of-process plugin model.
