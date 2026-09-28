import type { MetaFunction } from "@remix-run/node";
import { Link } from "@remix-run/react";
import { Code } from "~/components/Code";
import { CopyCommand } from "~/components/CopyCommand";
import { FileTree } from "~/components/FileTree";
import { TabDemo } from "~/components/TabDemo";
import { GITHUB_URL, INSTALL } from "~/lib/site";

export const meta: MetaFunction = () => [
  { title: "monom — your file tree is your command tree" },
  {
    name: "description",
    content:
      "monom turns the executable scripts already in your repo into a tab-completable CLI. No config file, any language, bash and zsh completion.",
  },
  { property: "og:title", content: "monom — your file tree is your command tree" },
  {
    property: "og:description",
    content:
      "Drop scripts into folders and they're commands: monom db migrate, with Tab completion. No config file, no registration, nothing generated.",
  },
];

// Every snippet below was run against the real mnmd/monom (bash and zsh) in a
// project laid out like ACME_PATHS, plus tools/Format.sh without a shebang.
const ACME_PATHS = ["db/migrate", "db/seed", "infra/cloud/deploy", "release"];

const DISCOVER = `$ mnmd discover
db/migrate
db/seed
infra/cloud/deploy
release`;

const RUN = `$ monom infra cloud <Tab>
$ monom infra cloud deploy
deploying to cloud...
$ monom db migrate
running db migrations...`;

const CHECK_WARN = `$ mnmd check
warning: [root-contents] executable not registered: tools/Format.sh (no shebang, and the name does not match ^[a-z0-9][a-z0-9_-]*$); add a shebang, rename it, or declare it in the monom config file
✔ 4 commands OK (default discovery), 1 warning(s)
$ echo $?
0`;

const CHECK_ERROR = `$ echo 'check.root-contents = error' > monom
$ mnmd check
error: [root-contents] executable not registered: tools/Format.sh (no shebang, and the name does not match ^[a-z0-9][a-z0-9_-]*$); add a shebang, rename it, or declare it in the monom config file
mnmd check: 1 error(s), 0 warning(s)
$ echo $?
1`;

const DECLARATIVE = `# register an executable the gate would skip
tools/Format.sh
# make mnmd check fail on unregistered executables
check.root-contents = error
# treat more entries as hidden (repeatable)
discover.hide = *.TXT
discover.hide = tools/wip-*`;

const HOOK_SCRIPT = `#!/usr/bin/env bash
# every hook is optional: without complete,
# default discovery lists the commands
case "$1" in
  run)
    shift
    # monom ship → infra cloud deploy
    if [ "$1" = ship ]; then echo infra cloud deploy; fi
    ;;
  config)
    echo "discover.hide = tools"
    ;;
esac`;

const GROUP = `$ monom infra
monom: 'infra' is a command group
available: cloud, local`;

// Outcomes of the discovery gate for fixtures/zero-config-project in the repo.
const GATE: { path: string; why: string; outcome: "yes" | "warn" | "skip"; label: string }[] = [
  { path: "deploy", why: "starts with #!/bin/sh", outcome: "yes", label: "monom deploy" },
  { path: "scripts/setup.sh", why: "starts with #!/bin/sh", outcome: "yes", label: "monom scripts setup.sh" },
  { path: "build/compile", why: "no shebang, but a command-shaped name (compiled binaries pass this way)", outcome: "yes", label: "monom build compile" },
  { path: "notes.TXT", why: "executable, no shebang, not a command-shaped name", outcome: "warn", label: "mnmd check warns" },
  { path: "docs/guide", why: "not executable", outcome: "skip", label: "ignored" },
  { path: "_lib/helper, .hidden/tool", why: "_-prefixed and hidden entries are private", outcome: "skip", label: "never scanned" },
  { path: "node_modules/pkg/cli", why: "node_modules, vendor, venv, target, dist, __pycache__", outcome: "skip", label: "never scanned" },
];

