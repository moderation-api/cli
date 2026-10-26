## modapi queue items resolve-many

Resolve queue items

### Synopsis

Mark up to 100 items of a queue as resolved in one request. Each item is resolved the same as with `POST /queue/{id}/items/{itemId}/resolve`. IDs that do not exist in the project are skipped and returned in `notFoundIds`.

```
modapi queue items resolve-many [id] [flags]
```

### Examples

```
  modapi queue items resolve-many --id <id> --item-ids clx8k2v0e000108l4a9bz3f7q --item-ids clx8k2v0e000208l4h2mc6d1r
```

### Options

```
      --body string            Request body as JSON (alternative to individual flags). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF. Use --schema to print the exact JSON Schema.
  -h, --help                   help for resolve-many
  -i, --id string              The queue ID. (or pass it as the [id] argument)
      --item-ids stringArray   The item IDs to resolve. Up to 100. [required]
  -p, --project-id string      The project to act on, by ID or name. Required with an organization key unless the key reaches only one project; list projects with GET /projects. A project key acts on its own project and needs no value: naming another one is an error.
      --schema                 Print the exact JSON Schema of the request body and exit
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

* [modapi queue items](modapi_queue_items.md)	 - Operations for items

### Machine interface

* `modapi queue items resolve-many --usage` — this command's flags, defaults and env vars as machine-readable KDL
* `modapi queue items resolve-many --schema` — the exact JSON Schema of the request body (all `$ref`s bundled)
* `modapi queue items resolve-many --dry-run` — preview the request without OS-keychain access or a network call (human preview on stderr)
* `--dry-run --output-format json` (or a caller-explicit `--jq`) writes one preview object per request as NDJSON on stdout; jq is not applied to previews
* `--output-format json` or `--jq <expr>` for machine-readable live output; in agent mode errors are a JSON envelope on stderr

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
