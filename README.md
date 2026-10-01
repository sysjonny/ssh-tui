# ssh-tui

A small terminal menu for connecting to host aliases from your OpenSSH config.
OpenSSH reads the selected host's settings, so options such as `HostName`,
`User`, keys, ports, and `Include` directives continue to work as usual.

## Run

```sh
go run .
```

The app reads `~/.ssh/config` by default and lists concrete `Host` aliases
(wildcard patterns are not selectable). Choose a number to connect, or `q` to
quit. Use `-config /path/to/config` or `SSH_TUI_CONFIG` to select the app's
configuration file.

## Password environment variable

Create `~/.config/ssh-tui/config` to name the environment variable that holds
the SSH password:

```ini
ssh_config = ~/.ssh/config
password_env = SSH_TUI_PASSWORD
```

Set that variable in the environment before starting the app:

```sh
export SSH_TUI_PASSWORD='your-password'
go run .
```

The password is not stored in the configuration file. For the selected host,
ssh-tui passes it to OpenSSH through a temporary askpass helper and the child
process environment; the helper is removed after the SSH command exits. If
`password_env` is omitted, OpenSSH's normal authentication and prompting apply.
