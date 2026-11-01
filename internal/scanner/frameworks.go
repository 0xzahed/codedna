package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// detectFrameworks infers frameworks from manifests and code layout.
// Signals are evidence-based; unknown projects get an empty list.
func detectFrameworks(root string, cfg map[string]string, langs map[string]int) []string {
	out := []string{}
	add := func(name string) {
		for _, x := range out {
			if x == name {
				return
			}
		}
		out = append(out, name)
	}

	pkgJSON := map[string]interface{}{}
	if p, ok := cfg["package.json"]; ok {
		if data, err := os.ReadFile(filepath.Join(root, p)); err == nil {
			if err := jsonUnmarshal(data, &pkgJSON); err == nil {
				deps := collectDepsStrings(pkgJSON)
				for _, fw := range []struct{ fw, pkg string }{
					{"Next.js", "next"}, {"React", "react"}, {"Vue", "vue"},
					{"Svelte", "svelte"}, {"Angular", "@angular/core"},
					{"Express", "express"}, {"Fastify", "fastify"}, {"NestJS", "@nestjs/core"},
					{"React Native", "react-native"}, {"Electron", "electron"},
				} {
					if hasDep(deps, fw.pkg) {
						add(fw.fw)
					}
				}
			}
		}
	}
	// Next.js / Nuxt app layout signatures even without manifest.
	if exists(filepath.Join(root, "next.config.js")) || exists(filepath.Join(root, "next.config.mjs")) || exists(filepath.Join(root, "next.config.ts")) {
		add("Next.js")
	}
	if exists(filepath.Join(root, "nuxt.config.ts")) {
		add("Nuxt")
	}

	if _, ok := cfg["go.mod"]; ok {
		data, err := os.ReadFile(filepath.Join(root, cfg["go.mod"]))
		if err == nil {
			s := string(data)
			if strings.Contains(s, "github.com/gin-gonic/gin") {
				add("Gin")
			}
			if strings.Contains(s, "github.com/gofiber/fiber") {
				add("Fiber")
			}
			if strings.Contains(s, "github.com/labstack/echo") {
				add("Echo")
			}
			if strings.Contains(s, "github.com/gorilla/mux") {
				add("Gorilla Mux")
			}
		}
	}

	if _, ok := cfg["pyproject.toml"]; ok {
		data, err := os.ReadFile(filepath.Join(root, cfg["pyproject.toml"]))
		if err == nil {
			s := string(data)
			for fw, marker := range map[string]string{
				"Django": "django", "FastAPI": "fastapi", "Flask": "flask",
				"Typer": "typer", "Click": "click",
			} {
				if strings.Contains(s, marker) {
					add(fw)
				}
			}
		}
	}
	if _, ok := cfg["requirements.txt"]; ok {
		data, err := os.ReadFile(filepath.Join(root, cfg["requirements.txt"]))
		if err == nil {
			s := strings.ToLower(string(data))
			for fw, marker := range map[string]string{
				"Django": "django", "FastAPI": "fastapi", "Flask": "flask", "Pytest": "pytest",
			} {
				if strings.Contains(s, marker) {
					add(fw)
				}
			}
		}
	}

	if _, ok := cfg["composer.json"]; ok {
		data, err := os.ReadFile(filepath.Join(root, cfg["composer.json"]))
		if err == nil {
			s := string(data)
			for fw, marker := range map[string]string{
				"Laravel": "laravel/framework", "Symfony": "symfony/framework-bundle", "WordPress": "wordpress",
			} {
				if strings.Contains(s, marker) {
					add(fw)
				}
			}
		}
	}

	if _, ok := cfg["pom.xml"]; ok {
		data, err := os.ReadFile(filepath.Join(root, cfg["pom.xml"]))
		if err == nil {
			s := string(data)
			for fw, marker := range map[string]string{
				"Spring Boot": "spring-boot-starter", "Quarkus": "quarkus", "Micronaut": "micronaut",
			} {
				if strings.Contains(s, marker) {
					add(fw)
				}
			}
		}
	}

	// Django layout signature.
	if langs["python"] > 0 && exists(filepath.Join(root, "manage.py")) {
		add("Django")
	}

	return out
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return jsonDecode(data, v)
}

func jsonDecode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func collectDepsStrings(pkg map[string]interface{}) []string {
	var names []string
	for _, section := range []string{"dependencies", "devDependencies", "peerDependencies"} {
		m, ok := pkg[section].(map[string]interface{})
		if !ok {
			continue
		}
		for k := range m {
			names = append(names, k)
		}
	}
	return names
}

func hasDep(deps []string, pkg string) bool {
	for _, d := range deps {
		if d == pkg {
			return true
		}
	}
	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
