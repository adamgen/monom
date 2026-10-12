import { useState } from "react";

export function CopyCommand({ command, label }: { command: string; label?: string }) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1400);
    } catch {
      // Clipboard can be unavailable (insecure origin, permissions); the text stays selectable.
    }
  };
  return (
    <div className="copy-command">
      {label && <span className="copy-label">{label}</span>}
      <code>
        <span className="ps1">$</span> {command}
      </code>
      <button type="button" onClick={copy} aria-label={`Copy: ${command}`}>
        {copied ? "copied" : "copy"}
      </button>
    </div>
  );
}