const LANGS: [string, string][] = [
  ["db migrate", "python3"],
  ["db seed", "node"],
  ["infra cloud deploy", "bash"],
  ["release", "bash"],
];

type Mark = "yes" | "no" | "part";
const COMPARE: { row: string; cells: [Mark, string?][] }[] = [
  {
    row: "Each command is its own executable file",
    cells: [["yes"], ["yes"], ["no", "recipes in a justfile"], ["no", "tasks in Taskfile.yml"], ["yes", "file tasks"], ["no", "compiled into one script"]],
  },
  {
    row: "Nested command groups from folders",
    cells: [["yes", "any depth"], ["no", "flat"], ["part", "modules"], ["part", "includes"], ["yes", "as a:b"], ["yes", "declared in YAML"]],
  },
  {
    row: "Commands in any language",
    cells: [["yes"], ["yes"], ["yes", "shebang recipes"], ["part", "via shell"], ["yes"], ["no", "bash"]],
  },
  {
    row: "No generate or build step",
    cells: [["yes"], ["yes"], ["yes"], ["yes"], ["yes"], ["no", "bashly generate"]],
  },
  {
    row: "Your code can decide the command list",
    cells: [["yes", "optional complete hook"], ["no"], ["no"], ["no"], ["no"], ["no"]],
  },
  {
    row: "Arguments, flag parsing and generated help",
    cells: [["no", "on the backlog"], ["part", "help from comments"], ["yes"], ["part"], ["yes", "usage spec"], ["yes"]],
  },
  {
    row: "Dependencies between tasks",
    cells: [["no", "not a build tool"], ["no"], ["yes"], ["yes", "+ caching"], ["yes"], ["no"]],
  },
];
const TOOLS = ["monom", "sub", "just", "Task", "mise", "bashly"];

