---
name: monom-bash-conventions
description: Best practices for writing bash/sh code in the monom project — shUnit2 tests, helper functions, heredocs, and test structure. Use when writing or editing shell scripts, shUnit2 tests, or test helpers in the monom repo.
---

# monom bash conventions

## shUnit2 test structure

- Test files have no extension, named `${script_name}_test`, colocated with the file under test.
- Test functions are named `test_descriptive_name()`.
- Use `assertEquals`, `assertNotEquals` for exit codes.
- For line-exact output matching, use `grep -qFx` rather than substring assertions — a substring check on `command1` will also match `sub_command1`.

When a test suite needs line-exact output assertions, add shared helpers along these lines rather than using raw `assertContains`:

```sh
assert_output_contains_line() {
  local msg="$1" output="$2" line="$3"
  if ! printf '%s\n' "$output" | grep -qFx "$line"; then
    fail "$msg: expected line <$line> not found in: $output"
  fi
}

assert_output_lacks_line() {
  local msg="$1" output="$2" line="$3"
  if printf '%s\n' "$output" | grep -qFx "$line"; then
    fail "$msg: unexpected line <$line> found in: $output"
  fi
}
```

## Multi-line input

Pipe multi-line input into a command using a heredoc, not `printf '...\n...'`:

```sh
result=$(cmd arg <<'EOF'
line1
line2
EOF
)
```

When only the exit code matters, skip the subshell wrapper:

```sh
cmd arg <<'EOF'
line1
line2
EOF
assertEquals "..." 0 $?
```

## Writing file content

Use `cat > path <<'EOF'` with real newlines, not `printf` with `\n` escapes. Heredocs show the actual file content as humans would read it — `\n` escape sequences obscure structure and are harder to scan at a glance:

```sh
# good
cat > "$path" <<'EOF'
#!/bin/sh
echo ok
EOF
chmod +x "$path"

# avoid
printf '#!/bin/sh\necho ok\n' > "$path"
```

Use unquoted `<<EOF` only when the heredoc body needs variable expansion.

## Test helper functions

- Keep helpers at the smallest useful granularity.
- Inline single-use functions — if a function is only ever called from one place, fold its body into that call site and remove the function.
- If the same setup sequence appears in three or more tests, extract it into a named helper. Name it after what it produces, not what it does.

## macOS symlink resolution

`mktemp -d` on macOS returns a path under `/var/folders/...`, which is a symlink to `/private/var/folders/...`. Any tool that resolves symlinks will return the `/private/` form. Always resolve temp dirs with `cd "$d" && pwd -P` before comparing paths.

## Cleanup

Each test creates its own isolated temp dir and cleans up at the end of its own body with `rm -rf`. No global `tearDown` is needed.
