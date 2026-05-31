## 1. monom_error Function

- [ ] 1.1 Implement `monom_error` in `src/monom` — print message to stderr, exit with code (default 1)
- [ ] 1.2 Verify it works in both bash and zsh

## 2. monom_args Function

- [ ] 2.1 Implement `monom_args` in `src/monom` — parse declaration block, split on `--`, loop through flag specs
- [ ] 2.2 Handle modifier grouping — `--`-prefixed tokens before a bare word apply to that flag
- [ ] 2.3 For value flags, set variable via `printf -v` (or zsh equivalent) from `monomd args` output
- [ ] 2.4 For `--boolean` flags, set variable to `"true"` (exit 0) or `""` (exit 1) based on `monomd args` exit code
- [ ] 2.5 Error if `--` separator is missing from the call
- [ ] 2.6 Verify it works in both bash and zsh

## 3. Shell Tests

- [ ] 3.1 Create `tests/monom_error_test` shUnit2 test file
- [ ] 3.2 Test default exit code (1)
- [ ] 3.3 Test custom exit code
- [ ] 3.4 Test message printed verbatim to stderr (no prefix)
- [ ] 3.5 Create `tests/monom_args_test` shUnit2 test file
- [ ] 3.6 Test single value flag sets variable
- [ ] 3.7 Test multiple value flags set variables
- [ ] 3.8 Test absent flag sets empty variable
- [ ] 3.9 Test `--short` modifier passed through
- [ ] 3.10 Test `--boolean` flag sets "true" or empty
- [ ] 3.11 Test mixed modifiers across multiple flags
- [ ] 3.12 Test missing `--` separator errors

## 4. Validation

- [ ] 4.1 Run shellcheck on `src/monom` — no new errors
- [ ] 4.2 Run all shUnit2 tests — pass in bash
- [ ] 4.3 Run all shUnit2 tests — pass in zsh