export default function Index() {
  return (
    <main>
      <section className="hero">
        <div className="wrap">
          <p className="eyebrow">
            <span className="pill">early development</span> zero config · bash &amp; zsh · any
            language · Go engine
          </p>
          <h1>
            Your <span className="mono is-group-text">file tree</span>
            <br />
            is your <span className="mono is-command-text">command tree</span>.
          </h1>
          <p className="lede">
            Drop executable scripts into your repo's folders and they're already a CLI:{" "}
            <code>monom db migrate</code>, with Tab completion in bash and zsh. No config file, no
            registration, nothing generated.
          </p>
          <div className="hero-actions">
            <Link to="/docs" className="button is-primary">
              Quick start →
            </Link>
            <a href={GITHUB_URL} className="button">
              View source
            </a>
          </div>
          <CopyCommand command={INSTALL} label="install" />
        </div>
        <div className="wrap wide">
          <TabDemo />
          <p className="demo-caption">
            A live port of <code>mnmd filter</code> running against the repo's demo project. Folders
            complete as <span className="is-group-text">groups</span>, scripts as{" "}
            <span className="is-command-text">commands</span>.
          </p>
        </div>
      </section>

      <section className="section" id="adopt">
        <div className="wrap">
          <SectionHead kicker="adopting monom" title="The boilerplate is a folder of scripts." />
          <p className="section-lede">
            Install monom once per machine. After that, a project adopts it by doing nothing: the
            scripts it already has become commands the moment you <code>cd</code> into it.
          </p>
          <ol className="steps">
            <li className="step">
              <span className="step-n">01</span>
              <h3>Drop scripts in folders</h3>
              <p>
                Any executable with a shebang, in any folder of a git repo. Folders become command
                groups; scripts become commands. No file to create first.
              </p>
              <div className="step-visual tree-card">
                <FileTree paths={ACME_PATHS} config={null} />
              </div>
            </li>
            <li className="step">
              <span className="step-n">02</span>
              <h3>monom finds them</h3>
              <p>
                The nearest git root is the project. Default discovery scans it and registers what
                passes the gate. <code>mnmd discover</code> shows the result.
              </p>
              <div className="step-visual">
                <Code output>{DISCOVER}</Code>
              </div>
            </li>
            <li className="step">
              <span className="step-n">03</span>
              <h3>Press Tab, run it</h3>
              <p>
                Spaces become slashes. <code>monom infra cloud deploy</code> runs{" "}
                <code>infra/cloud/deploy</code>, from any folder inside the project.
              </p>
              <div className="step-visual">
                <Code output>{RUN}</Code>
              </div>
            </li>
          </ol>
          <div className="config-row">
            <div>
              <h3 className="small-head">Then keep it honest in CI</h3>
              <p>
                <code>mnmd check</code> is the doctor. It flags executables discovery found but
                didn't register, as a <strong>warning</strong> by default, so a fresh project
                passes. When you're ready to enforce it, one line in a <code>monom</code> file
                makes it an error.
              </p>
              <p className="muted">
                Outside git? <code>touch monom</code> marks a root. An empty file is a valid config.
              </p>
            </div>
            <div className="stack">
              <Code output>{CHECK_WARN}</Code>
              <Code output>{CHECK_ERROR}</Code>
            </div>
          </div>
        </div>
      </section>

      <section className="section" id="discovery">
        <div className="wrap">
          <SectionHead kicker="zero-config discovery" title="What becomes a command, and what doesn't." />
          <p className="section-lede">
            Discovery scans broadly, then a gate decides. An executable is registered if it starts
            with a shebang, if its name looks like a command (<code>^[a-z0-9][a-z0-9_-]*$</code>:
            lowercase, no extension), or if the <code>monom</code> file declares it. This is the
            repo's <code>fixtures/zero-config-project</code>:
          </p>
          <div className="table-scroll">
            <table className="ref gate">
              <thead>
                <tr>
                  <th>File</th>
                  <th>Why</th>
                  <th>Result</th>
                </tr>
              </thead>
              <tbody>
                {GATE.map((g) => (
                  <tr key={g.path}>
                    <td>{g.path}</td>
                    <td>{g.why}</td>
                    <td>
                      <span className={`outcome is-${g.outcome}`}>{g.label}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="footnote">
            Nested monom projects (a subfolder with its own <code>monom</code> file) and paths with
            spaces are never registered either. To hide more, add <code>discover.hide</code>{" "}
            patterns.
          </p>
        </div>
      </section>

      <section className="section" id="config">
        <div className="wrap">
          <SectionHead kicker="the monom file" title="Optional. Add one only when you need it." />
          <p className="section-lede">
            A file named <code>monom</code> at the root marks the project and customizes it. It
            comes in three shapes, decided by its first line.
          </p>
          <div className="features">
            <article className="feature">
              <h3>Absent</h3>
              <p>
                Zero config. The git root is the project and default discovery supplies the command
                tree. Most projects can stay here.
              </p>
              <Code output>{`$ cd ~/code/acme/infra/cloud
$ mnmd root
/home/you/code/acme`}</Code>
            </article>
            <article className="feature">
              <h3>Declarative</h3>
              <p>
                Any file without a shebang, including an empty one. monom parses it and never runs
                it: declare paths the gate skips, hide entries, set <code>mnmd check</code>'s
                severity.
              </p>
              <Code title="monom">{DECLARATIVE}</Code>
            </article>
            <article className="feature">
              <h3>Hook script</h3>
              <p>
                An executable with a shebang. monom runs it for hooks, all optional:{" "}
                <code>complete</code>, <code>run</code>, <code>config</code>, <code>debug</code>.
                Leave one out and the default applies.
              </p>
              <Code title="monom">{HOOK_SCRIPT}</Code>
            </article>
          </div>
          <div className="root-order">
            <strong>How the root is found</strong>, first match wins:
            <ol>
              <li>
                <code>$_MONOM_PROJECT_ROOT</code>, the pin an alias or your rc file sets;
              </li>
              <li>
                the nearest directory upward containing a file named <code>monom</code>;
              </li>
              <li>the nearest git root.</li>
            </ol>
            The current directory alone is never a root.
          </div>
        </div>
      </section>

      <section className="section" id="features">
        <div className="wrap">
          <SectionHead kicker="why monom" title="A CLI framework that stays out of your scripts." />
          <div className="features">
            <article className="feature">
              <h3>Any language with a shebang</h3>
              <p>
                monom resolves a path and <code>exec</code>s it. Your script never imports, sources,
                or links anything.
              </p>
              <ul className="lang-list">
                {LANGS.map(([cmd, lang]) => (
                  <li key={cmd}>
                    <span className="is-command-text">{cmd}</span>
                    <span className="shebang">#!/usr/bin/env {lang}</span>
                  </li>
                ))}
              </ul>
            </article>
            <article className="feature">
              <h3>Go owns logic, shell owns surface</h3>
              <p>
                Discovery, filtering and resolution live in <code>mnmd</code>, a compiled Go binary.
                The shell layer is only what Go can't do: define <code>monom</code>, register
                completion, <code>exec</code>.
              </p>
              <Code output>{`$ type -t monom mnmd
function
function`}</Code>
            </article>
            <article className="feature">
              <h3>Tab tested with real key presses</h3>
              <p>
                Completion is covered by YAML cases that start interactive bash and zsh in a
                pseudo-terminal, press Tab, and check the edit line and the listing on screen.
              </p>
              <Code title="tests/cases/monom_keys.yaml">{`- name: tab completes a unique prefix and adds a space
  input: monom in
  action: keys
  tabs: 1
  line: "monom infra "`}</Code>
            </article>
            <article className="feature">
              <h3>Completion never breaks mid-typing</h3>
              <p>
                <code>mnmd filter</code> always exits 0. A broken path yields no suggestion, not an
                error splashed over your prompt. Diagnostics live in <code>mnmd check</code>.
              </p>
              <Code output>{`$ echo 'bad path/x' | mnmd filter bad
$ echo $?
0`}</Code>
            </article>
            <article className="feature">
              <h3>Your code can own the list</h3>
              <p>
                A <code>complete</code> hook that prints paths replaces default discovery entirely:
                read <code>git ls-files</code>, parse a manifest, anything that prints lines.
              </p>
              <Code>{`  complete)
    cd "$(dirname "$0")" || exit
    # tracked executables, minus this file
    git ls-files -s | awk '$1 == 100755 { print $4 }' | grep -vx monom
    ;;`}</Code>
            </article>
            <article className="feature">
              <h3>No lock-in</h3>
              <p>
                Your commands are plain executables in plain folders. Delete monom tomorrow and{" "}
                <code>./infra/cloud/deploy</code> still runs.
              </p>
              <Code output>{`$ ./infra/cloud/deploy
deploying to cloud...`}</Code>
            </article>
          </div>
        </div>
      </section>

      <section className="section split" id="groups">
        <div className="wrap split-inner">
          <div>
            <SectionHead kicker="nouns and verbs" title="Folders are nouns. Scripts are verbs." />
            <p>
              Type a folder and monom tells you what's inside instead of failing, using the same
              pipeline as Tab, so the listing always matches what completion would offer.
            </p>
            <p className="muted">
              Want <code>monom infra</code> to do something? Map it in the <code>run</code> hook. A
              group is never runnable by accident.
            </p>
          </div>
          <Code output>{GROUP}</Code>
        </div>
      </section>

      <section className="section" id="who">
        <div className="wrap">
          <SectionHead kicker="who it's for" title="One framework, two kinds of CLI." />
          <div className="audiences">
            <article className="audience">
              <h3>The repo's front door</h3>
              <p>
                Every repo grows a pile of scripts nobody can navigate. With monom installed, a
                teammate's first <code>monom</code> <kbd>Tab</kbd> in the repo shows every workflow
                the team has, with nothing committed for monom's sake.
              </p>
              <Code output>{`$ monom <Tab>
db       infra    release
$ monom db <Tab>
migrate  seed`}</Code>
            </article>
            <article className="audience">
              <h3>Your own toolbox</h3>
              <p>
                Pin monom to a folder of personal scripts and they're tab-completable from any
                directory, organized by topic instead of prefixed <code>my-thing-do-x.sh</code>.
              </p>
              <Code>{`# ~/.zshrc, after the source line mnmd install added
export _MONOM_PROJECT_ROOT="$HOME/scripts"`}</Code>
            </article>
          </div>
        </div>
      </section>

      <section className="section" id="compare">
        <div className="wrap">
          <SectionHead kicker="compare" title="Honest about what it isn't." />
          <p className="section-lede">
            monom is closest to Basecamp's <a href="https://github.com/basecamp/sub">sub</a>, a
            folder of scripts with completion, grown nested and given a Go engine. It is{" "}
            <em>not</em> a build tool: if you need a task graph or caching, reach for just, Task, or
            mise. And today every word you type is part of the command path: commands don't take
            arguments yet.
          </p>
          <div className="table-scroll">
            <table className="compare">
              <thead>
                <tr>
                  <th scope="col" />
                  {TOOLS.map((t) => (
                    <th scope="col" key={t} className={t === "monom" ? "is-us" : undefined}>
                      {t}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {COMPARE.map((r) => (
                  <tr key={r.row}>
                    <th scope="row">{r.row}</th>
                    {r.cells.map(([mark, note], i) => (
                      <td key={TOOLS[i]} className={i === 0 ? "is-us" : undefined}>
                        <span className={`mark mark-${mark}`} aria-label={mark === "yes" ? "yes" : mark === "no" ? "no" : "partly"}>
                          {mark === "yes" ? "●" : mark === "no" ? "○" : "◐"}
                        </span>
                        {note && <span className="cell-note">{note}</span>}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="footnote">
            Compiled from each project's documentation, September 2026. Spotted something wrong?{" "}
            <a href={`${GITHUB_URL}/issues`}>Open an issue</a>.
          </p>
        </div>
      </section>

      <section className="section" id="principles">
        <div className="wrap">
          <SectionHead kicker="principles" title="Built on a written constitution." />
          <div className="principles">
            <Principle title="Go owns logic, shell owns surface">
              Shell code exists only where Go is technically impossible: sourcing into your
              session, registering completion, <code>exec</code>.
            </Principle>
            <Principle title="Minimize subprocess roundtrips">
              Every process boundary on the Tab path has to justify itself. Fewer hops, less
              latency, less to debug.
            </Principle>
            <Principle title="Errors carry their own exit code">
              One central registry. A command group exits 3, always, and nothing else does.
            </Principle>
            <Principle title="The config interface is a contract">
              monom requires nothing from your <code>monom</code> file, not even that it exists.
              Changing that takes a constitutional amendment.
            </Principle>
          </div>
          <a className="text-link" href={`${GITHUB_URL}/blob/main/constitution.md`}>
            Read the constitution →
          </a>
        </div>
      </section>

      <section className="section cta">
        <div className="wrap cta-inner">
          <h2>Give your scripts a front door.</h2>
          <p className="lede">Clone, build, one rc line. Then open any repo and press Tab.</p>
          <CopyCommand command={INSTALL} />
          <div className="hero-actions">
            <Link to="/docs" className="button is-primary">
              Read the quick start →
            </Link>
            <a href={GITHUB_URL} className="button">
              View on GitHub
            </a>
          </div>
          <p className="status-note">
            monom is in early development. Zero-config discovery, <code>mnmd check</code> and
            bash/zsh completion work today. Commands don't take arguments yet; flag parsing and
            scaffolding (<code>mnmd init</code>, <code>mnmd new command</code>) are on the{" "}
            <a href={`${GITHUB_URL}/blob/main/BACKLOG.md`}>backlog</a>.
          </p>
        </div>
      </section>
    </main>
  );
}

function SectionHead({ kicker, title }: { kicker: string; title: string }) {
  return (
    <header className="section-head">
      <p className="kicker">{kicker}</p>
      <h2>{title}</h2>
    </header>
  );
}

function Principle({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <article className="principle">
      <h3>{title}</h3>
      <p>{children}</p>
    </article>
  );
}
