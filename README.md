# commit-ai 🤖

Generate high-quality Git commit messages automatically using AI (Google Gemini or OpenAI), with support for Gitmoji and Conventional Commits.

## ✨ Features

- 🤖 AI-generated commit messages
- 🔌 **Multi-provider**: Gemini & OpenAI support
- 🎨 Gitmoji mode (`-e`)
- 📦 Conventional Commits mode (default)
- 💬 **Context hints** for better messages (`-m`)
- 🔧 Persistent configuration
- ⚙️ Interactive setup (`--setup`)
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
- ✅ AI Provider (Gemini/OpenAI)
- ✅ Model selection
- ✅ API key configuration

### Config File (`~/.commit-ai.conf`)

```ini
format=conventional
auto_confirm=false
ask_push=true
use_custom_prompt=false
provider=gemini
model=gemini-3-flash-preview
gemini_api_key=your_key
openai_api_key=your_key
```

---

## 🚀 Usage

```bash
git add .
commit-ai
```

### Examples

```bash
commit-ai              # Use defaults
commit-ai -e           # Gitmoji format
commit-ai -c           # Conventional format
commit-ai -m "fix login bug"  # Provide context to AI
commit-ai -e -m "new feature" # Gitmoji with hint
commit-ai -C           # Use custom prompt
commit-ai -e -p        # Preview only
commit-ai -y           # Auto-commit
commit-ai --setup      # Configure
commit-ai --edit-prompt # Edit custom prompt
```

---

## ⚙️ Options

| Flag | Description |
|------|-------------|
| `-e`, `--emoji` | Use Gitmoji format |
| `-c`, `--conv` | Use Conventional format |
| `-m`, `--message` | Provide context/hint for AI |
| `-C`, `--custom` | Use custom prompt file |
| `-p`, `--preview` | Preview only |
| `-y`, `--yes` | Skip confirmation |
| `-u`, `--undo` | Undo last commit |
| `-s`, `--setup` | Interactive setup |
| `--config` | Show config |
| `--edit-prompt` | Edit custom prompt file |
| `-h`, `--help` | Show help message |
| `-v`, `--version` | Show version number |

### Full Help Output

```
commit-ai v1.3.0 — AI-powered Git commit messages

USAGE:
  commit-ai [OPTIONS]

OPTIONS:
  -e, --emoji       Use Gitmoji commit format (emoji prefix)
  -c, --conv        Use Conventional Commits format (overrides config)
  -m, --message     Provide context/hint for AI (e.g., -m "fix login bug")
  -C, --custom      Use custom prompt file (~/.commit-ai-prompt.txt)
  -p, --preview     Preview commit message only (no commit)
  -y, --yes         Skip confirmation prompt (auto-commit)
  -u, --undo        Undo last commit (soft reset, keeps changes staged)
  -s, --setup       Interactive configuration setup
  --config          Show current configuration
  --edit-prompt     Edit custom prompt for advanced users
  -h, --help        Show this help message
  -v, --version     Show version number

PROVIDERS:
  gemini            Google Gemini (default)
  openai            OpenAI GPT models

EXAMPLES:
  commit-ai                          # Use configured defaults
  commit-ai -e                       # Gitmoji format
  commit-ai -c                       # Conventional format
  commit-ai -m "added user auth"     # AI uses hint for better message
  commit-ai -e -m "refactored api"   # Gitmoji with context
  commit-ai -e -p                    # Preview Gitmoji message
  commit-ai -y                       # Auto-commit without confirmation
  commit-ai --setup                  # Configure preferences
  commit-ai --edit-prompt            # Customize AI prompt

CONFIG FILE:
  Location: ~/.commit-ai.conf
  
  Available settings:
    format=conventional|gitmoji
    auto_confirm=true|false
    ask_push=true|false
    use_custom_prompt=true|false
    provider=gemini|openai
    model=<model-name>
    gemini_api_key=your_key
    openai_api_key=your_key

ENVIRONMENT:
  GEMINI_API_KEY     Google Gemini API key
  OPENAI_API_KEY     OpenAI API key
```

---

## 🔌 Supported Providers

### Gemini (Google)
- `gemini-3-flash-preview` (recommended)
- `gemini-2.5-flash`
- `gemini-2.0-flash`
- `gemini-2.5-pro-preview`

### OpenAI
- `gpt-4o-mini` (recommended)
- `gpt-4o`
- `gpt-4-turbo`
- `gpt-3.5-turbo`

---

## 📝 Custom Prompts

For advanced users who want to customize the AI prompt:

```bash
commit-ai --edit-prompt  # Create/edit the prompt file
commit-ai -C             # Force use custom prompt
```

You can also enable it permanently via `--setup` or by setting `use_custom_prompt=true` in your config.

The prompt file (`~/.commit-ai-prompt.txt`) supports placeholders:
- `{HISTORY}` - Recent commits
- `{FILES}` - Staged files
- `{DIFF}` - Code changes

---

## 🔑 Requirements

### Linux
- `git`, `jq`, `curl`
- API key (Gemini or OpenAI)

### Windows
- Git for Windows
- PowerShell 5.1+
- API key (Gemini or OpenAI)

---

## 📁 Project Structure

```
commit-ai/
├── any-linux/
│   ├── commit-ai.sh
│   └── install.sh
├── arch-linux/
│   ├── PKGBUILD
│   └── .SRCINFO
├── windows/
│   ├── commit-ai.ps1
│   ├── commit-ai.bat
│   └── install.ps1
└── docs/
```

---

## 📄 License

MIT License © 2025

## 🤝 Contributing

Pull requests welcome!

## ⭐ Credits

- [Gitmoji](https://gitmoji.dev/)
- [Google Gemini](https://deepmind.google/technologies/gemini/)
- [OpenAI](https://openai.com/)
