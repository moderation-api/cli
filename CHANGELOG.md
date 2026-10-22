# Changelog

## 0.2.1 (2026-10-22)

Full Changelog: [v0.2.0...v0.2.1](https://github.com/moderation-api/cli/compare/v0.2.0...v0.2.1)

### Chores

- Update the generated code.

## 0.2.0 (2026-10-22)

Full Changelog: [v0.1.0...v0.2.0](https://github.com/moderation-api/cli/compare/v0.1.0...v0.2.0)

### API changes

- Update to the latest API spec.

## 0.1.0 (2026-10-22)

The first release of `modapi`, the Moderation API command-line interface.

```bash
curl -fsSL https://raw.githubusercontent.com/moderation-api/cli/main/scripts/install.sh | bash
```

- Every API endpoint as a command, for example `modapi content submit` and `modapi authors list`.
- Reads your key from `MODAPI_SECRET_KEY`, the OS keychain (`modapi configure`) or `--secret-key`.
- JSON output and `--jq` filters for scripts, `--all` to fetch every page, and `--dry-run` to see a request without sending it.
- Built for AI coding agents: `--usage` and `--schema` describe every command and request body.
