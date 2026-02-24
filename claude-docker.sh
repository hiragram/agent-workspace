#!/bin/bash
set -euo pipefail

IMAGE_NAME="claude-code-docker"
VOLUME_NAME="claude-code-credentials"
CLAUDE_HOME="${CLAUDE_HOME:-$HOME/.claude}"

# --- Dockerが使えるかチェック ---
if ! command -v docker &> /dev/null; then
  echo "Error: docker is not installed or not in PATH." >&2
  exit 1
fi

if ! docker info &> /dev/null; then
  echo "Error: Docker daemon is not running." >&2
  exit 1
fi

# --- Dockerイメージのビルド ---
build_image() {
  echo "Building Docker image '$IMAGE_NAME'..."
  local tmpdir
  tmpdir=$(mktemp -d)
  trap "rm -rf '$tmpdir'" EXIT

  # entrypoint.sh を書き出し
  cat > "$tmpdir/entrypoint.sh" << 'ENTRYPOINT_EOF'
#!/bin/bash
set -e

# .claudeディレクトリを作成（read-onlyマウントされたファイルがある場合は既に存在する）
mkdir -p /root/.claude

# 認証情報のシンボリックリンクを作成
if [ -f /root/.claude-auth/.credentials.json ]; then
  ln -sf /root/.claude-auth/.credentials.json /root/.claude/.credentials.json
fi

exec "$@"
ENTRYPOINT_EOF

  # Dockerfile を書き出し
  cat > "$tmpdir/Dockerfile" << 'DOCKERFILE_EOF'
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends git curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

SHELL ["/bin/bash", "-c"]
RUN curl -fsSL https://claude.ai/install.sh | bash

ENV PATH="/root/.local/bin:${PATH}"

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /workspace

ENTRYPOINT ["/entrypoint.sh"]
CMD ["claude"]
DOCKERFILE_EOF

  docker build -t "$IMAGE_NAME" "$tmpdir"
  trap - EXIT
  rm -rf "$tmpdir"
  echo "Docker image '$IMAGE_NAME' built successfully."
}

# --- イメージ存在チェック ---
if ! docker image inspect "$IMAGE_NAME" > /dev/null 2>&1; then
  build_image
fi

# --- volume作成 ---
docker volume create "$VOLUME_NAME" > /dev/null 2>&1 || true

# --- マウントオプション組み立て ---
MOUNT_OPTS=()

if [ -f "$CLAUDE_HOME/settings.json" ]; then
  MOUNT_OPTS+=(-v "$CLAUDE_HOME/settings.json:/root/.claude/settings.json:ro")
fi

if [ -f "$CLAUDE_HOME/CLAUDE.md" ]; then
  MOUNT_OPTS+=(-v "$CLAUDE_HOME/CLAUDE.md:/root/.claude/CLAUDE.md:ro")
fi

if [ -d "$CLAUDE_HOME/hooks" ]; then
  MOUNT_OPTS+=(-v "$CLAUDE_HOME/hooks:/root/.claude/hooks:ro")
fi

if [ -d "$CLAUDE_HOME/plugins" ]; then
  MOUNT_OPTS+=(-v "$CLAUDE_HOME/plugins:/root/.claude/plugins:ro")
fi

# --- 認証チェック ---
has_credentials=false
if docker run --rm -v "$VOLUME_NAME:/root/.claude-auth" "$IMAGE_NAME" test -f /root/.claude-auth/.credentials.json 2>/dev/null; then
  has_credentials=true
fi

if [ "$has_credentials" = false ]; then
  echo "No credentials found. Starting authentication..."
  echo "Please complete the login in your browser."
  echo ""
  docker run -it --rm \
    -v "$VOLUME_NAME:/root/.claude-auth" \
    "${MOUNT_OPTS[@]+"${MOUNT_OPTS[@]}"}" \
    "$IMAGE_NAME" \
    bash -c 'claude login && cp /root/.claude/.credentials.json /root/.claude-auth/.credentials.json 2>/dev/null || true'
  echo ""
  echo "Authentication complete."
fi

# --- Claude起動 ---
exec docker run -it --rm \
  -v "$VOLUME_NAME:/root/.claude-auth" \
  -v "$(pwd):/workspace" \
  "${MOUNT_OPTS[@]+"${MOUNT_OPTS[@]}"}" \
  "$IMAGE_NAME" \
  claude "$@"
