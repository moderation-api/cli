## modapi channels update

Update a channel

### Synopsis

Update a channel. Fields you leave out keep their value. Needs the `channels:write` permission.

```
modapi channels update [id] [flags]
```

### Examples

```
  modapi channels update --id <id>
```

### Options

```
      --body string                       Request body as JSON (alternative to individual flags). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF. Use --schema to print the exact JSON Schema.
      --content-type string               The kind of content the channel moderates. (options: profile, message, post, comment, event, product, review, voice, other)
      --context-enabled                   Whether earlier messages in the conversation are read when moderating a message.
      --description string                The channel's description.
      --enable-audio-moderation           Whether audio is moderated.
      --enable-image-moderation           Whether images are moderated.
      --enable-video-moderation           Whether videos are moderated.
      --enable-voice-moderation           Whether real-time voice streams are moderated.
      --flagging-mode string              'ACTIVE' flags content. 'DRY_RUN' evaluates content without flagging it. (options: DRY_RUN, ACTIVE)
      --frame-capture-interval float      Seconds between frames captured from a video.
  -h, --help                              help for update
  -i, --id string                         The ID of the channel to update. (or pass it as the [id] argument)
      --include-transcript-in-response    Whether audio and video transcripts are returned in responses.
  -k, --key string                        The key you pass as 'channel' when you moderate content. Changing it breaks integrations that send the old key.
  -m, --max-frames-per-video int          The most frames moderated per video.
  -n, --name string                       The channel's name.
      --schema                            Print the exact JSON Schema of the request body and exit
  -s, --spread-frames                     Whether frames are spread evenly across the whole video.
  -t, --transcription-quality string      Trade-off between transcription speed and accuracy. (options: SPEED, BALANCED, ACCURACY)
  -u, --unicode-spoof-detection-enabled   Whether look-alike Unicode characters are normalized before moderation.
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

* [modapi channels](modapi_channels.md)	 - Operations for channels

### Machine interface

* `modapi channels update --usage` — this command's flags, defaults and env vars as machine-readable KDL
* `modapi channels update --schema` — the exact JSON Schema of the request body (all `$ref`s bundled)
* `modapi channels update --dry-run` — preview the request without OS-keychain access or a network call (human preview on stderr)
* `--dry-run --output-format json` (or a caller-explicit `--jq`) writes one preview object per request as NDJSON on stdout; jq is not applied to previews
* `--output-format json` or `--jq <expr>` for machine-readable live output; in agent mode errors are a JSON envelope on stderr

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
