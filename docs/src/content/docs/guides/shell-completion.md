---
title: Shell Completion
description: Set up tab completion for the Ancla CLI in your shell.
---

The Ancla CLI supports tab completion for commands, subcommands, and flags.

## Bash

```bash
# Load for current session
source <(ancla completion bash)

# Load permanently
ancla completion bash > /etc/bash_completion.d/ancla
# Or for a single user:
ancla completion bash > ~/.local/share/bash-completion/completions/ancla
```

## Zsh

```zsh
# Load for current session
source <(ancla completion zsh)

# Load permanently (add to an fpath directory)
ancla completion zsh > "${fpath[1]}/_ancla"
```

If completions aren't working, ensure `compinit` is called in your `.zshrc`:

```zsh
autoload -Uz compinit && compinit
```

## Fish

```fish
# Load for current session
ancla completion fish | source

# Load permanently
ancla completion fish > ~/.config/fish/completions/ancla.fish
```

## PowerShell

```powershell
# Load for current session
ancla completion powershell | Out-String | Invoke-Expression

# Load permanently (add to your profile)
ancla completion powershell >> $PROFILE
```
