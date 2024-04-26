# Gitea Update Tool

## Purpose

The Gitea Update Tool is a command-line utility designed to manage and automate the updating process for the `gitea` and `act_runner` binaries on Linux servers. It ensures that your binaries are kept up-to-date with the latest stable or development versions, managing backups and permissions seamlessly.

## Command Line Arguments

- `--info` - Displays the current, latest, and development versions of the binaries.
- `--install` - Installs the latest stable versions of the specified binaries.
- `--dev` - When used with `--install`, installs the latest development versions instead of the stable versions.
- `--version` - Displays the version of the Gitea Update Tool.
- `--path` - Specifies a custom path where the binaries are located (default is `/opt/gitea`).
- `--user` - Specifies the system user that should own the binary files (default is `app`).
- `--group` - Specifies the system group that should own the binary files (default is `app`).

## Example Usage

```bash
# Display current version information of both managed binaries
gitea-update --info

# Install the latest stable versions of gitea and act_runner
gitea-update --install

# Install the latest development versions of gitea and act_runner
gitea-update --install --dev

# Display the version of the Gitea Update Tool
gitea-update --version
```

## Configuration

The Gitea Update Tool uses a YAML configuration file to manage settings related to binary paths, user/group ownership, and source URLs for downloading updates. Below is the structure of the configuration file:

```yaml
settings:
  user: "app"
  group: "app"
  services:
    gitea:
      bin: "gitea"
      path: "/opt/gitea"
      urls:
        download: "https://dl.gitea.io/gitea/{version}/gitea-{version}-linux-amd64"
        api: "https://api.github.com/repos/go-gitea/gitea/tags"
    act_runner:
      bin: "act_runner"
      path: "/opt/gitea"
      urls:
        download: "https://dl.gitea.com/act_runner/{version}/act_runner-{version}-linux-amd64"
        api: "https://gitea.com/api/v1/repos/gitea/act_runner/tags"
```

### Configuration File Locations

The tool will search for the configuration file in the following locations, in order:

- The current directory from which the tool is run.
- The home directory of the current user (`~/gitea-update.yml`).
- `/etc/gitea-update/gitea-update.yml` for system-wide settings.

Make sure the configuration file is properly secured with the correct permissions, as it contains sensitive information regarding your server setup.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
