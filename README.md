# commit-ai 2.0

`commit-ai` gera mensagens de commit a partir das alterações já preparadas no Git. A versão 2.0 é um binário Go único: funciona nativamente em Linux, macOS e Windows, sem depender de Bash, PowerShell, `curl` ou `jq` para executar.

## Recursos

- Gitmoji e Conventional Commits;
- Gemini e qualquer API compatível com OpenAI: OpenAI, OpenRouter, Groq, DeepSeek, Ollama, LM Studio e Cerebras;
- suporte específico ao `gpt-oss-120b` do Cerebras, com orçamento de raciocínio seguro;
- prévia, confirmação/edição antes de criar o commit, desfazer, branch, push e sincronização;
- interface em Português ou English, configurável pelo `--setup`;
- envio ao remoto configurável: automático, perguntar antes ou não enviar;
- prompt customizado e contexto adicional;
- mesmo arquivo de configuração da versão 1.x: `~/.commit-ai.conf`.

## Instalação

### Linux e macOS

```bash
curl -fsSL https://raw.githubusercontent.com/jhowk14/commit-ai/v2.0.7/any-linux/install.sh | bash
commit-ai --setup
```

Também é possível compilar a partir do código-fonte:

```bash
go install github.com/jhowk14/commit-ai/v2/cmd/commit-ai@v2.0.7
```

### Windows

No PowerShell:

```powershell
irm https://raw.githubusercontent.com/jhowk14/commit-ai/v2.0.7/windows/install.ps1 | iex
commit-ai --setup
```

O instalador baixa o executável nativo e adiciona a pasta do usuário ao `PATH`.

### Arch Linux

Após a publicação no AUR:

```bash
paru -S commit-ai
commit-ai --setup
```

## Uso

```bash
git add .
commit-ai
```

| Opção | Ação |
| --- | --- |
| `-e`, `--emoji` | usa Gitmoji |
| `-c`, `--conv` | usa Conventional Commits |
| `-m`, `--message <texto>` | fornece contexto adicional |
| `-b`, `--branch <nome>` | cria/troca a branch e envia após o commit |
| `-s`, `-S`, `--sync` | guarda alterações, sincroniza e as prepara novamente |
| `-C`, `--custom` | usa `~/.commit-ai-prompt.txt` |
| `-p`, `--preview` | apenas mostra a mensagem |
| `--json` | retorna a prévia em JSON; requer `--preview` |
| `-y`, `--yes` | cria o commit sem interação |
| `-u`, `--undo` | desfaz o último commit, mantendo as alterações preparadas |
| `-B`, `--base-url <url>` | substitui a base URL OpenAI-compatível nesta execução |
| `--setup` | configuração interativa |
| `--config` | mostra a configuração atual, sem revelar chaves |
| `--edit-prompt` | cria/edita o prompt customizado |

Exemplos:

```bash
commit-ai -e -p
commit-ai -c -m "corrige login após expiração de sessão"
commit-ai -s -y
commit-ai -b feature/onboarding
```

## Configuração e migração

A versão 2.0 lê o mesmo arquivo da versão anterior, então uma instalação existente continua funcionando:

```ini
format=conventional
auto_confirm=false
language=pt-BR
push_mode=ask
use_custom_prompt=false
provider=openai
model=gpt-oss-120b
openai_base_url=https://api.cerebras.ai/v1
openai_api_key=your_key_here
```

As variáveis de ambiente têm prioridade sobre chaves salvas no arquivo: `GEMINI_API_KEY`, `OPENAI_API_KEY`, `CEREBRAS_API_KEY`, `OPENROUTER_API_KEY`, `GROQ_API_KEY` e `DEEPSEEK_API_KEY`. Para criar ou alterar a configuração, execute `commit-ai --setup`.

No setup, escolha o idioma da interface (`pt-BR` ou `en`) e o comportamento após o commit: enviar automaticamente, perguntar antes de enviar ou nunca enviar. Configurações antigas com `ask_push=true` continuam funcionando e passam a significar “perguntar antes de enviar”.

O prompt customizado pode usar `{HISTORY}`, `{FILES}` e `{DIFF}`.

## Desenvolvimento e testes

```bash
make test
make vet
make build
```

`make build` produz binários para Linux, macOS e Windows (`amd64` e `arm64`). A suíte cria repositórios Git temporários reais para testar staging, commits, undo, branch e sync; os provedores de IA são simulados com HTTP local.

## Licença

MIT © 2026 Jonathan Henrique Perozi Lourenço
