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
      "monom turns a folder of scripts into a tab-completable CLI. Any language, bash and zsh completion, nothing to register or generate.",
  },
  { property: "og:title", content: "monom — your file tree is your command tree" },
  {
    property: "og:description",
    content: "Put scripts in folders. Add one executable that lists them. Get a CLI with tab completion.",
  },
];

const CONFIG = `#!/usr/bin/env bash
# acme/monom — the only file monom needs from you.
case "$1" in
  complete)
    root="$(cd "$(dirname "$0")" && pwd)"
    find "$root" -type f -perm -u+x ! -name monom -print | sed "s#^$root/##" | sort
    ;;
esac`;

const COMPLETE_OUTPUT = `$ ./monom complete
db/migrate
db/seed
infra/cloud/deploy
infra/cloud/teardown
infra/local/start
infra/local/stop
release`;

const RUN_HOOK = `  run)
    shift
    # monom ship → infra cloud deploy
    if [ "$1" = ship ]; then echo infra cloud deploy; fi
    ;;`;

const GROUP = `$ monom infra
monom: 'infra' is a command group
available: cloud, local`;

const CHECK = `$ mnmd check
✔ 7 commands OK`;

const LANGS = [
  "db/migrate",
  "db/seed",
  "infra/cloud/deploy",
  "report",
];

const SHEBANGS: Record<string, string> = {
  "db/migrate": "python3",
  "db/seed": "node",
  "infra/cloud/deploy": "bash",
  report: "ruby",
};

