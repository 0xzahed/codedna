#!/usr/bin/env bash
# ==============================================================================
# CodeDNA — Universal Zero-Dependency Installer (Linux & macOS)
# "Turn Any Codebase Into a Living Architecture Map"
#
# Primary Objective: Native installation without requiring Docker or Git cloning.
# Priority:
#   1. Prebuilt binary from GitHub Releases (with SHA256 verification)
#   2. Native Go source compilation (when Go is available)
#   3. Optional Docker runtime runner (if native is unavailable)
#   4. Actionable error with remediation steps
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/0xzahed/codedna/main/install.sh | bash
#   or: bash install.sh [--uninstall]
#
# Repository: https://github.com/0xzahed/codedna
# Author: Abu Zahed <zahed04x@gmail.com>
# ==============================================================================

set -euo pipefail

# ANSI Styling
C_RESET="\033[0m"
C_BOLD="\033[1m"
C_CYAN="\033[36m"
C_GREEN="\033[32m"
C_YELLOW="\033[33m"
C_RED="\033[31m"
C_DIM="\033[2m"

info()    { printf "${C_CYAN}[CODEDNA]${C_RESET} %s\n" "$1"; }
success() { printf "${C_GREEN}[✓]${C_RESET} %s\n" "$1"; }
warn()    { printf "${C_YELLOW}[!]${C_RESET} %s\n" "$1"; }
error()   { printf "${C_RED}[✗] ERROR:${C_RESET} %s\n" "$1" >&2; exit 1; }

REPO_OWNER="0xzahed"
REPO_NAME="codedna"
GITHUB_REPO="https://github.com/${REPO_OWNER}/${REPO_NAME}"
RAW_BASE="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/main"

print_banner() {
    printf "${C_CYAN}${C_BOLD}"
    cat << "EOF"
   ____          _      ____  _   _    _      ____     ___  
  / ___|___   __| | ___|  _ \| \ | |  / \    |___ \   / _ \ 
 | |   / _ \ / _| |/ _ \ | | |  \| | / _ \     __) | | | | |
 | |__| (_) | (_| |  __/ |_| | |\  |/ ___ \   / __/ _| |_| |
  \____\___/ \__,_|\___|____/|_| \_/_/   \_\ |_____(_)\___/ 
EOF
    printf "${C_RESET}\n"
    printf "  ${C_BOLD}CodeDNA — Universal Software Architecture Engine${C_RESET}\n"
    printf "  ${C_DIM}Zero-Dependency Native Installation for Linux & macOS${C_RESET}\n"
    printf "====================================================================\n\n"
}

# Directories
CODEDNA_HOME="${HOME}/.codedna"
BIN_DIR="${CODEDNA_HOME}/bin"
WEB_DIR="${CODEDNA_HOME}/web/out"
LOCAL_BIN="${HOME}/.local/bin"
TARGET_LINK="${LOCAL_BIN}/codedna"
GLOBAL_LINK="/usr/local/bin/codedna"

# Uninstallation handler
if [ "${1:-}" = "--uninstall" ] || [ "${1:-}" = "uninstall" ]; then
    printf "${C_YELLOW}Uninstalling CodeDNA...${C_RESET}\n"
    rm -rf "${CODEDNA_HOME}"
    rm -f "${TARGET_LINK}"
    if [ -w "/usr/local/bin" ]; then
        rm -f "${GLOBAL_LINK}"
    fi
    success "CodeDNA has been completely removed from your system."
    exit 0
fi

print_banner

# 1. Platform Detection
OS_RAW="$(uname -s)"
ARCH_RAW="$(uname -m)"

case "$OS_RAW" in
    Linux)
        OS="linux"
        ;;
    Darwin)
        OS="darwin"
        ;;
    *)
        error "Unsupported operating system: $OS_RAW. Windows users should run install.ps1 in PowerShell."
        ;;
esac

case "$ARCH_RAW" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        error "Unsupported CPU architecture: $ARCH_RAW."
        ;;
esac

info "Target platform detected: ${OS}/${ARCH}"

# Ensure base directories exist
mkdir -p "${BIN_DIR}"
mkdir -p "${WEB_DIR}"
mkdir -p "${LOCAL_BIN}"

# Configure PATH safely in shell RC files
setup_path() {
    local target_dir="$1"
    if [[ ":$PATH:" != *":${target_dir}:"* ]]; then
        for rc in "${HOME}/.bashrc" "${HOME}/.zshrc" "${HOME}/.profile" "${HOME}/.bash_profile"; do
            if [ -f "$rc" ]; then
                if ! grep -q 'export PATH=.*\.local/bin' "$rc" 2>/dev/null; then
                    printf '\n# CodeDNA CLI path\nexport PATH="%s:$PATH"\n' "${target_dir}" >> "$rc"
                    info "Appended ${target_dir} to ${rc}"
                fi
            fi
        done
        export PATH="${target_dir}:$PATH"
    fi
}

