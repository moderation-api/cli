<!-- Start Custom [intro] -->
# Moderation API CLI

`modapi` is the official command-line interface for [Moderation API](https://moderationapi.com). Use it to moderate content, manage authors, review queues, actions, wordlists, webhooks, projects and channels from your terminal, in scripts and CI, and from AI coding agents.

- Get an API key in the [dashboard](https://dash.moderationapi.com).
- Read the [documentation](https://docs.moderationapi.com) and the [API reference](https://docs.moderationapi.com/api-reference/introduction).
<!-- End Custom [intro] -->

<!-- Start Summary [summary] -->
## Summary

Moderation API: Moderate text, images, video and audio in real time, manage authors, review queues and actions, and receive webhooks.

More information about the API can be found at https://docs.moderationapi.com
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [Moderation API CLI](#moderation-api-cli)
  * [CLI Installation](#cli-installation)
  * [Shell Completion](#shell-completion)
  * [CLI Example Usage](#cli-example-usage)
  * [Set up Moderation API in your project](#set-up-moderation-api-in-your-project)
  * [Moderating content](#moderating-content)
  * [For AI agents](#for-ai-agents)
  * [Authentication](#authentication)
  * [Commands](#commands)
  * [Request Body Input](#request-body-input)
  * [Server Selection](#server-selection)
  * [Output Formats](#output-formats)
  * [Pagination](#pagination)
  * [Error Handling](#error-handling)
  * [Diagnostics](#diagnostics)
  * [Support](#support)
* [Development](#development)
  * [Maturity](#maturity)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start CLI Installation [installation] -->
## CLI Installation

### Homebrew (macOS/Linux)

```bash
brew install moderation-api/tap/modapi
```

### Quick Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/moderation-api/cli/main/scripts/install.sh | bash
```

### Quick Install (Windows PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/moderation-api/cli/main/scripts/install.ps1 | iex
```

### Go Install

Alternatively, install directly via Go:

```bash
go install github.com/moderation-api/cli/cmd/modapi@latest
```

### Manual Download

Download pre-built binaries for your platform from the [releases page](https://github.com/moderation-api/cli/releases).
<!-- End CLI Installation [installation] -->

<!-- Start Shell Completion [completion] -->
## Shell Completion

Shell completions are available for Bash, Zsh, Fish, and PowerShell.

### Bash

```bash
# Add to ~/.bashrc:
source <(modapi completion bash)

# Or install permanently:
modapi completion bash > /etc/bash_completion.d/modapi
```

### Zsh

```zsh
# Add to ~/.zshrc:
source <(modapi completion zsh)

# Or install permanently:
modapi completion zsh > "${fpath[1]}/_modapi"
```

### Fish

```fish
modapi completion fish | source

# Or install permanently:
modapi completion fish > ~/.config/fish/completions/modapi.fish
```

### PowerShell

```powershell
modapi completion powershell | Out-String | Invoke-Expression
```
<!-- End Shell Completion [completion] -->

<!-- Start CLI Example Usage [usage] -->
## CLI Example Usage

### Example

```bash
modapi content submit --content '{"type":"text","text":"Hey, check out my profile for a deal!"}' --content-id message-48213 --author-id user-1092 --conversation-id room-general --metadata '{"platform":"ios"}'

```
<!-- End CLI Example Usage [usage] -->
<!-- Start Custom [guide] -->
## Set up Moderation API in your project

Run `init` in your project's directory. It signs you in if needed, creates a Moderation API project, and writes the project's keys to your env file:

```bash
modapi login   # once per machine; opens the browser
modapi init    # in your project
modapi doctor  # check the setup at any time
```

`init` detects your language, framework and package manager, writes `MODAPI_SECRET_KEY` and `MODAPI_WEBHOOK_SECRET` to `.env.local` (Next.js) or `.env`, and adds that file to `.gitignore`. Describe your platform so the project can suggest moderation policies:

```bash
modapi init --domain acme.com \
  --context "Marketplace where buyers and sellers chat before a booking" \
  --brief "Stop contact details shared outside the platform, scams and harassment"
```

`init` does not change your code. To add the moderation calls, let your coding agent do it with the [setup skill](https://github.com/moderation-api/skills):

```bash
npx skills add moderation-api/skills
```

Then ask your agent to "set up Moderation API". The skill finds where users create content, adds the moderation call, handles the result and sets up webhooks, asking you before each change.

Without a browser (for example over SSH), use `modapi login --device`. In CI, skip `login` and set `MODAPI_ORGANIZATION_KEY` to an organization key from the dashboard.

## Moderating content

Set your API key once, then submit content:

```bash
export MODAPI_SECRET_KEY="your-project-key"

modapi content submit --content '{"type":"text","text":"Your text here"}' --author-id user-123
```

Images, video and audio take a URL:

```bash
modapi content submit --content '{"type":"image","url":"https://example.com/image.jpg"}'
```

Use `--jq` to read one field of the result, for example the recommended action (`allow`, `review` or `reject`):

```bash
modapi content submit --content '{"type":"text","text":"Your text here"}' --jq '.recommendation.action'
```

Add `--dry-run` to any command to see the request it would send, without sending it.
<!-- End Custom [guide] -->


<!-- Start For AI agents [agents] -->
## For AI agents

This CLI is built to be driven by AI coding agents as well as people: everything an agent needs is discoverable from the binary itself, and every command can be validated without credentials. Work down this ladder:

| Run | You get |
|-----|---------|
| `modapi --help`, `modapi actions list --help` | Commands by category, runnable examples, flags |
| `modapi --usage`, `modapi actions list --usage` | The command surface as machine-readable [KDL](https://kdl.dev): commands, aliases, flags, defaults, env vars, config keys |
| `modapi actions create --schema` | The exact JSON Schema of the command's request body (all `$ref`s bundled) — build a valid `--body` from it |
| `modapi actions list --dry-run` | The exact HTTP request (method, URL, headers, body), with no credentials or network call |
| `modapi actions list --output-format json` (or `--jq`) | Machine-readable output |

### Discover the command surface

```bash
# Every command, flag, default, env var and config key, as KDL
modapi --usage

# One command's subtree only
modapi actions list --usage
```

### Read the exact request schema

`--schema` is available on every command that accepts a request body (`--body`, stdin, or a whole-body flag where the command has one), including intent commands. It prints the JSON Schema the request is validated against and exits without calling the API.

```bash
# JSON Schema (draft 2020-12) of the request body, with every $ref bundled under $defs
modapi actions create --schema
```

### Probe before you spend

Start quota-spending commands with `--dry-run`. It validates inputs, resolves the request, redacts secrets and binary payloads, makes no network call, and exits 0. It never reads the OS keychain; credentials supplied by flag, environment, or config file are included only as `[REDACTED]`.

```bash
# Human preview: the [DRY-RUN] block is on stderr and stdout is empty
modapi actions list --dry-run

# Machine preview: compact JSON on stdout and silent stderr
modapi actions list --dry-run --output-format json
```

The machine form writes one object per would-be request, one per line (NDJSON for multi-request commands), with exactly this shape:

```json
{"dry_run":true,"request":{"method":"POST","url":"https://…","headers":{"Accept":["application/json"],…},"body":<JSON value | string | null>}}
```

`body` is a parsed JSON value when the body is JSON, a string for text, `"<bytes:N>"` for binary data, and `null` when absent. An explicit caller `--jq` also selects this JSON preview protocol, but the filter is not applied to preview objects. Command-declared jq presets do not select or filter the preview.

Local mutation commands make no request under `--dry-run`: instead of a preview they emit one `{"dry_run":true,"local":true,"command":"…","message":"…"}` object. `select(.request)` keeps only would-be requests; `select(.local)` keeps the local no-ops.

### Machine-readable output

```bash
# JSON on stdout
modapi actions list --output-format json

# Filter or reshape with a jq expression (always emits JSON, overrides --output-format)
modapi actions list --jq '.'

# Print jq string results as plain text instead of JSON strings (like jq -r)
modapi actions list --jq '.' --raw-output
```

`--output-format toon` emits [TOON](https://github.com/toon-format/spec), a compact line-oriented format that uses fewer tokens than JSON; it is the default in agent mode.

### Interactive mode
Required-input prompts and guided `configure` / `auth login` forms are enabled by default. Required-input prompts require an interactive terminal; off-TTY forms read line input from stdin. Use `--no-interactive` to force flag-only execution.

```bash
# Prompt for missing command inputs
modapi actions list --interactive

# Open the guided configuration form
modapi configure --interactive

# Explicitly launch the terminal command explorer
modapi explore
```

### Agent mode and structured errors
Agent mode turns on automatically when a known agent environment is detected (`CLAUDECODE`, `CURSOR_AGENT`, `CODEX`, `AIDER`, `CLINE`, `WINDSURF_AGENT`, `GITHUB_COPILOT`, `AMAZON_Q`, `GEMINI_CODE_ASSIST`, `SRC_CODY`) or with `--agent-mode` (`--agent-mode=false` disables detection).
In agent mode interactive prompts never launch, output defaults to TOON, and every failure — API errors and CLI usage errors alike — is one JSON envelope on stderr:
`--output-format json` and `--jq` use the same error envelope without requiring agent mode.

```json
{
  "error": "...",
  "error_type": "validation_error",
  "error_reason": "CLI_VALIDATION",
  "exit_code": 2,
  "message": "human-readable message",
  "hints": ["what to try next"]
}
```

`error_type` is one of `authentication_error`, `authorization_error`, `not_found`, `validation_error`, `rate_limit_error`, `server_error`, `api_error`, `connection_error`, `protocol_error`, `runtime_error`, `unsupported_error`, `async_failed`, `async_timeout`, `async_unknown_state`. Classification derives from the HTTP status and transport evidence; `error_reason` is absent for API errors. Status-less local failures may use `CLI_VALIDATION`, `CLI_CONNECTION`, `CLI_PROTOCOL`, `CLI_RUNTIME`, `CLI_UNAVAILABLE`, `CLI_AUTHENTICATION`, or the async polling reasons `CLI_ASYNC_FAILED`, `CLI_ASYNC_TIMEOUT`, and `CLI_ASYNC_UNKNOWN_STATE`. `hints` preserves server guidance first, adds the most specific local taxonomy guidance, then typed CLI and command-specific guidance, removing exact duplicates. `exit_code` is always the code for the final `error_type` shown in the envelope: 1 runtime, 2 usage, or 3 authentication/authorization.

### Lists, streams, and files

List commands accept `--all` to fetch every page and stream results as they arrive (one JSON value per line with `--output-format json`; `--max-pages N` bounds the walk).

Structured output and agent mode never write pagination hints to stderr; if a later page fails or the server repeats a cursor, the command exits non-zero after the pages already written.

```bash
modapi actions list --all --output-format json
```
<!-- End For AI agents [agents] -->

<!-- Start Authentication [security] -->
## Authentication

Authentication credentials can be configured in four ways (in order of priority):

### 1. Command-line flags

Pass credentials directly as flags to any command:

```bash
modapi --secret-key "$MODAPI_SECRET_KEY" actions list
```

### 2. Environment variables

Set credentials via environment variables:

| Variable | Description |
|----------|-------------|
| `MODAPI_SECRET_KEY` | HTTP Bearer |

### 3. OS Keychain (recommended for workstations)

Credentials are stored securely in your operating system's keychain when you run:

```bash
modapi configure
```

Secret credentials (tokens, API keys, passwords) are automatically stored in:
- **macOS**: Keychain
- **Linux**: GNOME Keyring / KWallet (via D-Bus Secret Service)
- **Windows**: Windows Credential Locker

If no keychain is available (e.g., in CI environments), credentials fall back to the config file.

Where the keychain cannot be unlocked (e.g., headless Linux or SSH sessions with a locked GNOME Keyring), skip it entirely with `--no-keyring`, `MODAPI_NO_KEYRING=true`, or `no_keyring: true` in the config file. The keychain is then never read or written, and secrets are stored in the config file instead; flags and environment variables still take precedence.

### 4. Configuration file

Run the interactive `configure` command to store non-secret settings:

```bash
modapi configure
```

Configuration is stored in `~/.config/modapi/config.yaml`.
<!-- End Authentication [security] -->
<!-- Start Custom [auth] -->
### Project and organization keys

Moderation API has two kinds of API keys:

- A **project key** works with one project: content, authors, review queues, actions, wordlists, webhooks and the account.
- An **organization key** works with the [management API](https://docs.moderationapi.com/api-reference/management-api): `modapi projects` and `modapi channels`. It can also use any project's endpoints: name the project with `--project-id`.

A project key can't run the management commands: the API returns `403` with the code `organization_key_required`. To run a management command while `MODAPI_SECRET_KEY` holds your project key, pass the organization key for that command:

```bash
modapi projects list --secret-key "$MODAPI_ORGANIZATION_KEY"
```
<!-- End Custom [auth] -->


<!-- Start Commands [operations] -->
## Commands

<details open>
<summary>Available commands</summary>

* [`authors`](docs/modapi_authors.md) - Operations for authors
  * [`list`](docs/modapi_authors_list.md) - List authors
  * [`create`](docs/modapi_authors_create.md) - Create a new author
  * [`retrieve`](docs/modapi_authors_retrieve.md) - Get author details
  * [`update`](docs/modapi_authors_update.md) - Update author details
  * [`delete`](docs/modapi_authors_delete.md) - Delete an author
* [`queue`](docs/modapi_queue.md) - Operations for queue
  * [`retrieve`](docs/modapi_queue_retrieve.md) - Get a queue
  * [`get-stats`](docs/modapi_queue_get-stats.md) - Get queue statistics
  * [`items`](docs/modapi_queue_items.md) - Operations for items
    * [`list`](docs/modapi_queue_items_list.md) - Get queue items
    * [`resolve`](docs/modapi_queue_items_resolve.md) - Resolve a queue item
    * [`unresolve`](docs/modapi_queue_items_unresolve.md) - Unresolve a queue item
* [`actions`](docs/modapi_actions.md) - Operations for actions
  * [`list`](docs/modapi_actions_list.md) - List moderation actions
  * [`create`](docs/modapi_actions_create.md) - Create an action
  * [`retrieve`](docs/modapi_actions_retrieve.md) - Get an action
  * [`update`](docs/modapi_actions_update.md) - Update an action
  * [`delete`](docs/modapi_actions_delete.md) - Delete an action
  * [`execute`](docs/modapi_actions_execute.md) - Execute moderation action
* [`webhooks`](docs/modapi_webhooks.md) - Operations for webhooks
  * [`list`](docs/modapi_webhooks_list.md) - List webhooks
  * [`create`](docs/modapi_webhooks_create.md) - Create a webhook
  * [`retrieve`](docs/modapi_webhooks_retrieve.md) - Get a webhook
  * [`update`](docs/modapi_webhooks_update.md) - Update a webhook
  * [`delete`](docs/modapi_webhooks_delete.md) - Delete a webhook
* [`webhook-secret`](docs/modapi_webhook-secret.md) - Operations for webhook-secret
  * [`retrieve`](docs/modapi_webhook-secret_retrieve.md) - Get the webhook signing secret
* [`account`](docs/modapi_account.md) - Operations for account
  * [`retrieve`](docs/modapi_account_retrieve.md) - Get account details
* [`wordlist`](docs/modapi_wordlist.md) - Operations for wordlist
  * [`list`](docs/modapi_wordlist_list.md) - List wordlists
  * [`retrieve`](docs/modapi_wordlist_retrieve.md) - Get wordlist
  * [`update`](docs/modapi_wordlist_update.md) - Update wordlist
  * [`get-embedding-status`](docs/modapi_wordlist_get-embedding-status.md) - Get embedding status
  * [`words`](docs/modapi_wordlist_words.md) - Operations for words
    * [`add`](docs/modapi_wordlist_words_add.md) - Add words to wordlist
    * [`remove`](docs/modapi_wordlist_words_remove.md) - Remove words from wordlist
* [`content`](docs/modapi_content.md) - Operations for content
  * [`submit`](docs/modapi_content_submit.md) - Submit content
* [`projects`](docs/modapi_projects.md) - Operations for projects
  * [`list`](docs/modapi_projects_list.md) - List projects
  * [`create`](docs/modapi_projects_create.md) - Create a project
  * [`retrieve`](docs/modapi_projects_retrieve.md) - Get a project
  * [`update`](docs/modapi_projects_update.md) - Update a project
  * [`delete`](docs/modapi_projects_delete.md) - Delete a project
  * [`duplicate`](docs/modapi_projects_duplicate.md) - Duplicate a project
  * [`api-keys`](docs/modapi_projects_api-keys.md) - Operations for api-keys
    * [`create`](docs/modapi_projects_api-keys_create.md) - Create a project API key
* [`channels`](docs/modapi_channels.md) - Operations for channels
  * [`list`](docs/modapi_channels_list.md) - List channels
  * [`create`](docs/modapi_channels_create.md) - Create a channel
  * [`retrieve`](docs/modapi_channels_retrieve.md) - Get a channel
  * [`update`](docs/modapi_channels_update.md) - Update a channel
  * [`delete`](docs/modapi_channels_delete.md) - Delete a channel
  * [`duplicate`](docs/modapi_channels_duplicate.md) - Duplicate a channel

</details>
<!-- End Commands [operations] -->

<!-- Start Request Body Input [stdinpiping] -->
## Request Body Input

Commands that accept a request body take it three ways, with a clear priority chain. The examples use `modapi actions create`; every body-bearing command works the same way and prints its exact request schema with `--schema`.

### `--body` flag

Provide the entire request body as a JSON string:

```bash
modapi actions create --body '{"name":"Suspend for 24 hours","key":"suspend-24h","type":"AUTHOR_BLOCK_TEMP","queueBehaviour":"REMOVE"}'
```

### Stdin piping (lowest priority)

Pipe JSON into any command that accepts a request body:

```bash
echo '{"name":"Suspend for 24 hours","key":"suspend-24h","type":"AUTHOR_BLOCK_TEMP","queueBehaviour":"REMOVE"}' | modapi actions create
```

This is useful for chaining commands, reading from files, or scripting:

```bash
# Read body from a file
modapi actions create < request.json

# Pipe from another command
curl -s https://example.com/request.json | modapi actions create
```

### Priority

When multiple input methods are used, the priority is:

| Priority | Source | Description |
|----------|--------|-------------|
| 1 (highest) | Individual flags | A field flag always wins |
| 2 | `--body` flag | Whole-body JSON via flag |
| 3 (lowest) | Stdin | Piped JSON input |
<!-- End Request Body Input [stdinpiping] -->

<!-- Start Server Selection [server] -->
## Server Selection

### Override Server URL

Use `--server-url` to override the server URL entirely, bypassing any named or indexed server selection:

```bash
modapi --server-url https://custom-api.example.com actions list
```

**Precedence**: `--server-url` > `--server` > default
<!-- End Server Selection [server] -->

<!-- Start Output Formats [output-formats] -->
## Output Formats

Every command supports a `--output-format` flag that controls how the response is rendered to stdout.

### Available formats

| Format | Flag | Description |
|--------|------|-------------|
| Pretty | `--output-format pretty` (default) | Aligned key-value pairs with color, nested indentation. Human-readable at a glance. |
| JSON | `--output-format json` | JSON output. Passthrough when the response is already JSON (preserves original field order and numeric precision). Falls back to typed marshaling otherwise. |
| YAML | `--output-format yaml` | YAML output via standard marshaling. |
| Table | `--output-format table` | Tabular output for array responses. |
| TOON | `--output-format toon` | [Token-Oriented Object Notation](https://github.com/toon-format/spec) — a compact, line-oriented format that typically uses 30–60% fewer tokens than JSON. Well-suited for piping responses into LLM prompts. |

```bash
# Default pretty output
modapi actions list

# Machine-readable JSON
modapi actions list --output-format json

# TOON for LLM-friendly compact output
modapi actions list --output-format toon

# Pipe JSON to jq without using --output-format
modapi actions list --output-format json | jq '.'
```

### jq filtering

Use `--jq` to filter or transform the response inline using a [jq](https://jqlang.org) expression. This always outputs JSON and overrides `--output-format`:

```bash
# Extract a single field
modapi actions list --jq '.'

# Reshape with any jq program; --raw-output prints string results as plain text (like jq -r)
modapi actions list --jq '.' --raw-output
```

### Color control

Use `--color` to control terminal colors:

| Value | Behavior |
|-------|----------|
| `auto` (default) | Color when stdout is a TTY, plain text otherwise |
| `always` | Always colorize |
| `never` | Never colorize |

The `NO_COLOR` and `FORCE_COLOR` environment variables are also respected.

### Streaming and pagination

When using `--all` (pagination) or streaming operations, output is written incrementally as items arrive:

| Format | Streaming behavior |
|--------|-------------------|
| `json` | One compact JSON object per line ([NDJSON](https://github.com/ndjson/ndjson-spec)) |
| `yaml` | YAML documents separated by `---` |
| `toon` | One TOON-encoded object per block, separated by blank lines |
| `pretty` (default) | Pretty-printed items separated by blank lines |
<!-- End Output Formats [output-formats] -->

<!-- Start Pagination [pagination] -->
## Pagination

Some operations in this CLI support automatic pagination. These operations accept `--all` to automatically fetch all pages and stream results incrementally.

### Basic usage

```bash
# Fetch a single page (default behavior)
modapi actions list

# Automatically fetch all pages
modapi actions list --all
```

### Limiting pages

Use `--max-pages` with `--all` to cap the number of pages fetched. A negative value is invalid; `0` means unlimited. Passing `--max-pages` without `--all` is an error.

```bash
# Fetch at most 5 pages
modapi actions list --all --max-pages 5
```

### Output formats

When using `--all`, output is streamed as each page is fetched. Operations whose pagination declaration names an `outputs.results` array emit one item at a time. Other operations emit one complete page object at a time, preserving the single-page response shape and any continuation cursor.

| Format | Behavior |
|--------|----------|
| `--output-format json` | One JSON object per line ([NDJSON](https://github.com/ndjson/ndjson-spec)) |
| `--output-format yaml` | YAML documents separated by `---` |
| `--output-format toon` | One TOON-encoded block per item, separated by blank lines |
| Default (pretty) | Pretty-printed items separated by blank lines |

```bash
# Stream all results as NDJSON
modapi actions list --all --output-format json

# Pipe to jq for further processing
modapi actions list --all --output-format json | jq '.'

# Use the built-in --jq flag
modapi actions list --all --jq '.'
```

### How it works

Under the hood, `--all` calls the operation once, then follows the underlying `Next()` pagination closure to fetch subsequent pages. Results are written to stdout as they arrive rather than buffered in memory, so this works well even with large result sets.

Without `--all`, paginated operations behave like any other command — pass cursor, page, offset, or limit flags manually and get a single page of results. In pretty or table output, a cursor response that proves another page exists prints a hint on stderr. JSON, YAML, TOON, `--jq`, and agent mode keep stderr silent on success. Offset/limit responses do not guess from a full result page. Cursor operations that declare both a results array and a mutable limit also suppress the hint because the client cannot safely reproduce the SDK's runtime limit check.

Pagination can fail after earlier pages have already been written. A later-page API failure or a repeated/cyclic continuation cursor stops with a non-zero exit status; callers should treat stdout as partial whenever the command exits non-zero. `--all` tracks cursor values and stops before issuing another request when the server repeats one; the same applies to next URLs when the target generator supports them.
<!-- End Pagination [pagination] -->

<!-- Start Error Handling [errors] -->
## Error Handling

The CLI uses standard exit codes to indicate success or failure:

| Exit Code | Meaning |
|-----------|---------|
| `0` | Success |
| `1` | Runtime/API failure |
| `2` | Usage or input failure |
| `3` | Authentication or authorization failure |

On success, the response data is printed to **stdout** as JSON. On failure, error details are printed to **stderr**.

```bash
# Capture output and handle errors
modapi actions list --output-format json > output.json 2> error.log
if [ $? -ne 0 ]; then
  echo "Error occurred, see error.log"
fi
```
In pretty mode, each error is printed once as `Error (<type>): <message>`, followed by its reason/HTTP status, actionable `Fix:` bullets, and only non-duplicative residual details.

In agent mode, or with explicit `--output-format json`, `--output-format toon`, or `--jq` machine output, stderr is one classified JSON envelope with `exit_code`, `error_type`, optional `error_reason`, `message`, `hints`, and optional `status_code` — see [For AI agents](#for-ai-agents).
<!-- End Error Handling [errors] -->

<!-- Start Diagnostics [diagnostics] -->
## Diagnostics

The CLI includes two diagnostic flags available on all commands:

### Dry Run

Preview what would be sent without making any network calls:

```bash
modapi actions list --dry-run
```

In human output modes, stdout is empty and the `[DRY-RUN]` block goes to stderr. It includes:
- HTTP method and URL
- Request headers (sensitive values redacted)
- Request body preview (sensitive fields redacted)

With `--output-format json`, or with a caller-explicit `--jq`, stderr is silent and stdout is NDJSON: one compact preview object per would-be request. The jq filter is not applied, and command-declared jq presets do not select the JSON protocol.

```json
{"dry_run":true,"request":{"method":"POST","url":"https://…","headers":{"Accept":["application/json"],…},"body":<JSON value | string | null>}}
```

JSON bodies remain structured; text bodies are strings; binary bodies are `"<bytes:N>"`; absent bodies are `null`. Headers retain all values as arrays, with credentials replaced by `[REDACTED]`. Dry-run never reads the OS keychain, but credentials supplied by flag, environment, or config file still appear redacted. The command exits successfully without contacting the API.

Local mutation commands emit one `{"dry_run":true,"local":true,"command":"…","message":"…"}` object in place of a preview; filter with `select(.request)` or `select(.local)`.

### Debug

Log request and response diagnostics while running normally:

```bash
modapi actions list --debug
```

Debug output goes to stderr and includes:
- Request method, URL, headers, and body preview
- Response status, headers, and body preview
- Transport errors (if any)

The command still executes normally and produces its regular output on stdout.

### Flag Precedence

If both `--dry-run` and `--debug` are set, `--dry-run` takes precedence and no network calls are made.

### Security

Sensitive information is automatically redacted in diagnostic output:
- **Headers**: `Authorization`, `Cookie`, `Set-Cookie`, `X-API-Key`, and other security headers show `[REDACTED]`
- **Body**: JSON fields named `password`, `secret`, `token`, `api_key`, `client_secret`, etc. show `[REDACTED]`
- **Binary data**: binary media and canonical base64 strings are replaced with `<bytes:N>`
- **URL query**: credential-like query parameters are replaced with `[REDACTED]`

Diagnostic output should still be treated as potentially sensitive operational data.
<!-- End Diagnostics [diagnostics] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

<!-- Start Custom [more] -->
## Support

- [API reference](https://docs.moderationapi.com/api-reference/introduction)
- [Errors and error codes](https://docs.moderationapi.com/api-reference/errors)
- [Rate limits](https://docs.moderationapi.com/api-reference/rate-limits)
- [Guides](https://docs.moderationapi.com/get-started/introduction)
- Email [support@moderationapi.com](mailto:support@moderationapi.com)

Found a bug? Open an issue in this repository. This package is generated from the [Moderation API OpenAPI spec](https://api.moderationapi.com/v1/openapi.json), so pull requests to the generated code can't be merged as they are, but we use them to fix the generator.
<!-- End Custom [more] -->

# Development

## Maturity

This CLI is in beta, and there may be breaking changes between versions without a major version update. Therefore, we recommend pinning usage
to a specific package version. This way, you can install the same version each time without breaking changes unless you are intentionally
looking for the latest version.

## Contributions

This CLI is generated programmatically. Edits to generated files are overwritten on regeneration. To customize it:

- **Configuration and behavior:** Use [OpenAPI overlays](https://www.speakeasy.com/docs/prep-openapi/overlays/create-overlays) in the Speakeasy workflow with `x-speakeasy-*` extensions (for example, `x-speakeasy-cli-commands`) to define commands, flags, help text, examples, authentication, and grouping.
- **Persistent code changes:** Store unified diffs as [patch files](https://www.speakeasy.com/docs/sdks/customize/code/patch-files/patch-files) at `.speakeasy/patches/<path-of-generated-file>.patch`; they are re-applied on every generation.

### CLI Created by [Speakeasy](https://www.speakeasy.com/?utm_source=github-com/moderation-api/cli&utm_campaign=cli)
