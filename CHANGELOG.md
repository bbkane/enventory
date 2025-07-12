# Changelog

All notable changes to this project will be documented in this file. The format
is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

Note the the latest version is usually work in progress and may have not yet been released.

# v0.0.20

## Added

- `shell zsh chdir` - this does a fine-grained update of the environment when you change envs instead of the previous unexport all, then export all

## Changed

- `shell zsh init` now uses `shell zsh chdir` instead of `shell zsh unexport; shell zsh export`. If bugs are found, revert back to the old strategy with `shell zsh init --chpwd-strategy v0.0.19`. I aso removed `--print-export-env` and `--print-chwpd-hook` and now unconditionally print these in init

# v0.0.19

## Added

- `var ref update`!

## Fixed

- All operations now run in top level transactions (fixes #87)
- Add completions for `var update --new-env` (fixes #100)
- Allow '' (empty string) as string flag value (with `warg` update) (fixes #101)

# v0.0.18

## Fixed

- Fix completion for bool flags (true/false) and --ref-var (complete from env passed to --ref-env)

# v0.0.17

## Changed

- Rename app from envelope to enventory. "envelope" turns out to be a very similar CLI: https://github.com/mattrighetti/envelope

### Migration steps

Install `enventory`: `brew install bbkane/tap/enventory` or alternatives

Change `~/.zshrc`: replace `eval "$(envelope shell zsh init)"` with `eval "$(enventory shell zsh init)"`

Rename database: `cp ~/.config/envelope.db ~/.config/enventory.db`

Clear shell completions cache if needed: `rm ~/.zcompdump`

Open a new shell and test that everything works

Uninstall envelope and delete `~/.config/envelope.db`

# v0.0.15 and v0.0.16

## Fixed

- Fix zsh tab completion install in generated Homebrew formula

# v0.0.14

## Added

- Added zsh tab completion!!

Example (note that `~/fbin` is in my `$FPATH`):

```zsh
enventory completion zsh > ~/fbin/_enventory
```

## Changed

- Shortened flag names:
  - `--env-name` -> `--env`
  - `--ref-env-name` -> `--ref-env`
  - `--ref-var-name` -> `--ref-var`
  - `--new-env-name` -> `--new-env`

# v0.0.13

## Added

- `ENVELOPE_MASK` env var to control `--mask`

## Changed

- Due to warg update (v0.0.26), flags must now be passed after commands

## Fixed

- #58 - print error if deleting or updating non-existent envs/vars/refs

# v0.0.12

## Changed

- `env print-script --shell zsh --type export` -> `shell zsh export`.
- `env print-script --shell zsh --type unexport` -> `shell zsh unexport`
- Moved `env var` subcommands to toplevel. For example, `env var create` becomes `var create`
- Moved `env ref` subcommands under var. For example, `env ref create` becomes `var ref create`

# v0.0.11

## Changed

- terser print-script output

## Fixed

- `--width` flag for `env var show` and `env show`

## Removed

- Removed `keyring` commands in favor of planned `secret` commands using age + the existing db

# v0.0.10

## Added

- `--width` flag for output using tables. Defaults to the current terminal width.

## Changed

- Skip printing Comment if it's blank
- Skip printing UpdateTime if it equals CreateTime
- Moved `init zsh` to `shell zsh init`. If you see an error when you start your shell, you need to update this line in `~/.zshrc`

# v0.0.9

## Added

- `init zsh` has a `--print-autoload` flag now
- `env var update` command

## Changed

- `env var create --value` is optional, and the value is prompted for if not given
- `init` -> `init zsh` so we can add zsh-specific flags and make subcommands for other shells

# v0.0.8

## Added

- `--format` flag to change output format (currently only supports the default (`table` and `value-only` for vars and refs))

## Fixed

- `--mask` flag now hides values in `env var show`

# v0.0.7

## Added

- `--mask` flag to show commands to hide sensitive values

# v0.0.6

## Added

- `init` now takes flags to gate stuff to print
- `init` now adds `export-env` and `unexport-env` to the environment

## Fixed

- Fixed spelling for `env ref create`
- Unexport `$OLDPWD` env before exporting `$PWD`, so if they share an export name, the new one isn't deleted

# v0.0.5

## Added

- `env ref` commands
- `print-script --shell` flag

## Changed

- `env localvar` commands renamed to `env var`
- Use key-value tables for output
- Show `env ref`s in `env show`
- Export `env ref`s in `env export`
- Show `env ref`s in `env var show`
- When listing the same type of item, print a single table with multiple sections instead of separate tables

# v0.0.4

## Added

- `--confirm` flag to deletes / updates
- `enventory env print-script --type unexport`
- `enventory init`

## Changed

- `--sqlite-dsn` -> `--db-path`. Reads from `ENVELOPE_DB_PATH` env var now too
- made all tests parallel
- more concise date format
- use `--help detailed` by default

# v0.0.3

## Added

- `--no-env-no-problem` flag
