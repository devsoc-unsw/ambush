FROM ubuntu:26.04

ENV DEBIAN_FRONTEND=noninteractive
# install linux bash commands which wil be used
RUN apt-get update && apt-get install -y \
    sudo \
    curl \
    git \
    vim \
    unzip \
    ripgrep \
    golang \
    neovim \
    nodejs \
    # mason uses npm
    npm \
    # telescope uses fd find
    fd-find \
    # build-essential grabs gcc/make/libc headers (needed by cgo and nvim)
    build-essential \
    # ubuntu ships fd as `fdfind`; telescope looks for `fd`
    && ln -s "$(command -v fdfind)" /usr/local/bin/fd \
    && rm -rf /var/lib/apt/lists/*

ENV BUN_INSTALL="/root/.bun"
ENV PATH="/root/.bun/bun:%${PATH}"
RUN curl -fsSL https://bun.com/install | bash

# default go tooling
ENV PATH="/root/go/bin:${PATH}"
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin

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
BASHRC

# make ambush folder
RUN mkdir /root/ambush
WORKDIR /root/ambush
