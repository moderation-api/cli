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
			lines[i] = key + "=" + value
			written[key] = true
		}
	}
	for _, key := range order {
		if value, set := values[key]; set && !written[key] {
			lines = append(lines, key+"="+value)
		}
	}
	mode := os.FileMode(0o600)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), mode)
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
			raw := strings.TrimSpace(line[strings.Index(line, "=")+1:])
			value = strings.Trim(raw, `"'`)
		}
	}
	return value
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

// gitIgnoreState is whether git ignores a file: "ignored", "not_ignored", or "no_git".
func gitIgnoreState(dir, file string) string {
	if _, err := exec.LookPath("git"); err != nil {
		return "no_git"
	}
	inside := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	if out, err := inside.Output(); err != nil || strings.TrimSpace(string(out)) != "true" {
		return "no_git"
	}
	check := exec.Command("git", "-C", dir, "check-ignore", "-q", file)
	if check.Run() == nil {
		return "ignored"
	}
	return "not_ignored"
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
