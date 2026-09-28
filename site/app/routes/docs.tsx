import type { MetaFunction } from "@remix-run/node";
import { Code } from "~/components/Code";
import { CopyCommand } from "~/components/CopyCommand";
import { GITHUB_URL, INSTALL } from "~/lib/site";

export const meta: MetaFunction = () => [
  { title: "Quick start — monom" },
  {
    name: "description",
    content: "Install monom, turn a folder of scripts into a tab-completable CLI, and check it in CI.",
  },
];

const TOC = [
  ["install", "Install"],
  ["project", "Make a project"],
  ["config", "Add the config"],
  ["try", "Press Tab"],
  ["check", "Check it in CI"],
  ["hooks", "Hooks"],
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
          From nothing to a tab-completing <code>monom infra cloud deploy</code> in five minutes.
          You'll need <code>git</code>, <code>make</code>, Go 1.24+, and bash or zsh.
        </p>

        <h2 id="install">1. Install</h2>
        <p>
          Clone the repo, build the <code>mnmd</code> engine, and let it add one{" "}
          <code>source</code> line to your shell's rc file.
        </p>
        <CopyCommand command={INSTALL} />
        <p>
          <code>mnmd install</code> picks <code>~/.zshrc</code> for zsh, or{" "}
          <code>~/.bash_profile</code> (falling back to <code>~/.bashrc</code>) for bash. It's
          idempotent — a second run prints <code>already installed</code>. Open a new shell, or
          source the file it names:
        </p>
        <Code output>{`$ ~/.monom/bin/mnmd install
added to /Users/you/.zshrc
restart your shell or run: source /Users/you/.zshrc`}</Code>

        <h2 id="project">2. Make a project</h2>
        <p>
          A monom project is a folder. Each command is an executable file; each subfolder is a
          command group. Use any language — the shebang decides.
        </p>
        <Code>{`$ mkdir -p acme/infra/cloud acme/db && cd acme
$ printf '#!/usr/bin/env bash\\necho "deploying to cloud..."\\n' > infra/cloud/deploy
$ printf '#!/usr/bin/env python3\\nprint("running db migrations...")\\n' > db/migrate
$ chmod +x infra/cloud/deploy db/migrate`}</Code>
        <p>
          Leave off file extensions: the file's path <em>is</em> the command, so{" "}
          <code>deploy.sh</code> would be typed as <code>monom infra cloud deploy.sh</code>.
        </p>

        <h2 id="config">3. Add the config</h2>
        <p>
          Create an executable named <code>monom</code> at the project root. Its one required job:
          when called as <code>monom complete</code>, print every command path, slash-delimited,
          one per line.
        </p>
        <Code title="acme/monom">{`#!/usr/bin/env bash
case "$1" in
  complete)
    root="$(cd "$(dirname "$0")" && pwd)"
    find "$root" -type f -perm -u+x ! -name monom -print | sed "s#^$root/##" | sort
    ;;
esac`}</Code>
        <Code output>{`$ chmod +x monom
$ ./monom complete
db/migrate
infra/cloud/deploy`}</Code>
        <p>
          The directory containing this file is the <strong>project root</strong>. monom finds it
          by walking up from wherever you are, so commands work from any subfolder.
        </p>

        <h2 id="try">4. Press Tab</h2>
        <Code output>{`$ monom <Tab>
db     infra
$ monom in<Tab>         # completes to: monom infra
$ monom infra cloud deploy
deploying to cloud...
$ monom infra
monom: 'infra' is a command group
available: cloud`}</Code>
        <p>
          Typing a folder instead of a script lists what's inside — the same list Tab would show.
          Nothing runs.
        </p>

        <h2 id="check">5. Check it in CI</h2>
        <p>
          Completion stays silent when something's wrong so it never clutters your prompt.{" "}
          <code>mnmd check</code> is the loud version: it runs your <code>complete</code> and exits
          non-zero if any path would be undiscoverable (for instance, a segment containing a
          space).
        </p>
        <Code output>{`$ mnmd check
✔ 2 commands OK`}</Code>

        <h2 id="hooks">6. Hooks</h2>
        <p>
          Hooks are optional subcommands of your config. Implement one to customize a default;
          leave it out and monom behaves as if it weren't there. There's nothing to register.
        </p>

        <h3>
          <code>run</code> — rewrite arguments before they resolve
        </h3>
        <p>
          monom calls <code>./monom run &lt;args…&gt;</code> before resolving a command. Print
          replacement arguments to reroute; print nothing to keep the originals.
        </p>
        <Code title="acme/monom">{`  run)
    shift
    # \`monom ship\` runs \`monom infra cloud deploy\`
    if [ "$1" = "ship" ]; then echo "infra cloud deploy"; fi
    ;;`}</Code>
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
                <td>Uses your original arguments.</td>
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

        <h3>
          <code>debug</code> — a project-local debug log
        </h3>
        <p>
          Print one writable file path and monom (shell and engine alike) appends timestamped
          debug lines there for this project. Globally, set <code>MONOM_DEBUG_LOG</code> instead;
          the hook wins when both are set.
        </p>
        <Code>{`  debug)
    echo "$(cd "$(dirname "$0")" && pwd)/.monom-debug.log"
    ;;`}</Code>

        <h2 id="anywhere">7. Use it from anywhere</h2>
        <p>
          Pre-set <code>_MONOM_PROJECT_ROOT</code> to skip the upward search and pin one project —
          handy for a personal <code>~/scripts</code> toolbox.
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
                <td>Resolves words to a script under the project root and runs it.</td>
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
                <td>Validates every path from <code>complete</code>. Non-zero on problems.</td>
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
            No executable <code>monom</code> file in this directory or any parent. Check it exists
            and has <code>chmod +x</code>.
          </dd>
          <dt>A command is missing from Tab</dt>
          <dd>
            Run <code>mnmd check</code>. Paths with spaces in any segment are skipped by completion.
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
