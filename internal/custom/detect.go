package custom

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// stack is what `init` and `doctor` find out about the project in a directory.
type stack struct {
	Language       string `json:"language"`
	Framework      string `json:"framework,omitempty"`
	PackageManager string `json:"package_manager,omitempty"`
	// SDKPackage is empty for languages without an official SDK.
	SDKPackage     string `json:"sdk_package,omitempty"`
	SDKInstalled   bool   `json:"sdk_installed"`
	InstallCommand string `json:"install_command,omitempty"`
	// EnvFile is where this kind of project keeps local secrets.
	EnvFile string `json:"env_file"`
	Note    string `json:"note,omitempty"`
}

// detectStack looks at the manifest files in dir. It returns nil when it
// finds no project it knows.
func detectStack(dir string) *stack {
	for _, detect := range []func(string) *stack{detectNode, detectPython, detectRuby, detectPHP, detectGo} {
		if s := detect(dir); s != nil {
			return s
		}
	}
	return nil
}

func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func readText(dir string, names ...string) string {
	var b strings.Builder
	for _, name := range names {
		if data, err := os.ReadFile(filepath.Join(dir, name)); err == nil {
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

const nodeSDK = "@moderation-api/sdk"

func detectNode(dir string) *stack {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil
	}
	var manifest struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	_ = json.Unmarshal(data, &manifest)
	has := func(name string) bool {
		_, dep := manifest.Dependencies[name]
		_, dev := manifest.DevDependencies[name]
		return dep || dev
	}
	hasPrefix := func(prefix string) bool {
		for name := range manifest.Dependencies {
			if strings.HasPrefix(name, prefix) {
				return true
			}
		}
		return false
	}

	s := &stack{Language: "javascript", SDKPackage: nodeSDK, SDKInstalled: has(nodeSDK), EnvFile: ".env"}
	if exists(dir, "tsconfig.json") || has("typescript") {
		s.Language = "typescript"
	}
	switch {
	case has("next"):
		s.Framework, s.EnvFile = "Next.js", ".env.local"
	case hasPrefix("@remix-run/") || has("react-router"):
		s.Framework = "Remix / React Router"
	case has("nuxt"):
		s.Framework = "Nuxt"
	case has("@sveltejs/kit"):
		s.Framework = "SvelteKit"
	case has("astro"):
		s.Framework = "Astro"
	case has("@nestjs/core"):
		s.Framework = "NestJS"
	case has("hono"):
		s.Framework = "Hono"
	case has("fastify"):
		s.Framework = "Fastify"
	case has("express"):
		s.Framework = "Express"
	case has("koa"):
		s.Framework = "Koa"
	}

	switch {
	case exists(dir, "pnpm-lock.yaml"):
		s.PackageManager, s.InstallCommand = "pnpm", "pnpm add "+nodeSDK
	case exists(dir, "yarn.lock"):
		s.PackageManager, s.InstallCommand = "yarn", "yarn add "+nodeSDK
	case exists(dir, "bun.lock") || exists(dir, "bun.lockb"):
		s.PackageManager, s.InstallCommand = "bun", "bun add "+nodeSDK
	default:
		s.PackageManager, s.InstallCommand = "npm", "npm install "+nodeSDK
	}
	return s
}

const pythonSDK = "moderation_api"

func detectPython(dir string) *stack {
	manifests := []string{"pyproject.toml", "requirements.txt", "Pipfile", "setup.py"}
	found := false
	for _, name := range manifests {
		found = found || exists(dir, name)
	}
	if !found {
		return nil
	}
	text := strings.ToLower(readText(dir, manifests...))
	s := &stack{
		Language:     "python",
		SDKPackage:   pythonSDK,
		SDKInstalled: strings.Contains(text, "moderation_api") || strings.Contains(text, "moderation-api"),
		EnvFile:      ".env",
	}
	switch {
	case strings.Contains(text, "django"):
		s.Framework = "Django"
	case strings.Contains(text, "fastapi"):
		s.Framework = "FastAPI"
	case strings.Contains(text, "flask"):
		s.Framework = "Flask"
	}
	switch {
	case exists(dir, "uv.lock"):
		s.PackageManager, s.InstallCommand = "uv", "uv add "+pythonSDK
	case exists(dir, "poetry.lock"):
		s.PackageManager, s.InstallCommand = "poetry", "poetry add "+pythonSDK
	case exists(dir, "Pipfile"):
		s.PackageManager, s.InstallCommand = "pipenv", "pipenv install "+pythonSDK
	default:
		s.PackageManager, s.InstallCommand = "pip", "pip install "+pythonSDK
		if exists(dir, "requirements.txt") {
			s.Note = "Also add moderation_api to requirements.txt."
		}
	}
	return s
}

func detectRuby(dir string) *stack {
	if !exists(dir, "Gemfile") {
		return nil
	}
	text := readText(dir, "Gemfile")
	s := &stack{
		Language:       "ruby",
		PackageManager: "bundler",
		SDKPackage:     "moderation_api",
		SDKInstalled:   strings.Contains(text, "moderation_api"),
		InstallCommand: "bundle add moderation_api",
		EnvFile:        ".env",
	}
	if strings.Contains(text, "rails") {
		s.Framework = "Rails"
	}
	return s
}

const phpSDK = "moderation-api/sdk-php"

func detectPHP(dir string) *stack {
	if !exists(dir, "composer.json") {
		return nil
	}
	text := readText(dir, "composer.json")
	s := &stack{
		Language:       "php",
		PackageManager: "composer",
		SDKPackage:     phpSDK,
		SDKInstalled:   strings.Contains(text, phpSDK),
		InstallCommand: "composer require " + phpSDK,
		EnvFile:        ".env",
	}
	switch {
	case strings.Contains(text, "laravel/framework"):
		s.Framework = "Laravel"
	case strings.Contains(text, "symfony/"):
		s.Framework = "Symfony"
	}
	return s
}

func detectGo(dir string) *stack {
	if !exists(dir, "go.mod") {
		return nil
	}
	return &stack{
		Language: "go",
		EnvFile:  ".env",
		Note:     "There is no Go SDK yet. Call the REST API (https://docs.moderationapi.com/api-reference/introduction) with the key in MODAPI_SECRET_KEY.",
	}
}