install_symlink() {
    local source_binary="$1"
    local installed=false

    # Primary: User-writable ~/.local/bin (No sudo required)
    ln -sf "${source_binary}" "${TARGET_LINK}"
    chmod +x "${TARGET_LINK}"
    setup_path "${LOCAL_BIN}"
    success "Linked executable to ${TARGET_LINK}"
    installed=true

    # Secondary optional: /usr/local/bin if writable
    if [ -w "/usr/local/bin" ]; then
        ln -sf "${source_binary}" "${GLOBAL_LINK}" 2>/dev/null || true
    fi

    if [ "$installed" = false ]; then
        error "Unable to create binary symlink in ${TARGET_LINK}."
    fi
}

# ------------------------------------------------------------------------------
# Priority 1: Download Official Prebuilt Release Binary from GitHub
# ------------------------------------------------------------------------------
try_prebuilt_download() {
    info "Priority 1: Checking for official prebuilt release binary..."
    
    local tmp_dir
    tmp_dir="$(mktemp -d -t codedna-install-XXXXXX)"
    trap 'rm -rf "${tmp_dir}"' RETURN

    local release_json="${tmp_dir}/release.json"
    local api_url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"

    # Fetch release metadata via curl
    if ! curl -fsSL -H "Accept: application/vnd.github.v3+json" "${api_url}" -o "${release_json}" 2>/dev/null; then
        warn "Unable to fetch release metadata from GitHub API (might be offline or rate-limited)."
        return 1
    fi

    local tag_name
    tag_name="$(grep -o '"tag_name": *"[^"]*"' "${release_json}" | head -n 1 | cut -d '"' -f 4 || true)"
    if [ -z "${tag_name}" ]; then
        warn "No stable release tag found."
        return 1
    fi

    info "Found latest release: ${tag_name}"

    local archive_name="codedna-${tag_name}-${OS}-${ARCH}.tar.gz"
    local fallback_name="codedna-${OS}-${ARCH}.tar.gz"
    local download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${tag_name}/${archive_name}"
    local fallback_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${tag_name}/${fallback_name}"
    local archive_file="${tmp_dir}/${archive_name}"

    info "Downloading ${archive_name}..."
    if ! curl -fsSL "${download_url}" -o "${archive_file}" 2>/dev/null; then
        if ! curl -fsSL "${fallback_url}" -o "${archive_file}" 2>/dev/null; then
            warn "Prebuilt archive not found for ${OS}/${ARCH} in release ${tag_name}."
            return 1
        fi
    fi

    # Checksum verification (SHA256)
    local checksum_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${tag_name}/SHA256SUMS"
    local checksum_file="${tmp_dir}/SHA256SUMS"
    if curl -fsSL "${checksum_url}" -o "${checksum_file}" 2>/dev/null; then
        info "Verifying SHA256 checksum..."
        local expected_sha
        expected_sha="$(grep -F "${archive_name}" "${checksum_file}" | cut -d ' ' -f 1 || true)"
        if [ -n "${expected_sha}" ]; then
            local actual_sha=""
            if command -v sha256sum >/dev/null 2>&1; then
                actual_sha="$(sha256sum "${archive_file}" | cut -d ' ' -f 1)"
            elif command -v shasum >/dev/null 2>&1; then
                actual_sha="$(shasum -a 256 "${archive_file}" | cut -d ' ' -f 1)"
            fi

            if [ -n "${actual_sha}" ]; then
                if [ "${expected_sha}" = "${actual_sha}" ]; then
                    success "SHA256 verified successfully (${actual_sha:0:16}...)"
                else
                    error "Checksum verification failed! Expected: ${expected_sha}, Got: ${actual_sha}"
                fi
            fi
        fi
    fi

    # Extract archive
    info "Extracting release bundle..."
    tar -xzf "${archive_file}" -C "${tmp_dir}"

    local extracted_bin=""
    for cand in "${tmp_dir}/codedna" "${tmp_dir}/bin/codedna"; do
        if [ -f "$cand" ]; then
            extracted_bin="$cand"
            break
        fi
    done

    if [ -z "${extracted_bin}" ]; then
        warn "Extracted archive did not contain 'codedna' binary."
        return 1
    fi

    # Install binary
    cp -f "${extracted_bin}" "${BIN_DIR}/codedna"
    chmod +x "${BIN_DIR}/codedna"

    # Install web assets if bundled in release archive
    if [ -d "${tmp_dir}/web/out" ]; then
        cp -rf "${tmp_dir}/web/out/"* "${WEB_DIR}/"
        info "Extracted web studio UI assets to ${WEB_DIR}"
    fi

    install_symlink "${BIN_DIR}/codedna"
    return 0
}

# ------------------------------------------------------------------------------
# Priority 2: Native Go Compilation from Source
# ------------------------------------------------------------------------------
try_source_compilation() {
    info "Priority 2: Checking Go compiler availability for native source build..."

    if ! command -v go >/dev/null 2>&1; then
        warn "Go compiler is not installed."
        return 1
    fi

    local go_version
    go_version="$(go version | awk '{print $3}')"
    info "Found Go compiler: ${go_version}"

    local src_dir="${CODEDNA_HOME}/src"
    mkdir -p "${src_dir}"

    if [ -d "${src_dir}/.git" ]; then
        info "Updating existing CodeDNA source cache..."
        git -C "${src_dir}" pull --quiet origin main || true
    elif command -v git >/dev/null 2>&1; then
        info "Fetching CodeDNA source tree..."
        rm -rf "${src_dir}"
        git clone --depth=1 --quiet "${GITHUB_REPO}.git" "${src_dir}"
    else
        info "Downloading source tarball..."
        curl -fsSL "${GITHUB_REPO}/archive/refs/heads/main.tar.gz" | tar -xz -C "${src_dir}" --strip-components=1
    fi

    # Check for frontend build
    if [ ! -d "${src_dir}/web/out" ] && command -v npm >/dev/null 2>&1; then
        info "Building Next.js architecture studio UI..."
        (cd "${src_dir}/web" && npm ci --silent && npm run build --silent) || warn "Web build had non-fatal notices."
    fi

    info "Compiling CodeDNA binary from source..."
    (
        cd "${src_dir}"
        go build -ldflags="-s -w" -o "${BIN_DIR}/codedna" ./cmd/codedna
    )

    if [ -d "${src_dir}/web/out" ]; then
        cp -rf "${src_dir}/web/out/"* "${WEB_DIR}/" 2>/dev/null || true
    fi

    install_symlink "${BIN_DIR}/codedna"
    return 0
}

# ------------------------------------------------------------------------------
# Priority 3: Optional Docker Runtime Setup (Explicit & Optional)
# ------------------------------------------------------------------------------
offer_docker_fallback() {
    info "Priority 3: Checking optional Docker runtime availability..."

    if ! command -v docker >/dev/null 2>&1; then
        warn "Docker is not installed."
        return 1
    fi

    if ! docker info >/dev/null 2>&1; then
        warn "Docker daemon is not running."
        return 1
    fi

    info "Docker is available on this system."
    warn "Native installation could not be completed, but Docker can be configured as a fallback."

    # Create standalone CLI wrapper invoking Docker
    local wrapper="${BIN_DIR}/codedna"
    cat << 'WRAPPER_EOF' > "${wrapper}"
#!/usr/bin/env bash
# CodeDNA Containerized CLI Runner
IMAGE="${CODEDNA_IMAGE:-ghcr.io/0xzahed/codedna:latest}"

TARGET="${PWD}"
for ((i=1; i<=$#; i++)); do
    if [ "${!i}" = "--repo" ]; then
        next_i=$((i+1))
        TARGET="${!next_i}"
        break
    fi
done

exec docker run --rm -it \
    --name codedna-server \
    -p 8080:8080 \
    -v "${TARGET}:/workspace" \
    -e CODEDNA_HOME=/root/.codedna \
    "${IMAGE}" "$@"
WRAPPER_EOF

    chmod +x "${wrapper}"
    install_symlink "${wrapper}"
    info "Pre-fetching latest CodeDNA Docker image..."
    docker pull ghcr.io/0xzahed/codedna:latest || true
    return 0
}

# ------------------------------------------------------------------------------
# Execution Pipeline
# ------------------------------------------------------------------------------
INSTALLED=false

if try_prebuilt_download; then
    INSTALLED=true
elif try_source_compilation; then
    INSTALLED=true
elif offer_docker_fallback; then
    INSTALLED=true
fi

if [ "$INSTALLED" = false ]; then
    printf "\n"
    error "Installation failed. Neither prebuilt binary, Go compiler (https://go.dev), nor Docker could be utilized.
Remediation:
  1. Install Go 1.21+ on your machine: https://go.dev/dl/
  2. Or install Docker: https://docs.docker.com/get-docker/
  3. Or download release binaries directly: https://github.com/0xzahed/codedna/releases"
fi

# Verification
printf "\n====================================================================\n"
success "CodeDNA installed successfully!"
printf "====================================================================\n\n"

# Verify execution
if command -v codedna >/dev/null 2>&1; then
    codedna version || true
else
    "${BIN_DIR}/codedna" version || true
fi

printf "\n${C_BOLD}QUICK START (Native — No Docker / No Git Clone):${C_RESET}\n"
printf "  1. Navigate to ANY code project folder:\n"
printf "     ${C_CYAN}cd /path/to/any-project${C_RESET}\n\n"
printf "  2. Launch CodeDNA Architecture Studio:\n"
printf "     ${C_CYAN}codedna serve${C_RESET}\n\n"
printf "  3. Open the live architecture dashboard:\n"
printf "     ${C_GREEN}http://127.0.0.1:8080${C_RESET}\n\n"
printf "  Tip: Run diagnostics anytime with: ${C_CYAN}codedna doctor${C_RESET}\n"
printf "  Tip: To run in Docker container:     ${C_CYAN}codedna docker${C_RESET}\n"
printf "====================================================================\n\n"
