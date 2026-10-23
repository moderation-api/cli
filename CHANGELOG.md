# Changelog

## 0.3.2 (2026-10-23)

Full Changelog: [v0.3.1...v0.3.2](https://github.com/moderation-api/cli/compare/v0.3.1...v0.3.2)

### Chores

- Update the generated code.

## 0.3.1 (2026-10-23)

Full Changelog: [v0.3.0...v0.3.1](https://github.com/moderation-api/cli/compare/v0.3.0...v0.3.1)

### Chores

- Update the generated code.

## 0.3.0 (2026-10-22)

Full Changelog: [v0.2.1...v0.3.0](https://github.com/moderation-api/cli/compare/v0.2.1...v0.3.0)

### ⚠ Breaking changes

- `GET /actions`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /actions`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /actions/execute`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /actions/{actionId}/execute`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `DELETE /actions/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /actions/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /actions/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /auth`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /auth`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /authors`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /authors`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `DELETE /authors/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /authors/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /authors/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `DELETE /channels/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /channels/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /channels/{id}/duplicate`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /moderate`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /moderate/image`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /moderate/text`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /projects`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /projects/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /projects/{id}/duplicate`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /projects/{projectId}/channels`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /queue/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /queue/{id}/items`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /queue/{id}/items/{itemId}/resolve`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /queue/{id}/items/{itemId}/unresolve`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /queue/{id}/stats`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /webhooks`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /webhooks/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /wordlist`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /wordlist/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `PUT /wordlist/{id}`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `GET /wordlist/{id}/embedding-status`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `DELETE /wordlist/{id}/words`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`
- `POST /wordlist/{id}/words`: added the new `project_name_ambiguous` enum value to the `code` response property for the response status `400`

### API changes

- a breaking change was detected but the version is still `1.1.0`

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
