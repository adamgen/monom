export const GITHUB_URL = "https://github.com/adamgen/monom";
export const INSTALL_SCRIPT_URL = "https://monom.dev/install.sh";
// The same file straight from the repo, in case monom.dev is unreachable.
export const INSTALL_SCRIPT_MIRROR_URL = "https://raw.githubusercontent.com/adamgen/monom/main/install.sh";
export const INSTALL = `curl -fsSL ${INSTALL_SCRIPT_URL} | bash`;
export const SITE_URL = "https://monom.dev";
// Absolute: social crawlers do not resolve relative og:image URLs.
export const OG_IMAGE_URL = `${SITE_URL}/og-image.png`;
