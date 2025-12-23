# gitmoji-gemini 🚀✨

Generate high-quality Git commit messages automatically using Google Gemini, with support for Gitmoji and Conventional Commits — based only on staged changes.

## ✨ Features

- 🤖 AI-generated commit messages using Gemini
- 🎨 Gitmoji mode (`-e`)
- 📦 Conventional Commits mode (default)
- 🔍 Reads only staged diffs (`git add`)
- 👀 Preview mode (no commit)
- ✏️ Edit message before commit
- ↩️ Undo last commit (soft reset)
- 🧠 Uses your own commit history as style reference
- ⚡ Fast, lightweight Bash script

## 📦 Installation

### 🟦 Arch Linux (AUR)

```bash
yay -S gitmoji-gemini
```

or

```bash
paru -S gitmoji-gemini
```

### 🟨 Manual install (any Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/jhowk14/commit-ai/main/commit-ai.sh \
  | sudo tee /usr/bin/commit-ai > /dev/null

sudo chmod +x /usr/bin/commit-ai
```

## 🔑 Requirements

- `git`
- `jq`
- `curl`
- A Google Gemini API key

## 🔐 Environment variable

You must export your API key:

```bash
export GEMINI_API_KEY="your_api_key_here"
```

To persist it:

```bash
echo 'export GEMINI_API_KEY="your_api_key_here"' >> ~/.bashrc
```

(or `~/.zshrc`)

## 🚀 Usage

### 1️⃣ Stage your changes

```bash
git add .
```

### 2️⃣ Generate commit message

#### Conventional Commits (default)

```bash
gitmoji-gemini
```

Example output:

```
feat: add support for preview mode
```

#### Gitmoji mode

```bash
gitmoji-gemini -e
```

Example output:

```
✨ Add preview mode support
```

## ⚙️ Options

| Flag | Description |
|------|-------------|
| `-e`, `--emoji` | Use Gitmoji commit format |
| `-p`, `--preview` | Preview commit message only |
| `-y`, `--yes` | Skip confirmation prompt |
| `-u`, `--undo` | Undo last commit (soft reset) |

### 🔍 Preview only

```bash
gitmoji-gemini -e -p
```

### ⚡ Auto-commit (no prompt)

```bash
gitmoji-gemini -y
```

### ↩️ Undo last commit

```bash
gitmoji-gemini -u
```

> This keeps your changes staged

## 🧠 How it works

1. Reads only staged files
2. Extracts a trimmed diff
3. Sends context to Gemini
4. Enforces strict formatting rules
5. Normalizes output
6. Creates a clean commit message

## 🛡️ Security

- ❌ API keys are never hardcoded
- ❌ No repository data is stored
- ✅ Only staged diffs are sent
- ✅ Fully local execution

## 📄 License

MIT License © 2025

## 🤝 Contributing

Pull requests are welcome.

**Ideas:**

- Git hook support (`prepare-commit-msg`)
- Model selection flag
- Commit message cache
- Shell completion

## ⭐ Acknowledgments

- [Gitmoji](https://gitmoji.dev/) community
- [Google Gemini](https://deepmind.google/technologies/gemini/)
- Arch Linux AUR maintainers
