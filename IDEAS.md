# Ideas for Improvement

The following is a list of ideas/features/updates for this project.

## WASM Builds

It'd be quite handy to be able to build an app for WASM.

## Expand Allow to Open Ports

It'd be nice if the `allow` command (or similar) supported opening firewall ports. This would need to be cross-platform, i.e. supporting Windows Defender, iptables, and possibly ufw.

## TODO Report

`qgo todo` should produce a list of known TODO statements, serving as a task list and a reference directly to the line of code where the TODO is found. Many apps like VS Code do this, but this isn't a universally implemented feature (and VS Code isn't the only IDE used). It may be useful to create/sync the TODOs as non-failing tests in the test suite (i.e. stubs), or to output the values as JSON. Any type of report could be diffed in git, allowing tools to automatically recognize tasks that have been completed.

## Notes Report

`qgo notes` should parse all comments for `@note`/`@notes`, compiling a list of notes for developers to be aware of. There should be an option to update the `README.md` or a `notes.md` file. This is ultimately a way to highlight specific comments in an effort to draw attention to them.
