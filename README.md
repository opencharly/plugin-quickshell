# plugin-quickshell

The `quickshell:` check verb — IPC against any Quickshell desktop shell.

Quickshell reports IPC failures on **stdout with exit 0**, so a `command:` step asserting
"the menu opened" passes against a shell that never opened one. This verb reads those
responses and fails on them.

See the `quickshell` skill in `candy/plugin-quickshell/charly.yml` for authoring.
