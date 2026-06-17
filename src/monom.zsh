#!/usr/bin/env zsh
# src/monom.zsh — zsh-specific completion binding for monom.
# Sourced by src/monom when running under zsh.
# Requires compinit to have been called before this file is sourced
# (standard zsh setup order).

# _monom — zsh completion function registered via compdef.
# Always exits 0 and never writes to stderr; completion is interactive and
# any noise mid-typing degrades the user experience.
_monom() {
  # $words is set by zsh's completion system before _monom is called.
  # SC2154: referenced but not assigned — false positive for zsh completion variables.
  # shellcheck disable=SC2154
  _monom_log "[zsh] completion triggered: words=(${words[*]})"
  if ! _setup_monom 2>/dev/null; then
    _monom_log "[zsh] _setup_monom failed"
    return 0
  fi
  _monom_log "[zsh] _setup_monom OK: root=$_MONOM_PROJECT_ROOT"
  local raw_completions
  raw_completions=$(_monom_cfg complete 2>/dev/null)
  _monom_log "[zsh] _monom_cfg complete: $(printf '%s' "$raw_completions" | wc -l | tr -d ' ') lines"
  _monom_log "[zsh] raw_completions first line: $(printf '%s' "$raw_completions" | head -1)"
  local filter_words=("${words[@]:1}")
  _monom_log "[zsh] filter words=(${filter_words[*]})"
  local filter_output
  filter_output=$(printf '%s\n' "$raw_completions" | mnmd filter "${filter_words[@]}" 2>/dev/null)
  _monom_log "[zsh] filter raw output: $(printf '%s' "$filter_output" | wc -l | tr -d ' ') lines: $filter_output"
  # No completions (leaf reached): return without calling compadd. Splitting an
  # empty string with ${(@f)...} yields a single empty-string element, and
  # `compadd -- ""` registers an empty match that zsh "completes" by inserting a
  # trailing space on every Tab. Bailing out here avoids that spurious space.
  if [ -z "$filter_output" ]; then
    _monom_log "[zsh] no completions; skipping compadd"
    return 0
  fi
  local -a completions
  # ${(@f)var} is zsh syntax: the @f flag splits $var on newlines into an array.
  # SC2296 is a false positive here — the linter does not know zsh flags.
  # shellcheck disable=SC2296
  completions=("${(@f)filter_output}")
  _monom_log "[zsh] completions=(${completions[*]})"
  compadd -- "${completions[@]}"
  return 0
}

# Guard against sourcing before compinit has been called. compdef is defined
# by compinit; without it the registration would print an error to the terminal.
# Standard .zshrc order has compinit run first, so this guard is a safety net.
if (( ${+functions[compdef]} )); then
  compdef _monom monom
fi
