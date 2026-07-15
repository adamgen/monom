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

- **`complete` (discovery)** — prints every folder that contains a `run.sh`,
  slash-delimited, one per line. This is the command tree monom completes against.

  ```
  $ ./monom complete
  db/migrate
  db/seed
  infra/cloud/deploy
  infra/cloud/teardown
  infra/local/start
  infra/local/stop
  release
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

This is the seam for completing additional arguments and flags on top of the base
`monom <command>` completion. It keeps a command's runner and its completion logic
colocated in one folder.
