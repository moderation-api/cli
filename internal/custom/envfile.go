package custom

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// setEnvValues writes key=value lines into an env file: it replaces existing
// lines for the same keys and appends the rest. The file is created 0600.
func setEnvValues(path string, values map[string]string, order []string) error {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	lines := []string{}
	if len(data) > 0 {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}
	written := map[string]bool{}
	for i, line := range lines {
		key, ok := envLineKey(line)
		if !ok {
			continue
		}
		if value, set := values[key]; set {
			prefix := ""
			// Keep "export " so files that shells source keep working.
			if strings.HasPrefix(strings.TrimSpace(line), "export ") {
				prefix = "export "
			}
			lines[i] = prefix + key + "=" + quoteEnvValue(value)
			written[key] = true
		}
	}
	for _, key := range order {
		if value, set := values[key]; set && !written[key] {
			lines = append(lines, key+"="+quoteEnvValue(value))
		}
	}
	mode := os.FileMode(0o600)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), mode)
}

// quoteEnvValue double-quotes values that dotenv parsers would otherwise misread.
func quoteEnvValue(value string) string {
	if !strings.ContainsAny(value, " \t#$\"'\\`") {
		return value
	}
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "$", `\$`, "`", "\\`").Replace(value) + `"`
}

var envLinePattern = regexp.MustCompile(`^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=`)

func envLineKey(line string) (string, bool) {
	match := envLinePattern.FindStringSubmatch(line)
	if match == nil {
		return "", false
	}
	return match[1], true
}

// readEnvValue returns the value of key in an env file, without quotes.
func readEnvValue(path, key string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	value := ""
	for scanner.Scan() {
		line := scanner.Text()
		if k, ok := envLineKey(line); ok && k == key {
			value = parseEnvValue(line[strings.Index(line, "=")+1:])
		}
	}
	return value
}

// parseEnvValue reads a dotenv value: one matching pair of quotes is
// removed, and an unquoted value ends at an inline " #" comment.
func parseEnvValue(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 2 && (raw[0] == '"' || raw[0] == '\'') {
		if end := closingQuote(raw); end > 0 {
			inner := raw[1:end]
			if raw[0] == '"' {
				inner = strings.NewReplacer(`\"`, `"`, `\$`, "$", "\\`", "`", `\\`, `\`).Replace(inner)
			}
			return inner
		}
	}
	if i := strings.Index(raw, " #"); i >= 0 {
		raw = raw[:i]
	}
	return strings.TrimSpace(raw)
}

// closingQuote finds the quote that closes raw[0]; inside double quotes a
// backslash escapes the next character. It returns -1 when there is none.
func closingQuote(raw string) int {
	q := raw[0]
	for i := 1; i < len(raw); i++ {
		switch {
		case q == '"' && raw[i] == '\\':
			i++
		case raw[i] == q:
			return i
		}
	}
	return -1
}

// envFilesFor are the env files a project in dir may keep its key in, most specific first.
func envFilesFor(dir string, s *stack) []string {
	names := []string{".env.local", ".env"}
	if s != nil && s.EnvFile != "" && s.EnvFile != ".env.local" && s.EnvFile != ".env" {
		names = append([]string{s.EnvFile}, names...)
	}
	paths := make([]string, 0, len(names))
	for _, name := range names {
		paths = append(paths, filepath.Join(dir, name))
	}
	return paths
}

// findEnvValue looks for key in the project's env files and returns the value and the file.
func findEnvValue(dir string, s *stack, key string) (string, string) {
	for _, path := range envFilesFor(dir, s) {
		if v := readEnvValue(path, key); v != "" {
			return v, path
		}
	}
	return "", ""
}

// gitIgnoreState says how git treats a file: "tracked" (committed, so
// .gitignore cannot help), "ignored", "not_ignored", "no_git", or "unknown"
// (for example a file outside the repository).
func gitIgnoreState(dir, file string) string {
	if _, err := exec.LookPath("git"); err != nil {
		return "no_git"
	}
	inside := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	if out, err := inside.Output(); err != nil || strings.TrimSpace(string(out)) != "true" {
		return "no_git"
	}
	if exec.Command("git", "-C", dir, "ls-files", "--error-unmatch", "--", file).Run() == nil {
		return "tracked"
	}
	err := exec.Command("git", "-C", dir, "check-ignore", "-q", "--", file).Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return "ignored"
	case errors.As(err, &exitErr) && exitErr.ExitCode() == 1:
		return "not_ignored"
	default:
		return "unknown"
	}
}

// ensureGitIgnored adds the env file to .gitignore when git would track it.
// It reports whether it changed .gitignore.
func ensureGitIgnored(dir, file string) (bool, error) {
	if gitIgnoreState(dir, file) != "not_ignored" {
		return false, nil
	}
	path := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	text := string(data)
	if len(text) > 0 && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	text += "\n# Moderation API secrets\n" + filepath.Base(file) + "\n"
	return true, os.WriteFile(path, []byte(text), 0o644)
}
