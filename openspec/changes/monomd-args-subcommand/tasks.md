## 1. Go Logic Package

- [ ] 1.1 Create `internal/args/` package with core parsing function
- [ ] 1.2 Implement `--` separator handling — split input into mnmd arguments (before `--`) and raw args (after `--`); error if `--` is missing
- [ ] 1.3 Implement modifier parsing — consume `--`-prefixed tokens before the flag name as modifiers (`--boolean`, `--short`); support both `--mod=val` and `--mod val` forms; error on unknown modifiers
- [ ] 1.4 Implement long-form equals parsing (`--flag=value`)
- [ ] 1.5 Implement long-form space parsing (`--flag value`), skipping next token if it starts with `--` or `-`
- [ ] 1.6 Implement `--short` modifier — register a single-character alias (must be exactly one char, error otherwise), search for `-c=value`, `-c value`, and bundled forms (`-xc value`, `-xc=value`); value flag in non-last bundle position = absent
- [ ] 1.7 Implement last-wins for duplicate flags (across both long and short forms)
- [ ] 1.8 Implement `--boolean` behavior — exit-code-only presence check with `--no-<flag>` negation; last-wins between flag and its negation
- [ ] 1.9 Add `internal/args/args_test.go` covering all spec scenarios

## 2. Binary Dispatch

- [ ] 2.1 Wire `mnmd args [modifiers...] <flag> -- <raw args...>` dispatch in `cmd/mnmd/main.go`
- [ ] 2.2 Print resolved value to stdout and exit 0 on success (value mode)
- [ ] 2.3 Exit 1 silently when value flag is absent
- [ ] 2.4 Exit 0 (present) or 1 (absent/negated) with no stdout when `--boolean` is used
- [ ] 2.5 Exit 1 with stderr error when `--` separator is missing

## 3. e2e Tests

- [ ] 3.1 Create `tests/mnmd_args_test` shUnit2 e2e test file following the project pattern
- [ ] 3.2 Test long-form equals (`--prop=value`)
- [ ] 3.3 Test long-form space (`--prop value`)
- [ ] 3.4 Test short-form equals (`-p=value`) with `--short` modifier
- [ ] 3.5 Test short-form space (`-p value`) with `--short` modifier
- [ ] 3.6 Test last-wins across long and short forms
- [ ] 3.7 Test multi-char `--short` value produces error
- [ ] 3.8 Test bundled short flags — boolean flag recognized in bundle
- [ ] 3.9 Test bundled short flags — value flag last in bundle takes value
- [ ] 3.10 Test bundled short flags — value flag not last in bundle = absent
- [ ] 3.11 Test absent flag exits 1 silently
- [ ] 3.12 Test `--boolean` present flag exits 0, no stdout
- [ ] 3.13 Test `--boolean` absent flag exits 1, no stdout
- [ ] 3.14 Test `--boolean` with `--no-<flag>` negation
- [ ] 3.15 Test `--boolean` last-wins between flag and negation
- [ ] 3.16 Test duplicate flags last-wins
- [ ] 3.17 Test space-form followed by another flag (no value consumed)
- [ ] 3.18 Test other flags alongside target flag are ignored
- [ ] 3.19 Test unknown modifier produces error
- [ ] 3.20 Test modifier equals and space forms are equivalent
- [ ] 3.21 Test missing `--` separator produces error

## 4. Validation

- [ ] 4.1 Run `go test ./...` — all tests pass
- [ ] 4.2 Run `go vet ./...` — no errors
- [ ] 4.3 Run `bash tests/mnmd_args_test` — all shUnit2 tests pass
