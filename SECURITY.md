# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 2.x.x   | :white_check_mark: |
| < 2.0   | :x:                |

## Reporting a Vulnerability

The CodeDNA maintainers take the security of our platform and user source code seriously.

If you discover a security vulnerability, please **do not** create a public GitHub issue. Instead, report it privately via GitHub Private Vulnerability Reporting or email:

`zahed04x@gmail.com`

Please provide:
- A description of the issue and potential impact
- Detailed reproduction steps or Proof of Concept
- Affected files, endpoints, or dependencies

We commit to acknowledging your report within 48 hours and providing a remediation timeline.

## Privacy & Safety Principles

- **No Automatic Code Execution**: CodeDNA statically parses code with Tree-sitter. It never executes arbitrary or untrusted repository code automatically.
- **Privacy First**: CodeDNA operates locally by default using local SQLite storage. Source code and AST models are NEVER sent to third-party AI APIs without explicit opt-in confirmation (`userConsent: true`).
- **Path Traversal Protection**: All filesystem access endpoints enforce strict boundaries to prevent unauthorized directory traversal.
- **Credential Redaction**: Remote Git URLs with embedded authorization tokens are automatically sanitized before indexing and storage.
