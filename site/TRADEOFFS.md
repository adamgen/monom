# site/ tradeoffs

## Remix v2, not Remix 3 or React Router framework mode

**Chosen:** Remix v2 (`@remix-run/*` 2.17, the npm `latest` tag) on Vite, with every v3 future flag enabled.

**Rejected:**
- *Remix 3* (`remix@next`, 3.0.0-rc at the time of writing) — a ground-up rewrite that drops React.
- *React Router v7+/v8 framework mode* — the official successor to Remix v2, with the same file-route and loader model under a different package name.

**Why:** The site was asked to be Remix-based. Remix 3 was still a release candidate with a new component model and little documentation, so building a public page on it risked churn through each RC. React Router framework mode would work but isn't "Remix" by name. Remix v2 is stable, and with all v3 future flags on, it migrates mechanically to React Router framework mode (a codemod plus package renames).

**What it costs:** Remix v2 is in maintenance mode, and new features land in React Router instead. The site will need that migration eventually, or a rewrite if Remix 3 is chosen once it ships.

## The homepage picker reads the example project instead of porting `mnmd filter`

**Chosen:** `CommandMap` walks `app/lib/project.ts` and shows the command a file or folder becomes. Clicking a folder shows the group listing; Tab steps through scripts.

**Rejected:**
- A terminal that ports `mnmd filter` and `mnmd pack` to TypeScript and autoplays keystrokes.
- Compiling those packages to WebAssembly and calling the real code.

**Why:** The page's job is the mapping from path to command, which the example project already records. A second implementation of filter and pack had nothing checking it, and a Go WASM build is a lot of weight for a homepage. The quick start still shows real `mnmd check` output.

**What it costs:** The site no longer demonstrates prefix completion or `mnmd pack` resolving a typed line. Those stay covered by the Go tests and the shell harness.
