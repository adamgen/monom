## ADDED Requirements

### Requirement: Mapping table format with file entries and pkgjson directive entries
Both `mnmd resolve-run` and `mnmd resolve-complete` SHALL take a command mapping table as their first CLI argument: a string of newline-separated entries with whitespace-delimited fields. Blank lines and lines whose first field starts with `#` SHALL be skipped. Two entry kinds exist, distinguished by the first field:

- **File entry** — `<value> <key words...>`: `value` is an executable path (absolute, or project-root-relative) and the remaining fields are the key words it maps from. Lines with a value but no key words SHALL be skipped (a bare value can never match).
- **Directive entry** — `@pkgjson <package.json-path> [prefix]` (first field exactly `@pkgjson`): expands in place to one entry per script in that manifest's `scripts` object, keyed by the prefix words followed by the script name. `prefix` is a single slash-delimited token (e.g. `web`, `apps/web`); when omitted, the scripts are keyed at the top level. The path is absolute or project-root-relative (resolved via the same algorithm as `mnmd root`).

Entries — across both kinds — are processed in table order.

#### Scenario: Blank lines and comments are skipped
- **WHEN** the table contains blank lines and lines starting with `#` between valid entries
- **THEN** the valid entries are processed and the blank/comment lines contribute nothing

#### Scenario: Value-only file entry is skipped
- **WHEN** a line contains a single field that is not `@pkgjson`
- **THEN** that line contributes no entry

#### Scenario: Directive with a relative path resolves against the project root
- **WHEN** a directive line is `@pkgjson apps/web/package.json web` and the project root is `/repo`
- **THEN** the manifest at `/repo/apps/web/package.json` is used for that entry

#### Scenario: Missing table argument is a usage error
- **WHEN** `mnmd resolve-run` or `mnmd resolve-complete` is invoked with no table argument
- **THEN** a usage message is printed to stderr and the exit code is non-zero

### Requirement: resolve-run maps file entries to their value with passthrough on miss
`mnmd resolve-run <table> <word...>` SHALL print the value of the first file entry whose key words equal the words, followed by a newline, and exit 0. When no entry in the table matches, it SHALL print the words joined back into a single space-separated line and exit 0 with nothing on stderr — a miss is a non-event, so unmapped commands flow on to default tree resolution with no fallback branch in the hook. Invocation with a table but no command words SHALL exit non-zero with an error message on stderr.

#### Scenario: Matching key words print the value
- **WHEN** the table maps `/repo/scripts/migrate_db.sh` from key `db migrate` and the words are `db migrate`
- **THEN** stdout is `/repo/scripts/migrate_db.sh` and the exit code is 0

#### Scenario: Miss passes the words through unchanged
- **WHEN** the words are `release now` and no entry matches
- **THEN** stdout is `release now`, the exit code is 0, and nothing is printed to stderr

#### Scenario: No command words is an error
- **WHEN** `mnmd resolve-run <table>` is invoked without words
- **THEN** an error message is printed to stderr and the exit code is non-zero

### Requirement: resolve-complete prints every entry's key in discovery format
`mnmd resolve-complete <table>` SHALL print one line per entry, in table order, in the slash-delimited discovery format the completion pipeline expects. A file entry prints its key words joined with `/` (`db migrate` → `db/migrate`). A directive entry prints `<prefix>/<script-name>` per script (or `<script-name>` with no prefix), with script names in the order they appear in the manifest file. Mapped commands therefore tab-complete exactly like tree commands. File entries SHALL be printed without touching the filesystem — values need not exist on disk.

#### Scenario: File entry keys print as slash-delimited paths
- **WHEN** the table maps values from keys `db migrate` and `db seed`
- **THEN** stdout contains `db/migrate` and `db/seed`, one per line, in table order

#### Scenario: Prefixed directive prints prefixed script paths
- **WHEN** the table contains `@pkgjson /repo/apps/web/package.json web` and that manifest's scripts are `build` then `test`
- **THEN** stdout contains `web/build` then `web/test`, and the exit code is 0

#### Scenario: Directive without a prefix prints top-level script names
- **WHEN** a directive line has a path but no prefix and the manifest's scripts are `build` then `lint`
- **THEN** stdout contains `build` then `lint`

#### Scenario: Mixed table prints in table order
- **WHEN** the table has the file entry `db migrate` on the first line and a `web`-prefixed directive on the second
- **THEN** `db/migrate` is printed before any `web/...` line

#### Scenario: Script name with whitespace or slash is skipped silently
- **WHEN** a directive's manifest defines scripts `build`, `bad name`, and `bad/name`
- **THEN** only the `build` path is printed for that directive and nothing is printed to stderr for the skipped names

