import { THEME } from "~/lib/site";

export function Logo() {
  if (THEME === "pixel") {
    // Round-3 bitmap "Eclipse" mark: the Totality ring redrawn on a 14×14 pixel grid,
    // drawn at a whole-number scale with nearest-neighbour sampling.
    return (
      <span className="logo">
        <img className="logo-mark" src="/pixel/monom-mark.png" width="28" height="28" alt="" />
        monom
      </span>
    );
  }
  return (
    <span className="logo">
      <svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
        <path d="M5 4v16M5 9h7M5 16h7" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
        <circle cx="16" cy="9" r="2.6" fill="var(--group)" />
        <circle cx="16" cy="16" r="2.6" fill="var(--accent)" />
      </svg>
      monom
    </span>
  );
}
