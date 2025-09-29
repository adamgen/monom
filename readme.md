# Monom

## File structure ideation

Options:

1. Folder per command with many files that represent all command features
2. File per command with
  1. In file args checks that run the command or the command features conditionally
  2. Fnction exports that represent different command features

## Todo

- [ ] When using bash +x, how does this propogate to subshells?
- [ ] Make a command sourced instead of executed in a separate shell

### Critical for release

- [ ] Decide what happens when you have no run command but is only a namespace for subcommands
- [ ] Distinc between the "monom lib project root" and the "active monom project root"
- [ ] Give the project a better structure
   - [ ] Utils should move to a different root directory
   - [ ] Bin folder with all scripts
   - [ ] Map what's executed in a separate vs same shell, possibly name these shells for refference
   - [ ] Map all env vars for reference
   - [ ] Move all log files to a logs folder
   - [ ] Consider - Rename test_projects to integration/e2e test and move `_test` files into each project's folder
- [ ] Clean up comments and improve documentation
- [ ] Make monom's base commands (init, help, usage...) and their completion
- [ ] Make sure the test runner prints the "run this test TEST_COMMAND" correctly
- [ ] Have a singular convention for cd to the current file's folder

### Postponed

- [ ] Have nicer support for zsh
- [ ] Make the "map the completion back to the `run` file" an extension
- [ ] Finish the `complete_override` function that will allow a function based completion instead of a file based
- [ ] Publis to brew
- [ ] Add a solid `--args` function
- [ ] Add functions that run on the same shell (like sourcing python)

## Vision

- Have a oneline install utility that allows you to write CLI tools easily
- File as commands, file tree as command tree (commands & subcommands)
- Modularity / extensibility - have hooks where developers can extend both a single command and the entire project. Examples:
  - Cache & S3 cache
  - Web ui
  - Tests
  - Docs/manual
- Have an easy to understand convention for accepting cli arguments
- Support a global monom and a project specific monom (will project specific run from anywhere?)

## Motivation

[//]: # "todo"

Writing cli...

Make it easy to add test suits for each bash utility

## Goal

1. A tool for easily building cli utilities in any language with free bash completion.
2. Support independent cli utilities as well as a monorepo manager.
3. Opt-in to cache outputs of a command based on predefined inputs.
4. Easy to write documentation and `--help commands`
5. Make it easy to understand and test shell utilities using an opinionated structure and educational content.
6. Having many cli tools for different project on the same machine working together without conflicts

All while following [clig guidelines](https://clig.dev/)

## CLI tool vs monorepo

The 2 main use cases for cli usage are: a standalone cli tool, like `git` or the `aws` cli. And a company internal
project with many apps and libraries that are heavily correlated like `ui`, `back` and `interface`.

Ideally, the monom project can define a broad enough common api to serve both cases.

### Standalone cli tool structure

```tree
command-1
command-2
category/
 ├── sucbommand-2
 └── sucbommand-2
```

Assuming your base command is proj, typing `Tab` after `proj` will give you the suggestions: `command-1`, `command-2`
and `category`.

Typing `Tab` when the `proj command-1` will print its arguments as the suggestion if they exist, and nothing if they
don't.

Typing `Tab` when the `proj category` will print `subcommand-1` and `subcommand-2`. For each of them clicking `Tab` will
behave like it does for commands on the root level.

### Monorepo cli structure

Assuming the team developing the apps is called Acme, hence the cli root is called `acme`, and they have the following
projects and commands:

```tree
ui/
└── commands/
    ├── run
    ├── test
    ├── e2e
    └── build
backend/
└── commands
    ├── run
    ├── test
    └── build
commands
├── deploy
├── ci
├── run-all
└── status
    ├── back
    └── ui
```

Typing `Tab` in the following scenarios will give these results:

1. `acme` will suggest `ui`, `backend`, `deploy`, `ci`, `run-all`
2. `acme run` will suggest it's arguments same way a command in a standalone cli.
3. All commands, categories and their commands will behave in the primitive command paradigm.

## Concepts

### Abstract concepts

1. Root namespace - the name of the cli utility, or the name of the monorepo in the cli under which all commands and
   command categories are listed.
2. Primitive command - a command that follows an interface. A command can be a first descendants of the root namespace.
3. Command category - a group of commands listed under a single command. The command category command itself only prints
   a `usage` or `help` guidelines.

### monom concepts

1. Author and user - the author writes the cli library, the user uses the cli library.
2. User vs monom interfaces - the cli author write a "monom" app exposing commands that monom uses hence the author
   writes for the monom interface. Monom uses the user's commands and usage of monom interfaces to generate the user
   interface.
3. Primitive command trait/interface - a list of subcommands each command is exposing that monom uses internally to
   generate it's user interface like completion and input validation.
4. Command group/category - categories are used for encapsulation and discovery of many commands that relate closely. An
   example is a monorepo project `back` having its own `run` and `test` scripts. `back` is a command group with several
   commands that operate the backend app.

## Tech design

### Primitive command

A primitive command is a command where the author exposes several subcommands that monom uses to generate the user cli.

There are the commands that each primitive command should expose:

1. `run`
   1. Without file extension and with shebang
   2. Has an execution privilege
2. `complete`
3. `dependencies`
4. `inputs` - function arguments, stdin, and env vars

Each command is either stored in a dedicated folder with the above commands listed as files. Or have a `monom` file
that has exposes the above commands as its subcommands.
