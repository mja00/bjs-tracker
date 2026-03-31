# bjs-tracker

A lightweight Go utility that polls BJ's Wholesale Club inventory and sends Discord webhook notifications when product stock status changes.

## Features

- Polls one or more products at a configurable interval
- Sends a Discord embed when a product becomes available or goes out of stock
- Optional periodic summary notification listing the current status of all tracked products
- Cross-compiles for Linux (amd64, arm64, ARMv7/Raspberry Pi)

## Requirements

- Go 1.21+
- A Discord webhook URL
- Your BJ's club ID and store ID

## Configuration

Copy the example below to `config.yaml` and fill in your values.

```yaml
club_id: "0372"         # your BJ's club/membership ID
store_id: 10201         # store ID (default: 10201)
check_interval: 30m     # how often to poll — minimum 1m
summary_interval: 6h    # optional: how often to post a full status summary (omit to disable)
discord_webhook_url: "https://discord.com/api/webhooks/..."

products:
  - name: "Coca-Cola 35-pack"
    article_id: "24881"
    product_id: "3000000000000166373"
  - name: "Coca-Cola Variety"
    article_id: "234351"
    product_id: "3000000000001840765"
```

### Finding article_id and product_id

Both IDs can be found in the BJ's website URLs and API responses when browsing a product page.

- `article_id` — the short numeric ID used in inventory lookups (e.g. `24881`)
- `product_id` — the long numeric ID used in price lookups (e.g. `3000000000000166373`)

## Building

```sh
# Local binary
make build

# Raspberry Pi / ARMv7
make build-pi

# All targets (linux amd64 / arm64 / arm)
make build-all

# Clean build artifacts
make clean
```

## Running

```sh
./bjs-tracker                        # uses config.yaml in the current directory
./bjs-tracker -config /path/to/config.yaml
```

The tracker runs until it receives `SIGINT` or `SIGTERM` (Ctrl-C).

## Discord Notifications

**Stock change alert** — sent whenever a product transitions between available and unavailable. On first startup, a notification is only sent if the product is already available.

**Summary** — if `summary_interval` is set, a periodic embed is posted listing every tracked product with its current status, quantity, and price.

## Deployment on a Raspberry Pi

```sh
make build-pi
scp bjs-tracker-linux-arm pi@raspberrypi.local:~/
scp config.yaml pi@raspberrypi.local:~/
ssh pi@raspberrypi.local './bjs-tracker'
```

To run as a systemd service, create `/etc/systemd/system/bjs-tracker.service`:

```ini
[Unit]
Description=BJ's Stock Tracker
After=network-online.target

[Service]
ExecStart=/home/pi/bjs-tracker -config /home/pi/config.yaml
Restart=on-failure
User=pi

[Install]
WantedBy=multi-user.target
```

Then enable it:

```sh
sudo systemctl enable --now bjs-tracker
```
