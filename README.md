# commit-ai 🤖

Generate high-quality Git commit messages automatically using Google Gemini, with support for Gitmoji and Conventional Commits — based only on staged changes.

## ✨ Features

- 🤖 AI-generated commit messages using Gemini
- 🎨 Gitmoji mode (`-e`)
- 📦 Conventional Commits mode (default)
- 🔧 **Persistent configuration** (`~/.commit-ai.conf`)
- ⚙️ **Interactive setup** (`--setup`)
- 🔍 Reads only staged diffs
- ✏️ Edit message before commit
- ↩️ Undo last commit (soft reset)
- 🧠 Uses your commit history as reference
- 🐧 Linux & 🪟 Windows support

## 📦 Installation

### 🐧 Linux

#### Quick Install (Any Distro)

```bash
git clone https://github.com/jhowk14/commit-ai.git
cd commit-ai/any-linux
chmod +x install.sh
./install.sh
```

#### Arch Linux (AUR)

```bash
yay -S commit-ai
# Then run: commit-ai --setup
```

---

### 🪟 Windows

#### PowerShell Installation

```powershell
# Clone and install
git clone https://github.com/jhowk14/commit-ai.git
cd commit-ai\windows
.\install.ps1
```

The installer will:
- Copy scripts to `$HOME\bin\`
- Add to PATH
- Create PowerShell alias
- Run interactive setup

---

## ⚙️ Configuration

### Interactive Setup

Run the setup wizard to configure your preferences:

```bash
# Linux
commit-ai --setup

# Windows
commit-ai -Setup
```

### Config File

Configuration is stored in `~/.commit-ai.conf`:

```ini
# Default commit format: conventional | gitmoji
format=conventional

# Auto-confirm commits: true | false
auto_confirm=false

# Gemini model
model=gemini-2.0-flash

# API Key (optional - can use env var instead)
api_key=your_key_here
```

### View Current Config

```bash
commit-ai --config  # or -c
```

---

## 🚀 Usage

### 1️⃣ Stage your changes

```bash
git add .
```

### 2️⃣ Generate commit message

```bash
# Conventional Commits (default or set in config)
commit-ai

# Gitmoji mode
commit-ai -e
```

### Examples

```bash
commit-ai              # Use config defaults
commit-ai -e           # Gitmoji format
commit-ai -e -p        # Preview Gitmoji message
commit-ai -y           # Auto-commit without confirmation
commit-ai -u           # Undo last commit
commit-ai --setup      # Configure preferences
```

---

## ⚙️ Options

| Flag | Description |
|------|-------------|
| `-e`, `--emoji` | Use Gitmoji commit format |
| `-p`, `--preview` | Preview commit message only |
| `-y`, `--yes` | Skip confirmation prompt |
| `-u`, `--undo` | Undo last commit (soft reset) |
| `-s`, `--setup` | Interactive configuration |
| `-c`, `--config` | Show current configuration |
| `-h`, `--help` | Show help message |
| `-v`, `--version` | Show version |

---

## 🔑 Requirements

### Linux
- `git`, `jq`, `curl`
- Google Gemini API key

### Windows
- Git for Windows
- PowerShell 5.1+ (built-in on Windows 10/11)
- Google Gemini API key

---

## 📁 Project Structure

```
commit-ai/
├── any-linux/          # Linux scripts
│   ├── commit-ai.sh    # Main script
│   └── install.sh      # Installer
├── arch-linux/         # AUR package
│   ├── PKGBUILD
│   └── .SRCINFO
├── windows/            # Windows scripts
│   ├── commit-ai.ps1   # Main script
│   ├── commit-ai.bat   # Batch wrapper
│   └── install.ps1     # Installer
├── docs/               # Website
└── .commit-ai.conf.example
```

---

## 🛡️ Security

- ❌ API keys are never hardcoded
- ✅ Can store API key in config file or env var
- ✅ Only staged diffs are sent
- ✅ Fully local execution

---

## 📄 License

MIT License © 2025

## 🤝 Contributing

Pull requests are welcome!

**Ideas:**
- Git hook support (`prepare-commit-msg`)
- Model selection flag
- Shell completion

## ⭐ Acknowledgments

- [Gitmoji](https://gitmoji.dev/)
- [Google Gemini](https://deepmind.google/technologies/gemini/)
