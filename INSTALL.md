# Installation Guide

## Quick Install (Global Command)

The CLI has been installed globally on your Mac. You can now use the `tasks` command from anywhere:

```bash
tasks --version  # Should show: 1.0.0
tasks --help     # Show all commands
```

## How It Works

The installation created a global symlink using `npm link`. This means:

- The `tasks` command is available in any directory
- It points to: `/Users/wagnersilva/.nvm/versions/node/v22.21.1/bin/tasks`
- Each directory can have its own `tasks.md` file
- You can work on multiple projects with separate task boards

## Quick Start

### 1. Create your first board

```bash
cd ~/myproject
tasks init --name "My Project"
```

This creates a `tasks.md` file in the current directory.

### 2. Add a task

```bash
tasks add "Setup project structure" -p high -a yourname
```

### 3. View the board

```bash
tasks board
```

### 4. Start working

```bash
tasks start T-XXXXX  # Use the task ID from the board
```

### 5. Complete the task

```bash
tasks complete T-XXXXX
```

## Multiple Projects

You can have different task boards for different projects:

```bash
cd ~/project-a
tasks init --name "Project A"
tasks add "Feature 1" -p high

cd ~/project-b
tasks init --name "Project B"
tasks add "Feature 2" -p medium

# Each directory has its own tasks.md file
```

## Uninstall (if needed)

To remove the global command:

```bash
cd /Users/wagnersilva/Desktop/git/wag/clitasks
npm unlink
```

## Update the CLI

If you make changes to the source code:

```bash
cd /Users/wagnersilva/Desktop/git/wag/clitasks
npm run build  # Rebuild the changes
# The global command automatically uses the new version
```

## Environment Variable

You can set a custom default file location:

```bash
export TASK_BOARD_FILE=~/Documents/my-tasks.md
tasks board  # Will use ~/Documents/my-tasks.md
```

Add this to your `~/.zshrc` or `~/.bashrc` to make it permanent.

## Verify Installation

Run these commands to verify everything works:

```bash
# Check version
tasks --version

# Show help
tasks --help

# Create a test board
cd /tmp
tasks init --name "Test Board"
tasks add "Test task" -p high
tasks board
```

## Troubleshooting

### Command not found

If you get "command not found", your npm global bin directory might not be in your PATH. Check with:

```bash
npm bin -g
```

Make sure this directory is in your PATH environment variable.

### Permission errors

If you get permission errors, you may need to fix npm permissions:

```bash
mkdir -p ~/.npm-global
npm config set prefix '~/.npm-global'
echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.zshrc
source ~/.zshrc
```

Then reinstall:

```bash
cd /Users/wagnersilva/Desktop/git/wag/clitasks
npm link
```

## Next Steps

- Read [README.md](README.md) for detailed documentation
- Check [QUICKSTART.md](QUICKSTART.md) for a 5-minute tutorial
- See [COLLABORATION_EXAMPLE.md](COLLABORATION_EXAMPLE.md) for human-AI workflows
- View [example-tasks.md](example-tasks.md) for inspiration

Enjoy your new task management system!
