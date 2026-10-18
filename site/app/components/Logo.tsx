// The Totality header lockup (eclipse ring, gold glint, Sora wordmark), dark
// variant. Loaded with <img> rather than inlined so its SVG ids (masks,
// gradients) can never collide with another inlined SVG on the page.
// Source: the monom logo kit, svg/monom-horizontal-dark.svg (1660.4 x 473).
export function Logo({ height = 24 }: { height?: number }) {
  return (
    <img
      className="logo"
      src="/brand/monom-horizontal-dark.svg"
      alt="monom"
      height={height}
      width={Math.round((height * 1660.4) / 473)}
      decoding="async"
    />
  );
}
