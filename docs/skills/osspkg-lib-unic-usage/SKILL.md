---
name: osspkg-lib-unic-usage
description: Use the go.osspkg.com/unic library correctly in Go code for UNIC configuration parsing and serialization.
---

# UNIC usage skill

Use this skill when writing, reviewing, or explaining Go code that reads or
writes UNIC configuration with `go.osspkg.com/unic`.

## Core rules

- Decode into a non-nil pointer to a struct:
  `unic.Unmarshal(data, &cfg)`.
- Encode a struct or a non-nil pointer to a struct:
  `unic.Marshal(cfg)` or `unic.Marshal(&cfg)`.
- Maps, slices, scalar values, and `any` are supported as struct fields. A map
  or scalar cannot be passed as the top-level argument to `Marshal`.
- Exported fields need a `unic` tag to participate in encoding/decoding. Use
  `unic:"-"` or omit the tag to exclude a field.
- Use `default=value` for missing fields, `omitempty` to skip zero values,
  `attr=N` for positional block attributes, and `desc=value` for comments.
- Check and return errors from both `Unmarshal` and `Marshal`; malformed input,
  type mismatches, invalid tags, and invalid UTF-8 are reported as errors.
- Keep configuration structs explicit and typed. Use `map[string]any` only when
  the configuration shape is intentionally dynamic.

## Workflow

1. Define an exported configuration struct with stable `unic` names.
2. Read the bytes from the chosen source and call `Unmarshal` with `&cfg`.
3. Validate application-specific invariants after decoding; tags only perform
   format-level conversion and defaults.
4. For output, call `Marshal`, handle the error, then write the returned bytes.
5. Add a round-trip test for non-trivial nested configurations and tests for
   defaults, optional fields, maps, and attributes when they are used.

Read [references/code-examples.md](references/code-examples.md) for concrete
correct and incorrect patterns before implementing or reviewing integration
code.

