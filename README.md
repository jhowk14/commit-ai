# commit-ai 🤖

Generate high-quality Git commit messages automatically using AI (Google Gemini, OpenAI, or any OpenAI-compatible provider like OpenRouter, Groq, DeepSeek, Ollama, Cerebras, LM Studio, etc.), with support for Gitmoji and Conventional Commits.

## ✨ Features

- 🤖 AI-generated commit messages
- 🔌 **Multi-provider**: Google Gemini & **ANY OpenAI-Compatible Provider** (OpenAI, OpenRouter, Groq, DeepSeek, Ollama, Cerebras, LocalAI, LM Studio, vLLM, etc.)
- 🎨 Gitmoji mode (`-e`)
- 📦 Conventional Commits mode (default)
- 💬 **Context hints** for better messages (`-m`)
- 🌐 **Custom Base URL** (`--base-url` or `openai_base_url`)
- 🔧 Persistent configuration
- ⚙️ Interactive setup (`--setup`) with one-click provider presets
- 📝 **Custom prompts** for advanced users
- 🔍 Reads only staged diffs
- ✏️ Edit message before commit
- ↩️ Undo last commit
- 🐧 Linux & 🪟 Windows support

## 📦 Installation

### 🐧 Linux

```bash
git clone https://github.com/jhowk14/commit-ai.git
cd commit-ai/any-linux
chmod +x install.sh
./install.sh
```

### 🔷 Arch Linux (AUR)

```bash
yay -S commit-ai
commit-ai --setup
```

### 🪟 Windows

```powershell
git clone https://github.com/jhowk14/commit-ai.git
cd commit-ai\windows
.\install.ps1
```

---

## ⚙️ Configuration

### Interactive Setup

```bash
commit-ai --setup
```

The setup wizard guides you through:
- ✅ Commit format (conventional/gitmoji)
- ✅ Auto-confirm preference
- ✅ Ask to push after commit
- ✅ Use custom prompt
- ✅ AI Provider (Gemini or OpenAI / OpenAI-Compatible)
- ✅ Base URL selection (OpenAI, OpenRouter, Groq, DeepSeek, Ollama, Cerebras, LM Studio, or Custom)
- ✅ Model selection (with recommended models per provider)
- ✅ API key configuration (optional for local models like Ollama)

### Config File (`~/.commit-ai.conf`)

```ini
format=conventional
auto_confirm=false
ask_push=true
use_custom_prompt=false
provider=openai
model=gpt-4o-mini
openai_base_url=https://api.openai.com/v1
openai_api_key=your_key_here
```

---

## 🚀 Usage

```bash
git add .
commit-ai
```

### Examples

```bash
commit-ai                                              # Use defaults
commit-ai -e                                           # Gitmoji format
commit-ai -c                                           # Conventional format
commit-ai -m "fix login bug"                           # Provide context to AI
commit-ai -e -m "new feature"                          # Gitmoji with hint
commit-ai -C                                           # Use custom prompt
commit-ai -e -p                                        # Preview only
commit-ai -y                                           # Auto-commit
commit-ai --base-url "https://openrouter.ai/api/v1"   # Override base URL
commit-ai --setup                                      # Configure
commit-ai --edit-prompt                                # Edit custom prompt
```

---

## ⚙️ Options

| Flag | Description |
|------|-------------|
| `-e`, `--emoji` | Use Gitmoji format |
| `-c`, `--conv` | Use Conventional format |
| `-m`, `--message` | Provide context/hint for AI |
| `-b`, `--branch` | Switch/create branch and push after commit |
| `-s`, `-S`, `--sync` | Auto sync remote before commit |
| `-C`, `--custom` | Use custom prompt file |
| `-p`, `--preview` | Preview only |
| `-y`, `--yes` | Skip confirmation |
| `-u`, `--undo` | Undo last commit |
| `-B`, `--base-url` | Set OpenAI-compatible Base URL |
| `--setup` | Interactive setup wizard |
| `--config` | Show current configuration |
| `--edit-prompt` | Edit custom prompt file |
| `-h`, `--help` | Show help message |
| `-v`, `--version` | Show version number |

---

## 🔌 Supported Providers

### 1. OpenAI & OpenAI-Compatible Endpoints
Set `openai_base_url` to any compatible server:
- **OpenAI Official**: `https://api.openai.com/v1` (`gpt-4o-mini`, `gpt-4o`, `o3-mini`)
- **OpenRouter**: `https://openrouter.ai/api/v1` (`meta-llama/llama-3.3-70b-instruct`, `anthropic/claude-3.5-sonnet`, `deepseek/deepseek-chat`)
- **Groq**: `https://api.groq.com/openai/v1` (`llama-3.3-70b-versatile`, `llama-3.1-8b-instant`)
- **DeepSeek**: `https://api.deepseek.com/v1` (`deepseek-chat`, `deepseek-reasoner`)
- **Ollama (Local)**: `http://localhost:11434/v1` (`llama3.2`, `qwen2.5-coder:7b`, `deepseek-r1:8b`)
- **LM Studio / LocalAI**: `http://localhost:1234/v1`
- **Cerebras**: `https://api.cerebras.ai/v1` (`llama-3.3-70b`)
- **Any Custom OpenAI API Endpoint**: `https://your-custom-proxy/v1`

### 2. Google Gemini
- `gemini-3-flash-preview` (recommended)
- `gemini-2.5-flash`
- `gemini-2.0-flash`
- `gemini-2.5-pro-preview`

---

## 📝 Custom Prompts

For advanced users who want to customize the AI prompt:

```bash
commit-ai --edit-prompt  # Create/edit the prompt file
commit-ai -C             # Force use custom prompt
```

Placeholders:
- `{HISTORY}` - Recent commits
- `{FILES}` - Staged files
- `{DIFF}` - Code changes

---

## 🔑 Requirements

### Linux
- `git`, `jq`, `curl`
- API key (Gemini, OpenAI, or compatible provider; optional for local Ollama)

### Windows
- Git for Windows
- PowerShell 5.1+
- API key (Gemini, OpenAI, or compatible provider)

---

## 📄 License

MIT License © 2026 Jonathan Henrique Perozi Lourenço (jhowk14)
