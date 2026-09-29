import type { MetaFunction } from "@remix-run/node";
import { Code } from "~/components/Code";
import { CopyCommand } from "~/components/CopyCommand";
import { FileTree } from "~/components/FileTree";
import { CHECK } from "~/lib/project";
import { GITHUB_URL, INSTALL } from "~/lib/site";

export const meta: MetaFunction = () => [
  { title: "Quick start — monom" },
  {
    name: "description",
    content:
      "Install monom, turn the scripts in your repo into a tab-completable CLI with no config file, and check it in CI.",
  },
];

const TOC = [
  ["install", "Install"],
  ["project", "Pick a project"],
  ["commands", "Add commands"],
  ["try", "Press Tab"],
  ["check", "Check it in CI"],
  ["config", "The monom file"],
  ["hooks", "Hook scripts"],
  ["anywhere", "Use it from anywhere"],
  ["reference", "Reference"],
  ["troubleshooting", "Troubleshooting"],
] as const;

export default function Docs() {
  return (
    <main className="wrap docs">
      <aside className="docs-nav" aria-label="On this page">
        <p className="kicker">quick start</p>
        <ol>
          {TOC.map(([id, label]) => (
            <li key={id}>
              <a href={`#${id}`}>{label}</a>
            </li>
          ))}
        </ol>
      </aside>

      <article className="prose">
        <h1>Quick start</h1>
        <p className="lede">
          From nothing to a tab-completing <code>monom infra cloud deploy.sh</code>, with no config
          file. You'll need <code>git</code>, <code>make</code>, Go 1.24+, and bash or zsh.
        </p>

        <h2 id="install">1. Install</h2>
        <p>
          Clone the repo somewhere permanent, build the <code>mnmd</code> engine, and let it add
          one <code>source</code> line to your shell's rc file.
        </p>
        <CopyCommand command={INSTALL} />
        <p>
          <code>mnmd install</code> picks <code>~/.zshrc</code> for zsh, or{" "}
          <code>~/.bash_profile</code> (falling back to <code>~/.bashrc</code>) for bash. It's
          idempotent: a second run prints <code>already installed</code>. Open a new shell, or
          source the file it names:
        </p>
        <Code output>{`$ ~/.monom/bin/mnmd install
added to /Users/you/.zshrc
restart your shell or run: source /Users/you/.zshrc`}</Code>
        <p>
          The rc line points into the checkout, so don't move it or copy the binary out of it. In
          zsh, the line has to come after <code>compinit</code> (or after the framework line that
          calls it, like <code>source $ZSH/oh-my-zsh.sh</code>). Updating later is{" "}
          <code>git -C ~/.monom pull &amp;&amp; make -C ~/.monom build</code>.
        </p>
        <p className="callout">
          Setting this up with a coding agent? Point it at{" "}
          <a href={`${GITHUB_URL}/blob/main/INSTALL.md`}>INSTALL.md</a>, a step-by-step procedure
          with a check after every step.
        </p>

        <h2 id="project">2. Pick a project</h2>
        <p>
          Any git repository is already a monom project: its root is the project root. Outside git,
          mark a directory with an empty <code>monom</code> file. monom finds the root by walking
          up from wherever you are, first match wins:
        </p>
        <ol>
          <li>
            <code>$_MONOM_PROJECT_ROOT</code>, if you've pinned one (see{" "}
            <a href="#anywhere">below</a>);
          </li>
          <li>
            the nearest directory containing a file named <code>monom</code>;
          </li>
          <li>the nearest git root.</li>
        </ol>
        <p>The current directory alone is never a root.</p>
        <Code>{`$ mkdir acme && cd acme
$ touch monom          # or: git init`}</Code>

        <h2 id="commands">3. Add commands</h2>
        <p>
          Each command is an executable file; each subfolder is a command group. Use any language:
          the shebang decides, and compiled binaries work too. The examples on this site all use
          this project:
        </p>
        <div className="tree-card">
          <FileTree />
        </div>
        <p>Start with two of its commands:</p>
        <Code>{`$ mkdir -p infra/cloud db
$ printf '#!/usr/bin/env bash\\necho "deploying to cloud..."\\n' > infra/cloud/deploy.sh
$ printf '#!/usr/bin/env python3\\nprint("running db migrations...")\\n' > db/migrate.py
$ chmod +x infra/cloud/deploy.sh db/migrate.py`}</Code>
        <Code output>{`$ mnmd discover
db/migrate.py
infra/cloud/deploy.sh`}</Code>
        <p>
          That's it: there is no registration step. Default discovery registers an executable when
          it passes the <strong>gate</strong>:
        </p>
        <ul>
          <li>
            it starts with a shebang (<code>#!</code>), or
          </li>
          <li>
            its name looks like a command, <code>^[a-z0-9][a-z0-9_-]*$</code>: lowercase, no
            dots. This is how compiled binaries like <code>tools/lint</code> get in; or
          </li>
          <li>
            the <a href="#config">monom file</a> declares it.
          </li>
        </ul>
        <p>
          Hidden and <code>_</code>-prefixed files and folders, <code>node_modules</code>,{" "}
          <code>vendor</code>, <code>__pycache__</code>, <code>venv</code>, <code>target</code>,{" "}
          <code>dist</code>, and nested monom projects are never scanned. Paths with spaces are
          never registered, because they couldn't be typed as words.
        </p>
        <p>
          The file's path <em>is</em> the command, extension included:{" "}
          <code>db/migrate.py</code> is <code>monom db migrate.py</code>, and Tab completes the
          whole name. Since the name rule has no dots, a file with an extension needs a shebang;
          without one (<code>tools/Format.sh</code>) it isn't registered, and{" "}
          <code>mnmd check</code> says why. Drop the extension, as in{" "}
          <code>infra/local/start</code>, if you want a shorter command.
        </p>

        <h2 id="try">4. Press Tab</h2>
        <Code output>{`$ monom <Tab>
db     infra
$ monom in<Tab>         # completes to: monom infra
$ monom infra cloud deploy.sh
deploying to cloud...
$ monom infra
monom: 'infra' is a command group
available: cloud`}</Code>
        <p>
          Typing a folder instead of a script lists what's inside, the same list Tab would show.
          Nothing runs. Commands work from any folder inside the project.
        </p>
        <p className="callout">
          Every word you type is part of the command path, so commands don't take arguments yet:{" "}
          <code>monom release.rb v1.2</code> looks for <code>release.rb/v1.2</code>. Flag parsing (
          <code>mnmd args</code>) is on the <a href={`${GITHUB_URL}/blob/main/BACKLOG.md`}>backlog</a>.
        </p>

        <h2 id="check">5. Check it in CI</h2>
        <p>
          Completion stays silent when something's wrong, so it never clutters your prompt.{" "}
          <code>mnmd check</code> is the loud version. It validates the registered commands and
          prints one line per problem:
        </p>
        <div className="table-scroll">
          <table className="ref">
            <thead>
              <tr>
                <th>Check</th>
                <th>Severity</th>
                <th>Reports</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>
                  <code>root-contents</code>
                </td>
                <td>warning by default, configurable</td>
                <td>
                  Executables discovery found but didn't register, unreadable folders, a hook
                  script without its execute bit.
                </td>
              </tr>
              <tr>
                <td>
                  <code>path-spaces</code>
                </td>
                <td>always an error</td>
                <td>A registered path with a space in it, which completion would drop.</td>
              </tr>
              <tr>
                <td>
                  <code>declared-commands</code>
                </td>
                <td>always an error</td>
                <td>A declared path that isn't an executable file inside the root.</td>
              </tr>
              <tr>
                <td>
                  <code>config</code>
                </td>
                <td>always an error</td>
                <td>An invalid line, key, or value in the monom file or <code>config</code> hook.</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p>
          Warnings never change the exit code, so a fresh zero-config project passes. Any error
          exits 1. In the full example project:
        </p>
        <Code output>{CHECK}</Code>
        <p>
          To make <code>root-contents</code> an error, set <code>check.root-contents = error</code>{" "}
          in a declarative monom file (or print it from a hook script's <code>config</code> hook).
          Set <code>MONOM_CHECK_ROOT_CONTENTS=error</code> to make that your default everywhere; a
          project's own setting wins over it.
        </p>
        <p>
          <code>mnmd check</code> finds the root itself and needs no shell integration, so a CI job
          can build monom and run it from the project's checkout. <code>MONOM_ACTIVE=1</code>{" "}
          silences the install hint:
        </p>
        <Code>{`git clone --depth 1 https://github.com/adamgen/monom "$RUNNER_TEMP/monom"
make -C "$RUNNER_TEMP/monom" build
MONOM_ACTIVE=1 "$RUNNER_TEMP/monom/bin/mnmd" check`}</Code>

        <h2 id="config">6. The monom file</h2>
        <p>
          Optional. A file named <code>monom</code> at the root takes one of three shapes, decided
          by its first line:
        </p>
        <div className="table-scroll">
          <table className="ref">
            <thead>
              <tr>
                <th>Shape</th>
                <th>What it is</th>
                <th>monom does</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>Absent</td>
                <td>No file.</td>
                <td>Default discovery.</td>
              </tr>
              <tr>
                <td>Declarative</td>
                <td>
                  Any file whose first line isn't a shebang, including an empty one.
                </td>
                <td>Parses it, never runs it. Default discovery plus its declarations.</td>
              </tr>
              <tr>
                <td>Hook script</td>
                <td>An executable whose first line is a shebang.</td>
                <td>Runs it for hooks, never parses it.</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p>
          A declarative file has one entry per line. A line with <code>=</code> is a setting; any
          other line declares a command path relative to the root. Blank lines and{" "}
          <code>#</code> comments are ignored.
        </p>
        <Code title="acme/monom">{`# register an executable the gate would skip
tools/Format.sh
# make mnmd check fail on unregistered executables (default: warning)
check.root-contents = error
# hide more entries, like dot-files (one pattern per line)
discover.hide = infra/local`}</Code>
        <p>
          <code>discover.hide</code> patterns are{" "}
          <a href="https://pkg.go.dev/path#Match">path.Match</a> globs. One without a{" "}
          <code>/</code> matches a name at any depth; one with a <code>/</code> matches a path from
          the root, and <code>*</code> never crosses a <code>/</code>. A declared path is
          registered even when a pattern hides it.
        </p>

        <h2 id="hooks">7. Hook scripts</h2>
        <p>
          When declarations aren't enough, make <code>monom</code> an executable with a shebang.
          Every hook is optional: implement one to customize a default, leave it out and monom
          behaves as if it weren't there.
        </p>
        <div className="table-scroll">
          <table className="ref">
            <thead>
              <tr>
                <th>Hook</th>
                <th>Does</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>
                  <code>complete</code>
                </td>
                <td>
                  Prints every command path, one per line. Output replaces default discovery;
                  printing nothing falls back to it.
                </td>
              </tr>
              <tr>
                <td>
                  <code>run</code>
                </td>
                <td>Rewrites the typed words before they resolve: aliases, shortcuts, routing.</td>
              </tr>
              <tr>
                <td>
                  <code>config</code>
                </td>
                <td>
                  Prints settings in the declarative format (<code>check.root-contents</code>,{" "}
                  <code>discover.hide</code>).
                </td>
              </tr>
              <tr>
                <td>
                  <code>debug</code>
                </td>
                <td>Prints a writable file path for this project's debug log.</td>
              </tr>
            </tbody>
          </table>
        </div>
        <Code title="acme/monom">{`#!/usr/bin/env bash
case "$1" in
  run)
    shift
    # \`monom ship\` runs \`monom infra cloud deploy.sh\`
    if [ "$1" = "ship" ]; then echo "infra cloud deploy.sh"; fi
    ;;
  config)
    echo "discover.hide = tools"
    ;;
  debug)
    echo "$(cd "$(dirname "$0")" && pwd)/.monom-debug.log"
    ;;
esac`}</Code>
        <p>
          The <code>run</code> hook's exit code selects what happens:
        </p>
        <div className="table-scroll">
          <table className="ref">
            <thead>
              <tr>
                <th>Hook result</th>
                <th>monom does</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>exit 0, no output</td>
                <td>Uses your original words.</td>
              </tr>
              <tr>
                <td>exit 0, output</td>
                <td>Splits the output into words and resolves those instead.</td>
              </tr>
              <tr>
                <td>non-zero exit</td>
                <td>Shows the hook's stderr, returns its exit code, runs nothing.</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p className="callout">
          Prefer <code>if … fi</code> over <code>[ … ] &amp;&amp; echo</code> in hooks: when the
          test is false, <code>&amp;&amp;</code> leaves exit status 1, which monom reads as a
          failed hook.
        </p>
        <p>
          The repo's <a href={`${GITHUB_URL}/tree/main/fixtures`}>fixtures</a> are working
          examples of every shape: <code>demo-project</code> has a <code>complete</code> hook,{" "}
          <code>zero-config-project</code> has no monom file, and <code>zero-config-variants</code>{" "}
          covers declarations, severity, and <code>discover.hide</code>.
        </p>

        <h2 id="anywhere">8. Use it from anywhere</h2>
        <p>
          Pre-set <code>_MONOM_PROJECT_ROOT</code> to skip the search and pin one project, handy
          for a personal <code>~/scripts</code> toolbox. The pinned folder needs no monom file.
        </p>
        <Code>{`export _MONOM_PROJECT_ROOT="$HOME/scripts"`}</Code>

        <h2 id="reference">Reference</h2>
        <div className="table-scroll">
          <table className="ref">
            <thead>
              <tr>
                <th>Command</th>
                <th>What it does</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>
                  <code>monom &lt;words…&gt;</code>
                </td>
                <td>Resolves the words to a command under the project root and runs it.</td>
              </tr>
              <tr>
                <td>
                  <code>mnmd install</code>
                </td>
                <td>Adds the shell integration to your rc file. Idempotent.</td>
              </tr>
              <tr>
                <td>
                  <code>mnmd check</code>
                </td>
                <td>
                  Validates the registered commands and the monom file. Warnings exit 0, errors
                  exit 1.
                </td>
              </tr>
              <tr>
                <td>
                  <code>mnmd discover</code>
                </td>
                <td>Prints the commands default discovery registers, one path per line.</td>
              </tr>
              <tr>
                <td>
                  <code>mnmd root</code>
                </td>
                <td>Prints the active project root.</td>
              </tr>
              <tr>
                <td>
                  <code>mnmd pack &lt;words…&gt;</code>
                </td>
                <td>
                  Prints the absolute path the words resolve to. Exit 3 means “that's a command
                  group”.
                </td>
              </tr>
              <tr>
                <td>
                  <code>mnmd filter &lt;words…&gt;</code>
                </td>
                <td>Reads paths on stdin, prints the next completions. Always exits 0.</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h2 id="troubleshooting">Troubleshooting</h2>
        <dl className="faq">
          <dt>
            <code>hint: run 'mnmd install' to activate shell integration</code>
          </dt>
          <dd>
            <code>mnmd</code> ran in a shell that hasn't sourced monom. Run the installer, then open
            a new shell. Set <code>MONOM_ACTIVE=1</code> to silence it in scripts.
          </dd>
          <dt>
            <code>monom: no project root found</code>
          </dt>
          <dd>
            No <code>monom</code> file here or in any parent, no git repository, and no pinned{" "}
            <code>_MONOM_PROJECT_ROOT</code>. <code>cd</code> into a project, or{" "}
            <code>touch monom</code> where the root should be.
          </dd>
          <dt>A script is missing from Tab</dt>
          <dd>
            Run <code>mnmd check</code>. It names executables the gate rejected and why; add a
            shebang, rename it, or declare it in the monom file. Hidden, <code>_</code>-prefixed,
            and <code>discover.hide</code> matches are skipped on purpose.
          </dd>
          <dt>zsh: monom runs, but Tab does nothing</dt>
          <dd>
            The <code>source</code> line comes before <code>compinit</code> in{" "}
            <code>~/.zshrc</code>. Move it below.
          </dd>
          <dt>Something else</dt>
          <dd>
            Set <code>MONOM_DEBUG_LOG=/tmp/monom.log</code>, reproduce it, and read the log.
          </dd>
        </dl>

        <p className="docs-next">
          Going deeper? The <a href={`${GITHUB_URL}/blob/main/architecture.md`}>architecture</a>{" "}
          documents every contract, and the{" "}
          <a href={`${GITHUB_URL}/blob/main/constitution.md`}>constitution</a> explains why.
        </p>
      </article>
    </main>
  );
}
