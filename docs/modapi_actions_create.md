## modapi actions create

Create an action

### Synopsis

Create an action.

```
modapi actions create [flags]
```

### Examples

```
  modapi actions create --name 'Suspend for 24 hours'
```

### Options

```
      --body string                       Request body as JSON (alternative to individual flags). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF. Use --schema to print the exact JSON Schema.
  -b, --built-in string                   Whether the action is a built-in action or a custom one. (default "false")
      --description string                The description of the action.
      --filter-in-queue-ids stringArray   The IDs of the queues the action is available in.
      --free-text                         Whether the action allows any text to be entered as a value or if it must be one of the possible values.
  -h, --help                              help for create
  -k, --key string                        User defined key of the action.
  -n, --name string                       The name of the action. [required]
      --position string                   Show the action in all queues, selected queues or no queues (to use via API only). (options: ALL_QUEUES, SOME_QUEUES, HIDDEN) (default "ALL_QUEUES")
      --possible-values string            The possible values of the action. The user will be prompted to select one of these values when executing the action. (JSON array)
      --project-id string                 The project to act on, by ID or name. Required with an organization key unless the key reaches only one project; list projects with GET /projects. A project key acts on its own project and needs no value: naming another one is an error.
      --queue-behaviour string            Whether the action resolves and removes the item, unresolves and re-add it to the queue, or does not change the resolve status. (options: REMOVE, ADD, NO_CHANGE) (default "NO_CHANGE")
      --schema                            Print the exact JSON Schema of the request body and exit
  -t, --type string                       The type of the action. (options: AUTHOR_BLOCK, AUTHOR_BLOCK_TEMP, AUTHOR_UNBLOCK, AUTHOR_DELETE, AUTHOR_REPORT, AUTHOR_WARN, AUTHOR_CUSTOM, ITEM_REJECT, ITEM_ALLOW, ITEM_CUSTOM)
  -v, --value-required                    Whether the action requires a value to be executed.
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

* [modapi actions](modapi_actions.md)	 - Operations for actions

### Machine interface

* `modapi actions create --usage` — this command's flags, defaults and env vars as machine-readable KDL
* `modapi actions create --schema` — the exact JSON Schema of the request body (all `$ref`s bundled)
* `modapi actions create --dry-run` — preview the request without OS-keychain access or a network call (human preview on stderr)
* `--dry-run --output-format json` (or a caller-explicit `--jq`) writes one preview object per request as NDJSON on stdout; jq is not applied to previews
* `--output-format json` or `--jq <expr>` for machine-readable live output; in agent mode errors are a JSON envelope on stderr

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
