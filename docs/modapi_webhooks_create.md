## modapi webhooks create

Create a webhook

### Synopsis

Create a webhook subscribed to one or more event types. Deliveries use the v2 envelope and are signed with the project signing secret (see the signing secret endpoint).

```
modapi webhooks create [flags]
```

### Examples

```
  modapi webhooks create --name 'Moderation events' --url https://example.com/webhooks/moderation --event-types QUEUE_ITEM_REJECTED --event-types AUTHOR_BLOCKED
```

### Options

```
      --body string               Request body as JSON (alternative to individual flags). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF. Use --schema to print the exact JSON Schema.
      --description string        The webhook's description
  -e, --event-types stringArray   Event types this webhook subscribes to. One webhook URL receives all events you list here. (options: QUEUE_ITEM_NEW, QUEUE_ITEM_COMPLETED, QUEUE_ITEM_ACTION, QUEUE_ITEM_REJECTED, QUEUE_ITEM_ALLOWED, AUTHOR_BLOCKED, AUTHOR_UNBLOCKED, AUTHOR_SUSPENDED, AUTHOR_UPDATED, AUTHOR_TRUST_LEVEL_CHANGED, AUTHOR_ACTION) [required]
  -h, --help                      help for create
  -n, --name string               The webhook's name, used to identify it in the dashboard [required]
  -p, --project-id string         The project to act on, when you call with an organization key. Optional when the organization has one project. A project key acts on its own project: naming another one is an error.
      --schema                    Print the exact JSON Schema of the request body and exit
  -u, --url string                The webhook's URL. We'll call this URL when an event occurs. [required]
```

### Options inherited from parent commands

```
      --agent-mode             Enable structured errors and default TOON output for AI coding agents. Automatically enabled when a known agent environment is detected (CLAUDECODE, CURSOR_AGENT, etc.). Use --agent-mode=false to disable.
      --color string           Control colored output: auto (color when output is a TTY), always, or never. Respects NO_COLOR and FORCE_COLOR env vars. (default "auto")
  -d, --debug                  Log request and response diagnostics to stderr
      --dry-run                Preview API requests without sending them (no network, no OS keychain). Human preview on stderr; with -o json or --jq, one JSON object per request on stdout. Local mutation commands (auth login, auth logout and configure) make no request: they skip prompts and writes and report a no-op (stderr, or one JSON object on stdout in the machine form)
  -H, --header stringArray     Set a custom HTTP request header (format: "Key: Value"). Can be specified multiple times.
      --include-headers        Include HTTP response headers in the output
      --interactive            Prompt for missing inputs and open guided configure/auth forms (forms fall back to line prompts on stdin off-TTY) (default true)
  -q, --jq string              Filter and transform output using a jq expression (e.g., '.name', '.items[] | .id')
      --no-interactive         Disable all interactive features (auto-prompting, explorer auto-launch, TUI forms)
      --no-keyring             Never read or write the OS keychain; store secrets in the config file instead (env: MODAPI_NO_KEYRING)
  -o, --output-format string   Specify the output format. Options: pretty, json, yaml, table, toon. (default "pretty")
      --raw-output             Write --jq string results as raw text instead of JSON strings (like jq -r); non-string results stay JSON
      --secret-key string      HTTP Bearer
      --server string          Select a server by index (for indexed servers) or name (for named servers)
      --server-url string      Override the default server URL
      --timeout string         HTTP request timeout (e.g., 30s, 5m, 100ms)
      --usage                  Print the CLI Usage schema in KDL format
```

### SEE ALSO

* [modapi webhooks](modapi_webhooks.md)	 - Operations for webhooks

### Machine interface

* `modapi webhooks create --usage` — this command's flags, defaults and env vars as machine-readable KDL
* `modapi webhooks create --schema` — the exact JSON Schema of the request body (all `$ref`s bundled)
* `modapi webhooks create --dry-run` — preview the request without OS-keychain access or a network call (human preview on stderr)
* `--dry-run --output-format json` (or a caller-explicit `--jq`) writes one preview object per request as NDJSON on stdout; jq is not applied to previews
* `--output-format json` or `--jq <expr>` for machine-readable live output; in agent mode errors are a JSON envelope on stderr

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
