# shellcheck shell=bash
# Presses Tab at the end of $_case_line in bash, the way readline does: set
# COMP_* and call the function `complete -p <word>` registered, then print
# COMPREPLY one per line. cases_test.go runs this after sourcing src/monom and
# assigning _case_line.
# shellcheck disable=SC2154 # _case_line is assigned by the runner
# shellcheck disable=SC2034 # COMP_* are read by the completion function

COMP_LINE=$_case_line
read -ra COMP_WORDS <<< "$COMP_LINE"
case "$COMP_LINE" in *' '|'') COMP_WORDS+=("") ;; esac
COMP_POINT=${#COMP_LINE}
COMP_CWORD=$(( ${#COMP_WORDS[@]} - 1 ))
_spec=$(complete -p "${COMP_WORDS[0]}" 2>/dev/null) || { echo "no bash completion registered for '${COMP_WORDS[0]}'" >&2; exit 98; }
_fn=${_spec##*-F }
_fn=${_fn%% *}
COMPREPLY=()
"$_fn" "${COMP_WORDS[0]}" "${COMP_WORDS[COMP_CWORD]}" "${COMP_WORDS[COMP_CWORD-1]}"
[ ${#COMPREPLY[@]} -eq 0 ] || printf '%s\n' "${COMPREPLY[@]}"
