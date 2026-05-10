# Security Policy

## Threat model

`mira` is a **read-only** terminal file browser. It walks
directories, reads metadata (`os.Lstat` / `os.ReadDir`), and reads
`.gitignore` text contents. It does **not**:

- Open, parse, or execute any file other than `.gitignore`.
- Follow symlinks during traversal — symlinks are listed but not descended.
- Modify any file, write logs, or persist state.
- Make network calls.
- Run shell commands.

The only filesystem reads are directory listings the invoking user already
has permission to see, plus `.gitignore` text files within the discovered
git root.

## Supported versions

We support the latest released minor version. Older versions may receive
security fixes at the maintainers' discretion.

| Version | Supported |
| ------- | --------- |
| latest  | ✅        |
| < latest| ❌        |

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security problems.

Use GitHub's
[private vulnerability reporting](https://github.com/mondial7/mira/security/advisories/new)
instead. Include:

- A description of the issue.
- Steps to reproduce.
- The version of `mira` affected.
- Your assessment of the impact.

You should receive an acknowledgement within 5 business days. We will work
with you on a coordinated disclosure timeline (typically ≤ 90 days).

## Build & supply chain

- Releases are built reproducibly via [GoReleaser](https://goreleaser.com/)
  in GitHub Actions on tag push.
- Each release ships with SHA-256 checksums and an SBOM
  (`*.sbom.cdx.json`).
- Dependencies are tracked by Dependabot and scanned by `govulncheck` on
  every PR.
- Source is scanned by GitHub CodeQL on push and PR.