#### Scenario: Unreadable manifest warns on stderr but does not break discovery
- **WHEN** the table has a directive whose package.json does not exist or is not valid JSON, followed by other entries
- **THEN** a warning naming that path is printed to stderr, every other entry's lines are still printed to stdout, and the exit code is 0

### Requirement: resolve-run maps directive entries to a package-manager exec line
For directive entries, `mnmd resolve-run <table> <word...>` SHALL treat the words as a hit when they equal an entry's prefix words followed by a name in that manifest's `scripts` object. On a hit it SHALL print a single space-separated exec line — the package manager's absolute path followed by the arguments that run the script in the package's directory — and exit 0. First match wins across both entry kinds, in table order, so the author's line ordering decides precedence between a legacy file mapping and a same-keyed script. Directive manifests SHALL be read lazily: an entry's manifest is only opened when the typed words shape-match it (its prefix words plus exactly one remaining word).

The invocation form per package manager (`<dir>` is the absolute directory containing the directive's package.json):

- npm: `<npm> --prefix <dir> run <script>`
- pnpm: `<pnpm> -C <dir> run <script>`
- yarn: `<yarn> --cwd <dir> run <script>`
- bun: `<bun> --cwd <dir> run <script>`

#### Scenario: Directive hit resolves to an npm exec line by default
- **WHEN** the table contains `@pkgjson /repo/apps/web/package.json web`, that manifest has script `build` and no `packageManager` field, and the words are `web build`
- **THEN** stdout is one line — the absolute path of `npm` followed by `--prefix /repo/apps/web run build` — and the exit code is 0

#### Scenario: Top-level directive matches a single word
- **WHEN** a directive has no prefix, its manifest has script `lint`, and the words are `lint`
- **THEN** stdout is the exec line running the `lint` script and the exit code is 0

#### Scenario: Earlier file entry shadows a same-keyed script
- **WHEN** a file entry keyed `web build` precedes a `web`-prefixed directive whose manifest also defines `build`, and the words are `web build`
- **THEN** stdout is the file entry's value

#### Scenario: Prefix match without a known script is a miss
- **WHEN** a directive has prefix `web`, its manifest defines only `build`, and the words are `web deploy` with no other entry matching
- **THEN** stdout is `web deploy` and the exit code is 0

#### Scenario: Words that shape-match no directive read no manifest
- **WHEN** the words are `db migrate` and they match a file entry while the table also contains directives
- **THEN** the file entry's value is printed without any manifest being opened

### Requirement: resolve-run detects the package manager from the packageManager field
On a directive hit, `mnmd resolve-run` SHALL choose the package manager from that manifest's `packageManager` field: the name before `@` when the field is present (recognized names: `npm`, `pnpm`, `yarn`, `bun`), and `npm` when the field is absent. The chosen binary SHALL be resolved to an absolute path via a `$PATH` lookup at run time, on a hit only. An unrecognized package manager name, or a recognized one not found on `$PATH`, SHALL cause a non-zero exit with an error message naming it.

#### Scenario: packageManager selects pnpm
- **WHEN** the hit directive's manifest contains `"packageManager": "pnpm@9.1.0"` and `pnpm` is on `$PATH`
- **THEN** stdout is the absolute path of `pnpm` followed by `-C <dir> run <script>`

#### Scenario: Unrecognized package manager is a loud error
- **WHEN** the hit directive's manifest contains `"packageManager": "deno@2.0.0"`
- **THEN** an error message naming `deno` is printed to stderr and the exit code is non-zero

#### Scenario: Package manager missing from PATH is a loud error
- **WHEN** the hit directive's manifest selects `pnpm` and `pnpm` is not on `$PATH`
- **THEN** an error message naming `pnpm` is printed to stderr and the exit code is non-zero

### Requirement: resolve-run fails loudly when a shape-matched directive cannot be resolved
When the typed words shape-match a directive entry — its prefix words equal the leading words with exactly one word remaining — but that manifest cannot be read or parsed, `mnmd resolve-run` SHALL exit non-zero with an error message naming the manifest path, rather than passing the words through: falling through to tree resolution would turn a config bug into a baffling "command not found" for a command that tab-completed moments ago. Likewise, on a directive hit whose package directory or script name contains whitespace, it SHALL exit non-zero naming the offending value — whitespace cannot cross the space-split hook output boundary.

#### Scenario: Unreadable manifest for shape-matched words is an error
- **WHEN** the words are `web build`, a directive has prefix `web`, no earlier entry matches, and the directive's package.json does not exist
- **THEN** an error message naming that path is printed to stderr and the exit code is non-zero

#### Scenario: Package directory containing whitespace is an error
- **WHEN** the hit directive's package.json lives under a directory whose path contains a space
- **THEN** an error message naming that directory is printed to stderr and the exit code is non-zero
