## 1. Implementation

- [x] 1.1 Implemented `_mnmd_completion` in `src/monom.bash` — documented retrospectively
- [x] 1.2 Implemented `_mnmd` in `src/monom.zsh` — documented retrospectively
- [x] 1.3 Added `tests/mnmd_completion_test` (8 shUnit2 tests, all passing) — documented retrospectively

## 2. Gaps

- [x] 2.1 Add test coverage for the `compdef` guard in zsh (sourcing before `compinit`) — currently not exercised by `tests/mnmd_completion_test`
- [ ] 2.2 Consider documenting the hardcoded subcommand list as a source-of-truth concern: if a new subcommand is added to `main.go`, `src/monom.bash` and `src/monom.zsh` must be updated manually
