# Installation Guide

## Requirements

- Go 1.22 or newer

## Install from source

```bash
git clone https://github.com/GuitarWag/clitasks.git
cd clitasks
make install        # installs `tasks` into $GOBIN (or $GOPATH/bin)
```

Verify:

```bash
tasks --version
tasks --help
```

Make sure `$GOBIN` (or `$GOPATH/bin`) is on your `PATH`. If you used `make install` and the command isn't found, check:

```bash
go env GOBIN
go env GOPATH
```

If `GOBIN` is empty, the binary lands in `$GOPATH/bin`. Add that directory to your `PATH`.

## Local build (no install)

```bash
make build
./bin/tasks --version
```

## Quick start

```bash
cd ~/myproject
tasks init --name "My Project"
tasks add "Setup project structure" -p high -a yourname
tasks board
```

## Multiple projects

Each working directory can have its own `tasks.md`:

```bash
cd ~/project-a && tasks init --name "Project A"
cd ~/project-b && tasks init --name "Project B"
```

## Custom file location

Override the default `tasks.md` with a flag or env var:

```bash
tasks -f sprint-1.md board
TASK_BOARD_FILE=~/Documents/my-tasks.md tasks board
```

Put the `export` line in `~/.zshrc` to make it permanent.

## Uninstall

Delete the binary that `make install` placed in `$GOBIN`/`$GOPATH/bin`:

```bash
rm "$(go env GOPATH)/bin/tasks"   # or your $GOBIN equivalent
```

## Update

Pull and reinstall:

```bash
cd /path/to/clitasks
git pull
make install
```

## Next steps

- [README.md](README.md) — full reference
- [QUICKSTART.md](QUICKSTART.md) — short tour
- [TUI_GUIDE.md](TUI_GUIDE.md) — TUI keybindings and modes
