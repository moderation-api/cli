## modapi content submit

Submit content

### Synopsis

Submit content

```
modapi content submit [flags]
```

### Examples

```
  modapi content submit --content '{"type":"text","text":"Hey, check out my profile for a deal!"}'
```

### Options

```
  -a, --author-id string            The author of the content.
      --body string                 Request body as JSON (alternative to individual flags). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF. Use --schema to print the exact JSON Schema.
      --channel string              Provide a channel ID or key. Will use the project's default channel if not provided.
      --client-action string        A recommendation from your own client-side flagging (e.g. a banned-IP list or a third-party tool). Feeds the rules engine and can escalate or override the recommended action. Does not change whether our analysis flagged the content.
      --content string              JSON value (variants: text: { "text": string }, image: { "url": string, "data": string }, video: { "url": string }, audio: { "url": string }, object: { "data": object })
      --content-id string           The unique ID of the content in your database.
      --content.audio string        Audio_request variant as JSON
      --content.audio.url string    The URL of the audio content [required]
      --content.image string        Image_request variant as JSON
      --content.image.data string   Base64-encoded image data. Either url or data must be provided. Note: base64 images are not stored and will not appear in the review queue.
      --content.image.url string    A public URL of the image content. Either url or data must be provided.
      --content.object string       Object variant as JSON
      --content.text string         Text_request variant as JSON
      --content.text.text string    The content text [required]
      --content.video string        Video_request variant as JSON
      --content.video.url string    A public URL of the video content [required]
      --conversation-id string      For example the ID of a chat room or a post
      --do-not-store                Process the content without saving it on our servers. The item is still created and appears in the review queue, analytics and author history with its moderation results, but its content is empty, so reviewers cannot see what was submitted.
  -h, --help                        help for submit
      --meta-type string            The meta type of content being moderated (options: profile, message, post, comment, event, product, review, voice, other)
      --metadata string             Any metadata you want to store with the content
      --policies string             (Enterprise) override the channel policies for this moderation request only. (JSON array)
      --project-id string           The project to act on, by ID or name. Required with an organization key unless the key reaches only one project; list projects with GET /projects. A project key acts on its own project and needs no value: naming another one is an error.
      --schema                      Print the exact JSON Schema of the request body and exit
  -t, --timestamp float             Unix timestamp (in milliseconds) of when the content was created. Use if content is not submitted in real-time.
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

* [modapi content](modapi_content.md)	 - Operations for content

### Machine interface

* `modapi content submit --usage` — this command's flags, defaults and env vars as machine-readable KDL
* `modapi content submit --schema` — the exact JSON Schema of the request body (all `$ref`s bundled)
* `modapi content submit --dry-run` — preview the request without OS-keychain access or a network call (human preview on stderr)
* `--dry-run --output-format json` (or a caller-explicit `--jq`) writes one preview object per request as NDJSON on stdout; jq is not applied to previews
* `--output-format json` or `--jq <expr>` for machine-readable live output; in agent mode errors are a JSON envelope on stderr

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
