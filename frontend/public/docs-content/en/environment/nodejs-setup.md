# Node.js Setup

Many AI coding tools or setup scripts require Node.js. Install the current LTS version and make sure `node`, `npm`, and `npx` are available in your terminal.

## Check your environment

```bash
node -v
npm -v
npx -v
```

If all three commands print versions, your basic environment is ready.

## macOS

Use Homebrew or the official Node.js installer.

```bash
brew install node
```

Restart your terminal and run the version checks again.

## Windows

Use the official Node.js LTS installer and keep the default “Add to PATH” option enabled.

## Linux

Use your package manager or the official Node.js installation method. If the system package is too old, switch to an official source or version manager.

## Common issues

- `command not found`: restart the terminal and check PATH.
- `EACCES`: fix npm global directory permissions instead of relying on `sudo`.
- network timeout: check proxy, company network, or registry mirror settings.
