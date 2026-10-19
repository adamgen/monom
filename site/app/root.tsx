import type { LinksFunction } from "@remix-run/node";
import {
  isRouteErrorResponse,
  Link,
  Links,
  Meta,
  NavLink,
  Outlet,
  Scripts,
  ScrollRestoration,
  useRouteError,
} from "@remix-run/react";
import type { ReactNode } from "react";
import { Logo } from "~/components/Logo";
import { GITHUB_URL, OG_IMAGE_URL } from "~/lib/site";
import styles from "~/styles.css?url";

export const links: LinksFunction = () => [
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  { rel: "preconnect", href: "https://fonts.gstatic.com", crossOrigin: "anonymous" },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600;700&family=Geist+Mono:wght@400;500;600&display=swap",
  },
  { rel: "stylesheet", href: styles },
  // Favicons, in the logo kit's head-snippet.html order: .ico fallback, then SVG.
  { rel: "icon", href: "/favicon.ico", sizes: "32x32" },
  { rel: "icon", href: "/favicon.svg", type: "image/svg+xml" },
  { rel: "apple-touch-icon", href: "/apple-touch-icon.png" },
  { rel: "manifest", href: "/site.webmanifest" },
];

export function Layout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <meta name="theme-color" content="#07070A" />
        {/* Shared social card. Route meta() replaces, not merges, so these
            live here to cover every page, including the 404. */}
        <meta property="og:site_name" content="monom" />
        <meta property="og:type" content="website" />
        <meta property="og:image" content={OG_IMAGE_URL} />
        <meta property="og:image:width" content="1200" />
        <meta property="og:image:height" content="630" />
        <meta property="og:image:alt" content="The monom logo: your file tree is your command tree." />
        <meta name="twitter:card" content="summary_large_image" />
        <meta name="twitter:image" content={OG_IMAGE_URL} />
        <Meta />
        <Links />
      </head>
      <body>
        <SiteHeader />
        {children}
        <SiteFooter />
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export default function App() {
  return <Outlet />;
}

function SiteHeader() {
  return (
    <header className="site-header">
      <div className="wrap header-inner">
        <Link to="/" className="brand" aria-label="monom home">
          <Logo />
        </Link>
        <nav aria-label="Primary">
          <Link to="/#adopt">How it works</Link>
          <NavLink to="/docs">Quick start</NavLink>
          <Link to="/#compare">Compare</Link>
          <a href={GITHUB_URL} className="nav-gh">
            GitHub
          </a>
        </nav>
      </div>
    </header>
  );
}

function SiteFooter() {
  return (
    <footer className="site-footer">
      <div className="wrap footer-inner">
        <Logo height={22} />
        <p>Your file tree is your command tree.</p>
        <nav aria-label="Footer">
          <Link to="/docs">Quick start</Link>
          <a href={GITHUB_URL}>Source</a>
          <a href={`${GITHUB_URL}/blob/main/constitution.md`}>Constitution</a>
          <a href={`${GITHUB_URL}/blob/main/architecture.md`}>Architecture</a>
        </nav>
      </div>
    </footer>
  );
}

export function ErrorBoundary() {
  const error = useRouteError();
  const notFound = isRouteErrorResponse(error) && error.status === 404;
  return (
    <main className="wrap error-page">
      <pre className="error-term">
        <span className="ps1">$</span> monom {notFound ? "this-page" : "render"}
        {"\n"}
        <span className="is-err">
          {notFound
            ? "mnmd pack: pack: command not found: /this-page"
            : "monom: something went wrong rendering this page"}
        </span>
      </pre>
      <Link to="/" className="button">
        Back to the command tree
      </Link>
    </main>
  );
}
