package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// runDoctor diagnoses dependencies, installation status, and runtime environment.
func runDoctor(args []string) {
	fmt.Println("CodeDNA Environment Diagnostics (Doctor)")
	fmt.Println("==================================================")

	var warnings []string
	var remedies []string

	// 1. OS & Architecture
	osName := runtime.GOOS
	archName := runtime.GOARCH
	fmt.Printf("[OK] Platform:              %s (%s)\n", osName, archName)

	// 2. Binary Installation
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("[FAIL] Binary Location:     Unable to determine (%v)\n", err)
		warnings = append(warnings, "Cannot determine executable path")
	} else {
		resolvedPath, _ := filepath.EvalSymlinks(execPath)
		fmt.Printf("[OK] CodeDNA Binary:        %s (v%s)\n", resolvedPath, version)
	}

	// 3. Shell PATH check
	pathEnv := os.Getenv("PATH")
	inPath := false
	if execPath != "" {
		execDir := filepath.Dir(execPath)
		for _, p := range filepath.SplitList(pathEnv) {
			if filepath.Clean(p) == filepath.Clean(execDir) {
				inPath = true
				break
			}
		}
	}
	home, _ := os.UserHomeDir()
	localBin := filepath.Join(home, ".local", "bin")
	hasLocalBin := false
	for _, p := range filepath.SplitList(pathEnv) {
		if filepath.Clean(p) == filepath.Clean(localBin) {
			hasLocalBin = true
			break
		}
	}

	if inPath || hasLocalBin {
		fmt.Printf("[OK] Shell PATH:            Configured correctly\n")
	} else {
		fmt.Printf("[WARN] Shell PATH:          ~/.local/bin is missing from PATH\n")
		warnings = append(warnings, "User binary directory is not in PATH")
		remedies = append(remedies, fmt.Sprintf("Run 'codedna install' or add: export PATH=\"%s:$PATH\" to your ~/.bashrc or ~/.zshrc", localBin))
	}

	// 4. Local Storage
	codednaHome := filepath.Join(home, ".codedna")
	if h := os.Getenv("CODEDNA_HOME"); h != "" {
		codednaHome = h
	}
	if err := os.MkdirAll(codednaHome, 0o755); err != nil {
		fmt.Printf("[FAIL] Local Storage:       %s (Not writable: %v)\n", codednaHome, err)
		warnings = append(warnings, "Local storage directory is not writable")
	} else {
		fmt.Printf("[OK] Local Storage:         %s (Writable)\n", codednaHome)
	}

	// 5. Web Studio Assets
	webAssetFound := false
	webDirFound := ""
	searchDirs := []string{
		"web/out",
		"../web/out",
		filepath.Join(codednaHome, "web", "out"),
		filepath.Join(codednaHome, "repo", "web", "out"),
		"/usr/local/share/codedna/web/out",
	}
	if envDir := os.Getenv("CODEDNA_WEB_DIR"); envDir != "" {
		searchDirs = append([]string{envDir}, searchDirs...)
	}
	if execPath != "" {
		execDir := filepath.Dir(execPath)
		searchDirs = append(searchDirs,
			filepath.Join(execDir, "web", "out"),
			filepath.Join(execDir, "..", "web", "out"),
			filepath.Join(execDir, "..", "share", "codedna", "web", "out"),
		)
	}
	for _, d := range searchDirs {
		if fi, err := os.Stat(filepath.Join(d, "index.html")); err == nil && !fi.IsDir() {
			webAssetFound = true
			webDirFound = d
			break
		}
	}
	if webAssetFound {
		fmt.Printf("[OK] Web Studio Assets:     %s\n", webDirFound)
	} else {
		fmt.Printf("[WARN] Web Studio Assets:   Not found in standard search paths\n")
		warnings = append(warnings, "Web dashboard static assets not found")
		remedies = append(remedies, "Download release archive or run: 'curl -fsSL https://raw.githubusercontent.com/0xzahed/codedna/main/install.sh | bash'")
	}

	// 6. Port 8080 Availability
	l, err := net.Listen("tcp", "127.0.0.1:8080")
	if err == nil {
		l.Close()
		fmt.Printf("[OK] Default Port 8080:     Available on 127.0.0.1\n")
	} else {
		fmt.Printf("[WARN] Default Port 8080:   Currently occupied (%v)\n", err)
		warnings = append(warnings, "Port 8080 is in use")
		remedies = append(remedies, "Use a custom port with 'codedna serve --port 8081'")
	}

	// 7. Current Project Directory
	cwd, err := os.Getwd()
	if err == nil {
		if home != "" && filepath.Clean(cwd) == filepath.Clean(home) {
			fmt.Printf("[WARN] Working Directory:   User home directory (%s)\n", cwd)
			warnings = append(warnings, "Current directory is user home directory")
			remedies = append(remedies, "Navigate into a specific code project directory: 'cd /path/to/my-project'")
		} else {
			fmt.Printf("[OK] Working Directory:     %s (Ready)\n", cwd)
		}
	}

	// 8. Toolchain Checks (Informational)
	if out, err := exec.Command("git", "--version").Output(); err == nil {
		fmt.Printf("[--] Git CLI:               %s", string(out))
	} else {
		fmt.Printf("[--] Git CLI:               Not installed (optional)\n")
	}

	if out, err := exec.Command("go", "version").Output(); err == nil {
		fmt.Printf("[--] Go Toolchain:          %s", string(out))
	} else {
		fmt.Printf("[--] Go Toolchain:          Not installed (optional, prebuilt binaries used)\n")
	}

	if out, err := exec.Command("node", "--version").Output(); err == nil {
		fmt.Printf("[--] Node.js Runtime:       %s", string(out))
	} else {
		fmt.Printf("[--] Node.js Runtime:       Not installed (optional, prebuilt UI used)\n")
	}

	if err := exec.Command("docker", "info").Run(); err == nil {
		fmt.Printf("[--] Docker Runtime:        Installed & Daemon active (optional)\n")
	} else {
		fmt.Printf("[--] Docker Runtime:        Not running or not installed (optional)\n")
	}

	fmt.Println("==================================================")
	if len(warnings) == 0 {
		fmt.Println("Status: All checks passed. Ready to analyze any project!")
		fmt.Println("Run: 'codedna serve' in any project folder.")
	} else {
		fmt.Println("Actionable Remediation:")
		for i, r := range remedies {
			fmt.Printf("  %d. %s\n", i+1, r)
		}
	}
	fmt.Println()
}
