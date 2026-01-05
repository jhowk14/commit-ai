# commit-ai 🤖

Generate high-quality Git commit messages automatically using AI (Google Gemini or OpenAI), with support for Gitmoji and Conventional Commits.

## ✨ Features

- 🤖 AI-generated commit messages
- 🔌 **Multi-provider**: Gemini & OpenAI support
- 🎨 Gitmoji mode (`-e`)
- 📦 Conventional Commits mode (default)
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
- ✅ AI Provider (Gemini/OpenAI)
- ✅ Model selection
- ✅ API key configuration

### Config File (`~/.commit-ai.conf`)

```ini
format=conventional
auto_confirm=false
ask_push=true
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
commit-ai -e -p        # Preview only
commit-ai -y           # Auto-commit
commit-ai --setup      # Configure
commit-ai --edit-prompt # Custom prompt
```

---

## ⚙️ Options

| Flag | Description |
|------|-------------|
| `-e`, `--emoji` | Use Gitmoji format |
| `-c`, `--conv` | Use Conventional format |
| `-p`, `--preview` | Preview only |
| `-y`, `--yes` | Skip confirmation |
| `-u`, `--undo` | Undo last commit |
| `-s`, `--setup` | Interactive setup |
| `--config` | Show config |
| `--edit-prompt` | Custom prompt editor |

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
commit-ai --edit-prompt
```

This creates `~/.commit-ai-prompt.txt` with placeholders:
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
