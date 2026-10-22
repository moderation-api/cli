## modapi init

Set up Moderation API in this project

### Synopsis

Set up Moderation API in the project in the current directory.

init detects the language, framework and package manager, creates a
Moderation API project, and writes its secret key (MODAPI_SECRET_KEY) and
webhook signing secret (MODAPI_WEBHOOK_SECRET) to the project's env file. It
adds the env file to .gitignore when git would track it.

Describe the platform with --domain, --context and --brief: the project uses
them to suggest moderation policies. Coding agents can read these from the
codebase and confirm them with the user before running init.

init needs an organization key to create the project: sign in with
"modapi login", or set MODAPI_ORGANIZATION_KEY. When the env file already has
MODAPI_SECRET_KEY, init checks it and creates nothing (use --force for a new
project).

init does not change your code. Use the moderation-api-setup agent skill, or
the guide at https://docs.moderationapi.com, to add the moderation calls.

```
modapi init [flags]
```

### Examples

```
  modapi init
  modapi init --name "Acme Marketplace" --domain acme.com \
    --context "Marketplace where buyers and sellers chat before a booking" \
    --brief "Stop contact details and payment requests outside the platform, scams and harassment"
  modapi init --dry-run -o json
```

### Options

```
  -b, --brief string              What to moderate, the problems you see and your guidelines
  -c, --context string            What the platform is and who uses it
      --dir string                The project directory (default ".")
      --domain string             The website or app the project moderates content for, e.g. acme.com
      --env-file string           Env file to write the secrets to (default: .env.local for Next.js, else .env)
      --force                     Create a new project even when the env file already has a key
  -h, --help                      help for init
      --install                   Also install the SDK with the detected package manager
  -n, --name string               Project name (default: the package or directory name)
      --organization-key string   Organization key to create the project with (env: MODAPI_ORGANIZATION_KEY; default: the key from modapi login)
      --used-by-minors            Children use the platform (stricter policies)
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

* [modapi](modapi.md)	 - Moderation API: Moderate text, images, video and audio in real time, manage authors, review queues and actions, and receive webhooks

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
