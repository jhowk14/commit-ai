#!/usr/bin/env bash
# commit-ai installer — Linux and macOS
set -euo pipefail

VERSION="${COMMIT_AI_VERSION:-2.0.7}"
REPOSITORY="jhowk14/commit-ai"
INSTALL_DIR="${COMMIT_AI_INSTALL_DIR:-/usr/local/bin}"

case "$(uname -s)" in
  Linux) platform="linux" ;;
  Darwin) platform="darwin" ;;
  *) echo "❌ Sistema não suportado: $(uname -s)"; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64|amd64) architecture="amd64" ;;
  aarch64|arm64) architecture="arm64" ;;
  *) echo "❌ Arquitetura não suportada: $(uname -m)"; exit 1 ;;
esac

if ! command -v git >/dev/null; then
  echo "❌ Git é necessário. Instale-o e execute novamente."
  exit 1
fi
if ! command -v curl >/dev/null; then
  echo "❌ curl é necessário para baixar o binário."
  exit 1
fi

asset="commit-ai-${platform}-${architecture}"
url="https://github.com/${REPOSITORY}/releases/download/v${VERSION}/${asset}"
temporary_file="$(mktemp)"
trap 'rm -f "$temporary_file"' EXIT

echo "📥 Baixando commit-ai v${VERSION} para ${platform}/${architecture}..."
curl --fail --location --silent --show-error "$url" --output "$temporary_file"
chmod +x "$temporary_file"
if [[ -w "$INSTALL_DIR" ]]; then
  install -m 0755 "$temporary_file" "$INSTALL_DIR/commit-ai"
else
  echo "🔐 Instalando em $INSTALL_DIR com sudo..."
  sudo install -m 0755 "$temporary_file" "$INSTALL_DIR/commit-ai"
fi
echo "✅ commit-ai instalado: $($INSTALL_DIR/commit-ai --version)"
echo "Execute 'commit-ai --setup' para configurar o provedor de IA."
