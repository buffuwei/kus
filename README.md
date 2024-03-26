kus is a simple terminal UI for kuboard.

# Installation

**Remember to set the `RELEASE` variable to the latest release tag**
```sh
RELEASE="v0.1.2" && wget "https://github.com/buffuwei/kus/releases/download/${RELEASE}/kus" -O kus && chmod +x kus && mv kus /usr/local/bin
```

or 

`sh install.sh`

# Configuration

Configuration file path: `~/.kus/config.yaml`

Example:
```yaml
kuboard:
    host: kuboard.xxx.yyy
    username: Bob
    password: pass
    clusters:
        - DEV
        - TEST
selected:
    cluster: DEV
    namespace: business
```
