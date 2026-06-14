## ADDED Requirements

### Requirement: monomd install writes a source line to the user's rc file
`monomd install` SHALL detect the user's shell via `$SHELL`, resolve the absolute path to `src/monom` relative to the running binary, select the appropriate rc/profile file, check for an existing source line, and append the line if absent. On success it SHALL print the rc file path it modified and tell the user to restart their shell or re-source.

#### Scenario: Install into zsh rc file
- **WHEN** `$SHELL` ends in `/zsh` and `~/.zshrc` does not already contain a source line for `src/monom`
- **THEN** `monomd install` appends `source "/absolute/path/to/src/monom"` to `~/.zshrc`, prints the path of the modified file, and exits 0

#### Scenario: Install into bash profile file
- **WHEN** `$SHELL` ends in `/bash` and `~/.bash_profile` does not already contain a source line for `src/monom`
- **THEN** `monomd install` appends `source "/absolute/path/to/src/monom"` to `~/.bash_profile`, prints the path of the modified file, and exits 0

#### Scenario: Already installed — idempotent
- **WHEN** the target rc file already contains a line with the resolved `src/monom` path
- **THEN** `monomd install` prints "already installed" and exits 0 without modifying the file

#### Scenario: Unknown shell
- **WHEN** `$SHELL` does not end in `/zsh` or `/bash`
- **THEN** `monomd install` prints an error naming the detected shell and exits non-zero

### Requirement: monomd install resolves the src/monom path via the running binary
`monomd install` SHALL resolve the path to `src/monom` by calling `os.Executable()`, resolving any symlinks via `filepath.EvalSymlinks`, then computing `../src/monom` relative to the binary's directory.

#### Scenario: Binary is a symlink
- **WHEN** `monomd` is invoked via a symlink (e.g. a Homebrew shim)
- **THEN** the resolved `src/monom` path is based on the real binary location, not the symlink

#### Scenario: Binary is at expected location
- **WHEN** `monomd` is at `<root>/bin/monomd`
- **THEN** the resolved path is `<root>/src/monom`

### Requirement: monomd install prepends a newline before the appended line
When appending to an rc file that does not end with a newline, `monomd install` SHALL prepend a newline to ensure the source line appears on its own line.

#### Scenario: RC file missing trailing newline
- **WHEN** the target rc file's last byte is not a newline
- **THEN** the appended content is `\nsource "/path/to/src/monom"\n`

#### Scenario: RC file ends with a newline
- **WHEN** the target rc file's last byte is a newline
- **THEN** the appended content is `source "/path/to/src/monom"\n`
