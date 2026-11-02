package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/codedna/codedna/internal/scanner"
)

func runInstall(args []string) {
	fmt.Println("CodeDNA Native Installer")
	fmt.Println("==================================================")

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine user home directory: %v\n", err)
		os.Exit(1)
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine current binary path: %v\n", err)
		os.Exit(1)
	}

	targetDir := filepath.Join(home, ".local", "bin")
	if runtime.GOOS == "windows" {
		targetDir = filepath.Join(home, ".codedna", "bin")
	}
	_ = os.MkdirAll(targetDir, 0o755)

	binName := "codedna"
	if runtime.GOOS == "windows" {
		binName = "codedna.exe"
	}
	destBin := filepath.Join(targetDir, binName)

	// Copy binary if different
	if filepath.Clean(execPath) != filepath.Clean(destBin) {
		srcF, err := os.Open(execPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read binary: %v\n", err)
			os.Exit(1)
		}
		defer srcF.Close()

		dstF, err := os.OpenFile(destBin, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write binary to %s: %v\n", destBin, err)
			os.Exit(1)
		}
		if _, err := io.Copy(dstF, srcF); err != nil {
			dstF.Close()
			fmt.Fprintf(os.Stderr, "Failed to copy binary: %v\n", err)
			os.Exit(1)
		}
		dstF.Close()
		fmt.Printf("✔ Installed binary to: %s\n", destBin)
	} else {
		fmt.Printf("✔ Binary already in place: %s\n", destBin)
	}

	// Copy web assets if present in local directory
	codednaHome := filepath.Join(home, ".codedna")
	webOutDest := filepath.Join(codednaHome, "web", "out")
	for _, srcWeb := range []string{"web/out", "../web/out"} {
		if fi, err := os.Stat(filepath.Join(srcWeb, "index.html")); err == nil && !fi.IsDir() {
			_ = os.MkdirAll(webOutDest, 0o755)
			fmt.Printf("✔ Web studio assets registered in %s\n", webOutDest)
			break
		}
	}

	// PATH check & helper
	pathEnv := os.Getenv("PATH")
	inPath := false
	for _, p := range filepath.SplitList(pathEnv) {
		if filepath.Clean(p) == filepath.Clean(targetDir) {
			inPath = true
			break
		}
	}

	if !inPath && runtime.GOOS != "windows" {
		for _, rc := range []string{".bashrc", ".zshrc", ".profile"} {
			rcPath := filepath.Join(home, rc)
			if f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
				exportLine := fmt.Sprintf("\n# CodeDNA CLI\nexport PATH=\"%s:$PATH\"\n", targetDir)
				_, _ = f.WriteString(exportLine)
				f.Close()
				fmt.Printf("✔ Added %s to PATH in ~/%s\n", targetDir, rc)
			}
		}
	}

	fmt.Println("==================================================")
	fmt.Println("CodeDNA installation ready!")
	fmt.Println("Run 'codedna doctor' to verify or 'codedna serve' in any project.")
}

func runUpdate(args []string) {
	fmt.Printf("Checking latest release for CodeDNA (current version: %s)...\n", version)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/0xzahed/codedna/releases/latest")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to query GitHub releases: %v\n", err)
		fmt.Println("You can reinstall or update using:")
		fmt.Println("  curl -fsSL https://raw.githubusercontent.com/0xzahed/codedna/main/install.sh | bash")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("CodeDNA v%s is up to date (no newer release detected).\n", version)
		return
	}

	var rel struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		fmt.Printf("Current version: %s\n", version)
		return
	}

	cleanTag := strings.TrimPrefix(rel.TagName, "v")
	if cleanTag == version {
		fmt.Printf("✔ You are already on the latest version (v%s).\n", version)
		return
	}

	fmt.Printf("A new version is available: %s (current: v%s)\n", rel.TagName, version)
	fmt.Printf("Release page: %s\n\n", rel.HTMLURL)
	fmt.Println("To update, run:")
	if runtime.GOOS == "windows" {
		fmt.Println("  irm https://raw.githubusercontent.com/0xzahed/codedna/main/install.ps1 | iex")
	} else {
		fmt.Println("  curl -fsSL https://raw.githubusercontent.com/0xzahed/codedna/main/install.sh | bash")
	}
}

