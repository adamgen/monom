# site/ tradeoffs

## Remix v2, not Remix 3 or React Router framework mode

**Chosen:** Remix v2 (`@remix-run/*` 2.17, the npm `latest` tag) on Vite, with every v3 future flag enabled.

**Rejected:**
- *Remix 3* (`remix@next`, 3.0.0-rc at the time of writing) — a ground-up rewrite that drops React.
- *React Router v7+/v8 framework mode* — the official successor to Remix v2, with the same file-route and loader model under a different package name.

**Why:** The site was asked to be Remix-based. Remix 3 was still a release candidate with a new component model and little documentation, so building a public page on it risked churn through each RC. React Router framework mode would work but isn't "Remix" by name. Remix v2 is stable, and with all v3 future flags on, it migrates mechanically to React Router framework mode (a codemod plus package renames).

**What it costs:** Remix v2 is in maintenance mode, and new features land in React Router instead. The site will need that migration eventually, or a rewrite if Remix 3 is chosen once it ships.

## The homepage demo ports `mnmd filter` to TypeScript instead of running the real binary

**Chosen:** `app/lib/filter.ts` reimplements `internal/filter` in about 30 lines, and the demo runs it in the browser against the site's example project (`app/lib/project.ts`).

**Rejected:** compiling `internal/filter` to WebAssembly (`GOOS=js GOARCH=wasm`, or TinyGo) and calling the real code.

**Why:** Standard Go WASM output starts at ~2 MB (TinyGo ~100 KB) and needs `wasm_exec.js` plus a Go toolchain in the site build. That is a lot of weight and build coupling for one pure function that fits on a screen.

**What it costs:** The two implementations can drift. `architecture.md` lists the mirrored files. A change to filter semantics must be ported by hand, and nothing automated catches a miss.

## The pixel variant is a theme layer, not a fork

**Chosen:** `app/theme-pixel.css`, loaded after `styles.css` and scoped to `<html data-theme="pixel">`, switched by `THEME` in `app/lib/site.ts`. It restyles the existing markup (palette, VT323 headings, JetBrains Mono body, 2px borders, stepped shadows, box-drawing labels via pseudo-elements). The only component change is `Logo`, which shows the round-3 bitmap "Eclipse" mark when the theme is `pixel`. Fonts are self-hosted subsets in `public/fonts` (SIL OFL 1.1, licenses alongside).

**Rejected:** forking the routes or components into pixel copies.

**Why:** The variant must keep copy, sections and the shared `FileTree` identical, so it is easiest to compare and to keep in sync when the markup is literally the same. Setting `THEME = "default"` restores the original look (Geist from Google Fonts, the file-tree logo).

**What it costs:** Some decoration (frame labels such as `┤ ~/acme ├`, prompt glyphs, `[01/03]`) lives in CSS `content`, so it is invisible to the copy in the TSX. Each `content` also carries an empty alternative text (`content: "[" / ""`) so screen readers skip it; browsers without that syntax fall back to the plain declaration and may read it. Overrides depend on the base class names; renaming a class in `styles.css` means renaming it here too.
