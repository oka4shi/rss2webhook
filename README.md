# rss2webhook
## Overview

A lightweight command-line tool written in Go that periodically checks RSS feeds and sends notifications to specified webhook URLs of Discord when new items are found.

## Installation
```shell
git clone https://github.com/oka4shi/rss2webhook
cd rss2webhook
make build
```

## Usage

### Create a config file
#### Example
```yaml
items:
- target: https://examaple.com/rss
  webhook_url: https://discord.com/api/webhooks/**
  color: "#ad1a08"
  interval: 0
- target: https://examaple.com/rss
  webhook_url: https://discord.com/api/webhooks/**
  color: "#FFFF00"
  interval: 0
```

#### Fields
- target (string, required)

  The URL of the RSS feed to monitor.

- webhook_url (string, required)

  The Discord webhook URL where new feed items will be sent.

- color (string, optional, default: `"#000000"`)

  A hexadecimal color code used for Discord embed messages.

- interval (integer, optional, default: 0)

  The minimal interval in minutes. If the time since the last check is less than this value, the feed will be skipped.

### Run the Tool
Set the path to your configuration file using the `R2W_CONFIG` environment variable, then run:
```
./rss2webhook
```

Do not edit the configuration file while the script is running.

This tool does not run as a daemon. Use a scheduler such as cron or systemd-timer to run it periodically.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
