import type { MetaFunction } from "@remix-run/node";
import { Link } from "@remix-run/react";
import { Code } from "~/components/Code";
import { CommandMap } from "~/components/CommandMap";
import { CopyCommand } from "~/components/CopyCommand";
import { FileTree } from "~/components/FileTree";
import { CHECK } from "~/lib/project";
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
      "Drop scripts into folders and they're commands: monom db migrate, with Tab completion. Any language, no config file, no registration, nothing generated.",
  },
];

const CHECK_WARN = `${CHECK}
$ echo $?
0`;

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
          <h1>
            Your <span className="is-group-text">file tree</span>
            <br />
            is your <span className="is-command-text">command tree</span>.
          </h1>
          <p className="lede">
            Drop scripts in folders. The path is the command, with Tab in bash and zsh. No config file, nothing
            generated.
          </p>
          <CommandMap />
          <div id="install">
            <CopyCommand command={INSTALL} />
          </div>
        </div>
      </section>

      <section className="section" id="how">
        <div className="wrap">
          <ol className="beats">
            <li>
              <span className="beat-n">01</span>
              <div>
                <h2>Nothing to register.</h2>
                <p>
                  Install once on the machine. A git repo adopts monom when you <code>cd</code> in. Outside git, an
                  empty file named <code>monom</code> marks the root.
                </p>
              </div>
            </li>
            <li>
              <span className="beat-n">02</span>
              <div>
                <h2>Spaces are the slashes.</h2>
                <p>
                  <code>infra/cloud/deploy.sh</code> is <code>monom infra cloud deploy</code>, from any directory in
                  the project. The extension drops. A folder is a group: it lists what is inside, and does not run.
                </p>
              </div>
            </li>
            <li>
              <span className="beat-n">03</span>
              <div>
                <h2>A front door, not a build.</h2>
                <p>
                  There is no task graph, no cache, and no flags yet. If you need those, use just or Task. Delete
                  monom later and <code>./infra/cloud/deploy.sh</code> still runs.
                </p>
              </div>
            </li>
          </ol>
        </div>
      </section>

      <section className="section" id="discovery">
        <div className="wrap">
          <div className="stage-row">
            <div>
              <SectionHead kicker="the gate" title="monom finds every executable file." />
              <p className="section-lede">
                It looks at the executable files in your project and registers each one with a shebang or a
                command-style name, or that you declared. Then it maps the file tree to a ready-to-use CLI, with Tab
                completion in bash and zsh.
              </p>
              <p>
                Any language works: Python, Node, bash, Ruby, a compiled Go binary. The path is the command, minus
                the file's extension, so <code>db/migrate.py</code> is <code>monom db migrate</code>. A command-style
                name matches <code>^[a-z0-9][a-z0-9_-]*$</code> (lowercase, no dots), so a file with an extension
                needs a shebang, and a compiled binary like <code>tools/lint</code> gets in by name.
              </p>
              <p className="muted">
                Hidden and <code>_</code>-prefixed entries, <code>node_modules</code>, <code>vendor</code>,{" "}
                <code>venv</code>, <code>target</code>, <code>dist</code>, <code>__pycache__</code> and nested monom
                projects are never scanned. Add <code>discover.hide</code> patterns to skip more.
              </p>
            </div>
            <div className="tree-card">
              <FileTree annotate />
            </div>
          </div>

          <div className="check-row">
            <div>
              <h3 className="small-head">Then keep it honest in CI</h3>
              <p>
                <code>mnmd check</code> flags executables discovery found but didn't register, as a{" "}
                <strong>warning</strong> by default, so a fresh project passes. One line in a <code>monom</code> file
                makes that an error. The <Link to="/docs#check">quick start</Link> has the line.
              </p>
            </div>
            <Code output>{CHECK_WARN}</Code>
          </div>
        </div>
      </section>

      <section className="section" id="compare">
        <div className="wrap">
          <SectionHead kicker="compare" title="Honest about what it isn't." />
          <p className="section-lede">
            monom is closest to Basecamp's <a href="https://github.com/basecamp/sub">sub</a>, a folder of scripts
            with completion, grown nested and given a Go engine. It is <em>not</em> a build tool: if you need a task
            graph or caching, reach for just, Task, or mise. And today every word you type is part of the command
            path: commands don't take arguments yet.
          </p>
          <div className="closest">
            <article>
              <h3>monom</h3>
              <p>Nested folders, any language, no generate step. Your code can replace the command list.</p>
            </article>
            <article>
              <h3>sub</h3>
              <p>Basecamp's flat folder of scripts. The same idea, one level deep.</p>
            </article>
            <article>
              <h3>just</h3>
              <p>A justfile. The right tool when the job is a graph of tasks.</p>
            </article>
          </div>
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
                        <span
                          className={`mark mark-${mark}`}
                          aria-label={mark === "yes" ? "yes" : mark === "no" ? "no" : "partly"}
                        >
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
