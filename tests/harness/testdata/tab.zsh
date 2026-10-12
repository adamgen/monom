# Presses Tab at the end of $_case_line in zsh: split it the way the line
# editor does, then call the function compdef registered in $_comps.
# cases_test.go runs this after compinit, sourcing src/monom, and assigning
# _case_line.

words=(${(z)_case_line})
[[ $_case_line == *' ' || -z $_case_line ]] && words+=('')
CURRENT=${#words}
PREFIX=${words[CURRENT]}
_fn=${_comps[${words[1]}]}
[[ -n $_fn ]] || { print -u2 "no zsh completion registered for '${words[1]}'"; exit 98; }
# Candidates reach zsh through compadd, which offers only those matching the
# word under the cursor. Print exactly those instead, skipping options.
compadd() {
  while (( $# )) && [[ $1 == -* ]]; do [[ $1 == -- ]] && { shift; break; }; shift; done
  local c
  for c in "$@"; do [[ $c == "$PREFIX"* ]] && printf '%s\n' "$c"; done
  return 0
}
"$_fn"