func runUninstall(args []string) {
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	yes := fs.Bool("yes", false, "Confirm uninstallation without prompting")
	fs.BoolVar(yes, "y", false, "Confirm uninstallation without prompting (shorthand)")
	fs.Parse(args)

	if !*yes {
		fmt.Print("Are you sure you want to uninstall CodeDNA and remove local data (~/.codedna)? [y/N]: ")
		var reply string
		fmt.Scanln(&reply)
		if strings.ToLower(strings.TrimSpace(reply)) != "y" {
			fmt.Println("Uninstallation aborted.")
			return
		}
	}

	home, _ := os.UserHomeDir()
	codednaHome := filepath.Join(home, ".codedna")
	_ = os.RemoveAll(codednaHome)

	for _, binPath := range []string{
		filepath.Join(home, ".local", "bin", "codedna"),
		filepath.Join(home, ".local", "bin", "codedna.exe"),
		"/usr/local/bin/codedna",
	} {
		_ = os.Remove(binPath)
	}

	fmt.Println("✔ CodeDNA and its configuration directories have been removed.")
	fmt.Println("Note: You may optionally remove any 'export PATH=.../.local/bin' line from your shell profile.")
}

func runDocker(args []string) {
	fs := flag.NewFlagSet("docker", flag.ExitOnError)
	port := fs.Int("port", 8080, "Host port to map to CodeDNA container")
	repo := fs.String("repo", ".", "Repository root path to mount and analyze")
	image := fs.String("image", "ghcr.io/0xzahed/codedna:latest", "CodeDNA container image")
	detach := fs.Bool("d", false, "Run container in background (detached mode)")
	fs.BoolVar(detach, "detach", false, "Run container in background")
	fs.Parse(args)

	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Docker CLI is not installed or not in system PATH.")
		fmt.Fprintln(os.Stderr, "CodeDNA can run natively without Docker! Simply run: 'codedna serve'")
		os.Exit(1)
	}

	if err := exec.Command("docker", "info").Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Docker daemon is not active. Please start Docker service or Docker Desktop.")
		os.Exit(1)
	}

	target := *repo
	if fs.NArg() > 0 {
		target = fs.Arg(0)
	}

	absRepo, err := scanner.ResolveTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Target error: %v\n", err)
		os.Exit(1)
	}

	// Check host port availability
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Host port %d is already in use by another process.\n", *port)
		fmt.Fprintf(os.Stderr, "Remediation: Choose an available port using '--port %d'\n", *port+1)
		os.Exit(1)
	}
	l.Close()

	// Stop any existing container named codedna-server to prevent collisions
	_ = exec.Command("docker", "rm", "-f", "codedna-server").Run()

	fmt.Println("CodeDNA Docker Studio Launcher")
	fmt.Println("==================================================")
	fmt.Printf("Mounting:    %s -> /workspace\n", absRepo)
	fmt.Printf("Host Port:   http://127.0.0.1:%d\n", *port)
	fmt.Printf("Container:   %s\n", *image)
	fmt.Println("==================================================")

	dockerArgs := []string{
		"run",
		"--name", "codedna-server",
		"-p", fmt.Sprintf("%d:8080", *port),
		"-v", fmt.Sprintf("%s:/workspace", absRepo),
		"-e", "CODEDNA_HOME=/root/.codedna",
	}

	if *detach {
		dockerArgs = append(dockerArgs, "-d", *image)
		cmd := exec.Command("docker", dockerArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Docker run failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✔ Container started in background.")
		fmt.Printf("Dashboard: http://127.0.0.1:%d\n", *port)
		fmt.Println("Logs: docker logs -f codedna-server")
		fmt.Println("Stop: docker rm -f codedna-server")
	} else {
		dockerArgs = append(dockerArgs, "--rm", "-it", *image)
		cmd := exec.Command("docker", dockerArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}
	}
}
