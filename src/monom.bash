#!/usr/bin/env bash
# src/monom.bash — bash-specific completion binding for monom.
# Sourced by src/monom when running under bash.

# _monom_completion — bash completion handler registered via `complete -F`.
# Always exits 0 and never writes to stderr; completion is interactive and
# any noise mid-typing degrades the user experience.
_monom_completion() {
  COMPREPLY=()
  _monom_log "[bash] completion triggered: words=(${COMP_WORDS[*]})"
  if ! _setup_monom 2>/dev/null; then
    _monom_log "[bash] _setup_monom failed"
    return 0
  fi
  _monom_log "[bash] _setup_monom OK: root=$_MONOM_PROJECT_ROOT"
  local raw_completions
  raw_completions=$(_monom_cfg complete 2>/dev/null)
  _monom_log "[bash] _monom_cfg complete: $(printf '%s' "$raw_completions" | wc -l | tr -d ' ') lines"
  # shellcheck disable=SC2207
  COMPREPLY=($(printf '%s' "$raw_completions" | mnmd filter "${COMP_WORDS[@]:1}" 2>/dev/null))
  _monom_log "[bash] COMPREPLY=(${COMPREPLY[*]})"
  return 0
}

complete -F _monom_completion monom