type Mark = "yes" | "no" | "part";
const COMPARE: { row: string; note?: string; cells: [Mark, string?][] }[] = [
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
    row: "Your code decides the command list",
    cells: [["yes", "complete"], ["no"], ["no"], ["no"], ["no"], ["no"]],
  },
  {
    row: "Flag parsing and generated help",
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
            <span className="pill">early development</span> bash &amp; zsh · any language · Go engine
          </p>
          <h1>
            Your <span className="mono is-group-text">file tree</span>
            <br />
            is your <span className="mono is-command-text">command tree</span>.
          </h1>
          <p className="lede">
            Put scripts in folders. Add one executable that lists them. monom gives you a CLI with
            instant tab completion — no registration, no DSL, nothing generated.
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

      <section className="section" id="how">
        <div className="wrap">
          <SectionHead kicker="how it works" title="Three steps. The first one you've already done." />
          <ol className="steps">
            <li className="step">
              <span className="step-n">01</span>
              <h3>Organize</h3>
              <p>
                Executable scripts in folders. Folders become command groups; scripts become
                commands. That's the whole schema.
              </p>
              <div className="step-visual tree-card">
                <FileTree
                  paths={["db/migrate", "db/seed", "infra/cloud/deploy", "infra/cloud/teardown", "release"]}
                />
              </div>
            </li>
            <li className="step">
              <span className="step-n">02</span>
              <h3>List</h3>
              <p>
                An executable called <code>monom</code> at the root. Its <code>complete</code>{" "}
                subcommand prints every command path, one per line. Here it's <code>find</code>; it
                could be anything.
              </p>
              <div className="step-visual">
                <Code output>{COMPLETE_OUTPUT}</Code>
              </div>
            </li>
            <li className="step">
              <span className="step-n">03</span>
              <h3>Press Tab</h3>
              <p>
                Spaces become slashes. <code>monom infra cloud deploy</code> resolves to{" "}
                <code>./infra/cloud/deploy</code> and runs it, from anywhere in the project.
              </p>
              <div className="step-visual">
                <Code output>{`$ monom infra cloud <Tab>
deploy    teardown
$ monom infra cloud deploy
deploying to cloud...`}</Code>
              </div>
            </li>
          </ol>
          <div className="config-row">
            <div>
              <h3 className="small-head">The entire config, for real</h3>
              <p>
                This is the demo project's config, verbatim. monom requires exactly one subcommand
                from it — <code>complete</code> — and that contract is protected by the project's
                constitution, so a config that works today keeps working.
              </p>
              <p className="muted">
                Shell, Python, Go, a compiled binary: monom doesn't care how it's written, only that
                it prints paths.
              </p>
            </div>
            <Code title="acme/monom">{CONFIG}</Code>
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
                {LANGS.map((p) => (
                  <li key={p}>
                    <span className="is-command-text">{p.replaceAll("/", " ")}</span>
                    <span className="shebang">#!/usr/bin/env {SHEBANGS[p]}</span>
                  </li>
                ))}
              </ul>
            </article>
            <article className="feature">
              <h3>Fast enough to disappear</h3>
              <p>
                Filtering and resolution live in <code>mnmd</code>, a compiled Go binary. The shell
                layer is only what Go can't do: register completion and <code>exec</code>.
              </p>
              <div className="metric">
                <div>
                  <strong>2.6 ms</strong>
                  <span>mnmd filter</span>
                </div>
                <div>
                  <strong>12 ms</strong>
                  <span>full Tab round-trip</span>
                </div>
              </div>
              <p className="footnote">Demo project, Linux, mean of 200 runs.</p>
            </article>
            <article className="feature">
              <h3>Discovery is yours</h3>
              <p>
                <code>complete</code> can walk the disk, read <code>git ls-files</code>, parse a
                manifest, or hide anything under <code>_internal/</code>. monom only reads the
                output.
              </p>
              <Code>{`  complete)
    cd "$(dirname "$0")" || exit
    # tracked executables, minus this file and anything private
    git ls-files -s | awk '$1 == 100755 { print $4 }' | grep -v -e '^monom$' -e '^_internal/'
    ;;`}</Code>
            </article>
            <article className="feature">
              <h3>Hooks without ceremony</h3>
              <p>
                Add a <code>run</code> subcommand to remap arguments — aliases, shortcuts, routing.
                Leave it out and nothing changes. No registration, no opt-out flag.
              </p>
              <Code>{RUN_HOOK}</Code>
            </article>
            <article className="feature">
              <h3>Completion never breaks mid-typing</h3>
              <p>
                <code>mnmd filter</code> always exits 0. A broken path yields no suggestion, not an
                error splashed over your prompt. Diagnostics live in <code>mnmd check</code>.
              </p>
              <Code output>{CHECK}</Code>
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
              Type a folder and monom tells you what's inside instead of failing — using the same
              pipeline as Tab, so the listing always matches what completion would offer.
            </p>
            <p className="muted">
              Want <code>monom infra</code> to do something? Map it in the <code>run</code> hook.
              A group is never runnable by accident.
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
              <h3>The monorepo's front door</h3>
              <p>
                Every repo grows a <code>scripts/</code> folder nobody can navigate. Drop a{" "}
                <code>monom</code> file at the root and a new teammate's first <code>monom</code>{" "}
                <kbd>Tab</kbd> shows them every workflow the team has.
              </p>
              <Code output>{`$ monom <Tab>
db    infra    release    ui
$ monom ui <Tab>
build    lint    storybook`}</Code>
            </article>
            <article className="audience">
              <h3>Your own toolbox</h3>
              <p>
                Point monom at a folder of personal scripts and they're tab-completable from any
                directory — organized by topic instead of prefixed <code>my-thing-do-x.sh</code>.
              </p>
              <Code>{`# ~/.zshrc
export _MONOM_PROJECT_ROOT="$HOME/scripts"
source "$HOME/.monom/src/monom"`}</Code>
            </article>
          </div>
        </div>
      </section>

      <section className="section" id="compare">
        <div className="wrap">
          <SectionHead kicker="compare" title="Honest about what it isn't." />
          <p className="section-lede">
            monom is closest to Basecamp's <a href="https://github.com/basecamp/sub">sub</a> — a
            folder of scripts with completion — grown nested and given a fast engine. It is{" "}
            <em>not</em> a build tool: if you need a task graph or caching, reach for just, Task, or
            mise.
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
              Changing what monom requires from your <code>monom</code> file takes a constitutional
              amendment.
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
          <p className="lede">Five minutes from clone to your first Tab.</p>
          <CopyCommand command={INSTALL} />
          <div className="hero-actions">
            <Link to="/docs" className="button is-primary">
              Read the quick start →
            </Link>
            <a href={GITHUB_URL} className="button">
              Star on GitHub
            </a>
          </div>
          <p className="status-note">
            monom is in early development. The engine and bash/zsh completion work today; flag
            parsing and project scaffolding are on the{" "}
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
