#!/usr/bin/env bash
# ==============================================================================
# CodeDNA Installer
# "Turn Any Codebase Into a Living Architecture Map"
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/0xzahed/codedna/main/install.sh | bash
#
# Author: Abu Zahed <zahed04x@gmail.com>
# Repository: https://github.com/0xzahed/codedna
# ==============================================================================

set -euo pipefail

# Colors
C_RESET="\033[0m"
C_BOLD="\033[1m"
C_CYAN="\033[36m"
C_GREEN="\033[32m"
C_YELLOW="\033[33m"
C_RED="\033[31m"
C_DIM="\033[2m"

info() {
    printf "${C_CYAN}[CODEDNA]${C_RESET} %s\n" "$1"
}

success() {
    printf "${C_GREEN}[✓]${C_RESET} %s\n" "$1"
}

warn() {
    printf "${C_YELLOW}[!]${C_RESET} %s\n" "$1"
}

error() {
    printf "${C_RED}[✗] ERROR:${C_RESET} %s\n" "$1" >&2
    exit 1
}

# Banner
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
    printf "  ${C_BOLD}Universal Software Architecture Reverse Engineering Engine${C_RESET}\n"
    printf "  ${C_DIM}Zero-Clone Standalone Setup for Linux & macOS${C_RESET}\n"
    printf "====================================================================\n\n"
}

print_banner

# System Detection
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        error "Unsupported system architecture: $ARCH"
        ;;
esac

info "Detected platform: ${OS}-${ARCH}"

# Installation Directories
INSTALL_DIR="${HOME}/.codedna"
BIN_DIR="${INSTALL_DIR}/bin"
TARGET_BIN="/usr/local/bin/codedna"
LOCAL_BIN="${HOME}/.local/bin/codedna"

mkdir -p "${INSTALL_DIR}"
mkdir -p "${BIN_DIR}"
mkdir -p "${HOME}/.local/bin"

# Determine execution strategy
HAS_DOCKER=false
HAS_GO=false
HAS_GIT=false

if command -v docker >/dev/null 2>&1; then
    HAS_DOCKER=true
fi

if command -v go >/dev/null 2>&1; then
    HAS_GO=true
fi

if command -v git >/dev/null 2>&1; then
    HAS_GIT=true
fi

install_wrapper() {
    local binary_src="$1"
    local installed=false

    # Try /usr/local/bin first if writable or with sudo
    if [ -w "/usr/local/bin" ]; then
        ln -sf "${binary_src}" "${TARGET_BIN}"
        success "Installed standalone binary to ${TARGET_BIN}"
        installed=true
    elif command -v sudo >/dev/null 2>&1 && [ -t 0 ]; then
        info "Requesting sudo permissions to link into /usr/local/bin..."
        if sudo ln -sf "${binary_src}" "${TARGET_BIN}"; then
            success "Installed standalone binary to ${TARGET_BIN}"
            installed=true
        fi
    fi

    if [ "$installed" = false ]; then
        ln -sf "${binary_src}" "${LOCAL_BIN}"
        success "Installed user binary to ${LOCAL_BIN}"

        # Ensure ~/.local/bin is in PATH
        if [[ ":$PATH:" != *":${HOME}/.local/bin:"* ]]; then
            warn "${HOME}/.local/bin is not in your current PATH."
            for rc in "${HOME}/.bashrc" "${HOME}/.zshrc" "${HOME}/.profile"; do
                if [ -f "$rc" ]; then
                    printf '\n# CodeDNA CLI path\nexport PATH="${HOME}/.local/bin:$PATH"\n' >> "$rc"
                    info "Appended PATH to $rc"
                fi
            done
            export PATH="${HOME}/.local/bin:$PATH"
        fi
    fi
}

# Strategy 1: Build from Source if Go and Git are available
if [ "$HAS_GO" = true ] && [ "$HAS_GIT" = true ]; then
    info "Found Go compiler and Git. Building native high-performance binary..."

    REPO_DIR="${INSTALL_DIR}/repo"
    if [ -d "${REPO_DIR}/.git" ]; then
        info "Updating existing CodeDNA cache in ${REPO_DIR}..."
        git -C "${REPO_DIR}" pull --quiet origin main || true
    else
        info "Fetching latest CodeDNA source into ${REPO_DIR}..."
        rm -rf "${REPO_DIR}"
        git clone --depth=1 --quiet "https://github.com/0xzahed/codedna.git" "${REPO_DIR}"
    fi

    info "Compiling CodeDNA executable..."
    (
        cd "${REPO_DIR}"
        # Build Next.js web assets if npm is present and web/out missing
        if [ ! -d "web/out" ] && command -v npm >/dev/null 2>&1; then
            info "Building interactive web studio UI..."
            (cd web && npm ci --silent && npm run build --silent)
        fi
        go build -ldflags="-s -w" -o "${BIN_DIR}/codedna" ./cmd/codedna
    )

    install_wrapper "${BIN_DIR}/codedna"

# Strategy 2: Docker Zero-Dependency Runtime
elif [ "$HAS_DOCKER" = true ]; then
    info "Docker detected. Creating zero-dependency containerized CLI wrapper..."

    DOCKER_WRAPPER="${BIN_DIR}/codedna"
    cat << 'WRAPPER_EOF' > "${DOCKER_WRAPPER}"
#!/usr/bin/env bash
# CodeDNA Containerized CLI Runner
IMAGE="${CODEDNA_DOCKER_IMAGE:-ghcr.io/0xzahed/codedna:latest}"

# Ensure docker image is up to date silently
if [ "$1" = "update" ]; then
    echo "Updating CodeDNA Docker image..."
    docker pull "${IMAGE}"
    exit 0
fi

# Determine target directory
TARGET_DIR="${PWD}"
for arg in "$@"; do
    if [ "$arg" = "--repo" ]; then
        TARGET_DIR="$2"
        break
    fi
done

# Run container mounting current working directory to /workspace
exec docker run --rm -it \
    -p 8080:8080 \
    -v "${TARGET_DIR}:/workspace" \
    -e CODEDNA_HOME=/root/.codedna \
    "${IMAGE}" "$@"
WRAPPER_EOF

    chmod +x "${DOCKER_WRAPPER}"
    install_wrapper "${DOCKER_WRAPPER}"

    info "Pre-pulling latest CodeDNA container image from ghcr.io..."
    docker pull ghcr.io/0xzahed/codedna:latest || warn "Container pull will retry on first execution."

else
    error "Neither Go nor Docker was found on this system. Please install Go (https://go.dev) or Docker (https://docker.com) to run CodeDNA."
fi

# Verification
printf "\n====================================================================\n"
success "CodeDNA installed successfully!"
printf "====================================================================\n\n"

printf "${C_BOLD}QUICK START GUIDE (Zero-Clone):${C_RESET}\n"
printf "  1. Navigate to ANY code repository or project folder:\n"
printf "     ${C_CYAN}cd /path/to/any-project${C_RESET}\n\n"
printf "  2. Start CodeDNA Architecture Studio:\n"
printf "     ${C_CYAN}codedna serve${C_RESET}\n\n"
printf "  3. Open your browser:\n"
printf "     ${C_GREEN}http://localhost:8080${C_RESET}\n\n"
printf "====================================================================\n"
printf "Tip: To run as a background service:\n"
printf "     ${C_DIM}codedna serve --port 8080 &${C_RESET}\n"
printf "GitHub: https://github.com/0xzahed/codedna\n"
printf "Author: Abu Zahed <zahed04x@gmail.com>\n\n"
