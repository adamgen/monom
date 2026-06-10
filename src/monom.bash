#!/usr/bin/env bash
# src/monom.bash — bash-specific completion binding for monom.
# Sourced by src/monom when running under bash.

# monom_completion — bash completion handler registered via `complete -F`.
# Always exits 0 and never writes to stderr; completion is interactive and
# any noise mid-typing degrades the user experience.
monom_completion() {
  COMPREPLY=()
  _monom_log "[bash] completion triggered: words=(${COMP_WORDS[*]})"
  if ! setup_monom 2>/dev/null; then
    _monom_log "[bash] setup_monom failed"
    return 0
  fi
  _monom_log "[bash] setup_monom OK: root=$MONOM_PROJECT_ROOT"
  local raw_completions
  raw_completions=$(monom_cfg complete 2>/dev/null)
  _monom_log "[bash] monom_cfg complete: $(printf '%s' "$raw_completions" | wc -l | tr -d ' ') lines"
  # $MONOM_BIN is the monomd binary resolved at source time in src/monom — used
  # instead of the bare name `monomd`, which may only exist as a user alias.
  # shellcheck disable=SC2207
  COMPREPLY=($(printf '%s' "$raw_completions" | "$MONOM_BIN" filter "${COMP_WORDS[@]:1}" 2>/dev/null))
  _monom_log "[bash] COMPREPLY=(${COMPREPLY[*]})"
  return 0
}

complete -F monom_completion monom
