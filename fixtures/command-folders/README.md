# command-folders fixture

A self-contained example monom project that models a **different structure** from
[`demo-project`](../demo-project).

In `demo-project`, each command is a single executable file. Here, each command is
a **folder** that bundles two files:

| File          | Role                                                                       |
| ------------- | -------------------------------------------------------------------------- |
| `run.sh`      | The executable the command actually runs.                                  |
| `complete.sh` | Prints extra completions (flags/arguments) for tokens typed *after* the command. |

A folder is a **command** when it contains a `run.sh`. Folders without one are plain
**command categories** — the file tree is still the command tree.

## Tree

```
command-folders/
├── monom                       # config: implements `complete` and `run` hooks
├── release/                    # command
│   ├── run.sh
│   └── complete.sh
├── db/                         # category
│   ├── migrate/                # command
│   │   ├── run.sh
│   │   └── complete.sh
│   └── seed/                   # command
│       ├── run.sh
│       └── complete.sh
└── infra/                      # category
    ├── cloud/                  # category
    │   ├── deploy/             # command
    │   │   ├── run.sh
    │   │   └── complete.sh
    │   └── teardown/           # command
    │       ├── run.sh
    │       └── complete.sh
    └── local/                  # category
        ├── start/              # command
        │   ├── run.sh
        │   └── complete.sh
        └── stop/               # command
            ├── run.sh
            └── complete.sh
```

## How the structure maps onto monom

The `monom` config file bridges this folder-per-command layout to monom's model
using two hooks (see `architecture.md`):

- **`complete` (discovery)** — prints every command (a folder containing a
  `run.sh`), slash-delimited, and, for each, one deeper path per extra completion
  its `complete.sh` emits (e.g. `infra/cloud/deploy/--region`). This single
  stream is the command tree monom completes against — see
  [Per-command argument completion](#per-command-argument-completion-completesh).

  ```
  $ ./monom complete
  db/migrate
  db/migrate/--step
  db/migrate/--to
  db/migrate/--dry-run
  db/seed
  db/seed/--count
  db/seed/--truncate
  infra/cloud/deploy
  infra/cloud/deploy/--region
  infra/cloud/deploy/--profile
  infra/cloud/deploy/--force
  infra/cloud/deploy/--dry-run
  ...
  release
  release/--dry-run
  release/--tag
  release/--skip-tests
  ```

- **`run` (routing)** — receives the user's space-separated command tokens and
  appends `run.sh`, so `mnmd pack` resolves the real executable inside the folder.
  Bare categories pass through unchanged so monom's default command-group handling
  still applies.

  ```
  $ ./monom run infra cloud deploy
  infra cloud deploy run.sh
  # → mnmd pack infra cloud deploy run.sh → <root>/infra/cloud/deploy/run.sh
  ```

## Per-command argument completion (`complete.sh`)

Each command folder carries a `complete.sh` that prints the flags/arguments valid
*after* that command, one per line:

```
$ infra/cloud/deploy/complete.sh
--region
--profile
--force
--dry-run
```

**These run during tab completion.** The completion binding feeds one stream into
the filter — `_monom_cfg complete | mnmd filter <typed words>` — so `complete` is
the single source of truth for what `<Tab>` offers. The `complete` hook therefore
invokes each command's `complete.sh` and re-emits its output as deeper paths under
that command. `mnmd filter` then treats those as the command's children:

```
$ monom infra cloud deploy <Tab>
--region  --profile  --force  --dry-run
```

This keeps a command's runner (`run.sh`) and its completion logic (`complete.sh`)
colocated in one folder. Note the cost: `complete` spawns each command's
`complete.sh` on every `<Tab>` — acceptable for a small tree, but an author with a
large tree would cache or lazily resolve these.
