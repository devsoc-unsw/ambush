FROM ubuntu:26.04

ENV DEBIAN_FRONTEND=noninteractive
ENV GOTOOLCHAIN=local
ENV GOPATH=/root/go
ENV GOBIN=/root/go/bin
ENV BUN_INSTALL="/root/.bun"
ENV PATH="/usr/local/go/bin:/root/.bun/bin:/root/go/bin:${PATH}"

# install linux bash commands which wil be used
RUN apt-get update && apt-get install -y \
    sudo \
    curl \
    ca-certificates \
    git \
    vim \
    unzip \
    ripgrep \
    && rm -rf /var/lib/apt/lists/*

# Official Go 1.27.1 tarball (not apt golang). GOTOOLCHAIN=local keeps
# bind-mounted modules from being rewritten. Multi-arch: amd64 and arm64.
RUN arch="$(dpkg --print-architecture)" \
    && curl -fsSL "https://go.dev/dl/go1.27.1.linux-${arch}.tar.gz" \
    | tar -C /usr/local -xz

RUN curl -fsSL https://bun.com/install | bash

# default go tooling
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
    && go install github.com/pressly/goose/v3/cmd/goose@latest \
    && curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b /root/go/bin

# set-up bashrc to have basic colours
# NOTE: written via a quoted heredoc (not `echo`) so backslash sequences like
# the sed backreference \1 and the \033 colour codes are preserved verbatim.
# `/bin/sh` is dash here, and its `echo` mangles \1 -> SOH, which silently
# breaks the git-branch substitution.
RUN cat >> /root/.bashrc <<'BASHRC'

# Function to get the current git branch
parse_git_branch() {
  git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/ (\1)/'
}

# Set custom colored prompt: user@[devcontainer] path (git branch)
export PS1="\[\033[0;32m\]\u\[\033[0m\]@\[\033[1;32m\][devcontainer]\[\033[0m\] \[\033[0;34m\]\w\[\033[0m\]\[\033[0;33m\]\$(parse_git_branch)\[\033[0m\] $ "
# Enable colors for ls
alias ls='ls --color=auto'
alias ll='ls -alF --color=auto'
alias la='ls -A --color=auto'

# Enable colored auto-completion (for when you hit Tab)
bind "set colored-stats on"
bind "set colored-completion-prefix on"

# Keep Go / Bun tools on PATH in interactive and login shells.
export GOPATH="${GOPATH:-/root/go}"
export GOBIN="${GOBIN:-/root/go/bin}"
export BUN_INSTALL="${BUN_INSTALL:-/root/.bun}"
export PATH="/usr/local/go/bin:${BUN_INSTALL}/bin:${GOBIN}:${PATH}"
BASHRC

# Bind-mounted workspace often has a different owner than root.
# Trust only this image's workspace path, not every directory in the container.
RUN git config --global --add safe.directory /root/ambush

# make ambush folder
RUN mkdir /root/ambush
WORKDIR /root/ambush

