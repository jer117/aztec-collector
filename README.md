# Aztec Collector

A metrics collector for Aztec network nodes that exposes Prometheus metrics and provides health monitoring capabilities.

## Overview

The Aztec Collector queries Aztec nodes via JSON-RPC and collects metrics about:

- Block heights (latest, proven, pending)
- Node sync status
- Pending transaction count
- Base gas fees
- Node health and readiness
- Comparison against external reference block height

## Features

- **Prometheus Metrics**: Exposes metrics at `/metrics` endpoint for scraping
- **Health Checks**: Provides `/health` and `/ready` endpoints for monitoring
- **State API**: Returns full node state as JSON at `/state` endpoint
- **JSON Logging**: Writes collected state to a JSON file for debugging
- **Reference Service**: Compare block height against external Aztec nodes or reference services
- **Alerting**: Send alerts to Slack, Discord, Telegram, or custom webhooks when health thresholds are breached
- **Environment Variable Override**: All config options can be overridden via environment variables (perfect for Docker/Kubernetes)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/your-org/aztec-collector.git
cd aztec-collector

# Build
make build

# Run
./build/aztec-collector-$(uname -s | tr A-Z a-z)-$(go env GOARCH) --config example-config.yaml
```

### Using Docker

```bash
# Build Docker image
docker build -t aztec-collector:latest .

# Run with environment variables
docker run -d \
  -p 13285:13285 \
  -e AZTEC_COLLECTOR_PROTOCOL_RPC_URL=http://your-aztec-node:8080 \
  -e AZTEC_COLLECTOR_REF_SERVICE_ENABLED=true \
  -e AZTEC_COLLECTOR_REF_SERVICE_USE_EXTERNAL_RPC=true \
  -e AZTEC_COLLECTOR_REF_SERVICE_EXTERNAL_RPC_URL=http://reference-node:8080 \
  aztec-collector:latest
```

### Using Docker Compose

See [docker-compose.example.yaml](docker-compose.example.yaml) for a complete example.

```bash
# Copy and customize the example
cp docker-compose.example.yaml docker-compose.yaml

# Start all services
docker-compose up -d
```

## Configuration

### Config File

Create a configuration file (e.g., `config.yaml`):

```yaml
protocol:
  name: "aztec"
  rpc_url: "http://localhost:8080"
  admin_url: "http://localhost:8880"

interval: 30s

api:
  enabled: true
  listen: "0.0.0.0:13285"

health:
  min_peers: 3
  max_blocks_behind: 50
  max_blocks_ahead: 10

jsonlog:
  path: ./aztec-collector.json

# Reference service for block height comparison
ref_service:
  enabled: true
  use_external_rpc: true
  external_rpc_url: "http://reference-aztec-node:8080"
  timeout: 5s
```

### Environment Variables

All configuration options can be overridden with environment variables using the format:

```
AZTEC_COLLECTOR_<SECTION>_<KEY>
```

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `AZTEC_COLLECTOR_PROTOCOL_NAME` | Protocol identifier | `aztec` |
| `AZTEC_COLLECTOR_PROTOCOL_RPC_URL` | Aztec node RPC URL | `http://localhost:8080` |
| `AZTEC_COLLECTOR_PROTOCOL_ADMIN_URL` | Aztec node admin URL | `http://localhost:8880` |
| `AZTEC_COLLECTOR_INTERVAL` | Collection interval | `30s` |
| `AZTEC_COLLECTOR_API_ENABLED` | Enable HTTP API server | `true` |
| `AZTEC_COLLECTOR_API_LISTEN` | API server listen address | `localhost:13285` |
| `AZTEC_COLLECTOR_HEALTH_MIN_PEERS` | Minimum peer count | `1` |
| `AZTEC_COLLECTOR_HEALTH_MAX_BLOCKS_BEHIND` | Max blocks behind before error | `100` |
| `AZTEC_COLLECTOR_HEALTH_MAX_BLOCKS_AHEAD` | Max blocks ahead before warn | `10` |
| `AZTEC_COLLECTOR_JSONLOG_PATH` | Path to JSON log file | `./aztec-collector.json` |
| `AZTEC_COLLECTOR_REF_SERVICE_ENABLED` | Enable reference service | `false` |
| `AZTEC_COLLECTOR_REF_SERVICE_URL` | HTTP reference service URL | `` |
| `AZTEC_COLLECTOR_REF_SERVICE_PROTOCOL` | Protocol for ref service | `aztec` |
| `AZTEC_COLLECTOR_REF_SERVICE_NETWORK` | Network for ref service | `mainnet` |
| `AZTEC_COLLECTOR_REF_SERVICE_USE_EXTERNAL_RPC` | Use external RPC as reference | `false` |
| `AZTEC_COLLECTOR_REF_SERVICE_EXTERNAL_RPC_URL` | External Aztec RPC URL | `` |
| `AZTEC_COLLECTOR_REF_SERVICE_TIMEOUT` | Reference query timeout | `5s` |
| `AZTEC_COLLECTOR_ALERTING_ENABLED` | Enable alerting | `false` |
| `AZTEC_COLLECTOR_ALERTING_COOLDOWN` | Cooldown between alerts | `5m` |
| `AZTEC_COLLECTOR_ALERTING_NODE_NAME` | Node identifier in alerts | `aztec-node` |
| `AZTEC_COLLECTOR_ALERTING_SLACK_ENABLED` | Enable Slack alerts | `false` |
| `AZTEC_COLLECTOR_ALERTING_SLACK_WEBHOOK_URL` | Slack webhook URL | `` |
| `AZTEC_COLLECTOR_ALERTING_DISCORD_ENABLED` | Enable Discord alerts | `false` |
| `AZTEC_COLLECTOR_ALERTING_DISCORD_WEBHOOK_URL` | Discord webhook URL | `` |
| `AZTEC_COLLECTOR_ALERTING_TELEGRAM_ENABLED` | Enable Telegram alerts | `false` |
| `AZTEC_COLLECTOR_ALERTING_TELEGRAM_BOT_TOKEN` | Telegram bot token | `` |
| `AZTEC_COLLECTOR_ALERTING_TELEGRAM_CHAT_ID` | Telegram chat ID | `` |
| `AZTEC_COLLECTOR_ALERTING_GENERIC_ENABLED` | Enable generic webhook | `false` |
| `AZTEC_COLLECTOR_ALERTING_GENERIC_WEBHOOK_URL` | Generic webhook URL | `` |

### Docker Compose Example

```yaml
services:
  aztec-collector:
    image: aztec-collector:latest
    environment:
      - AZTEC_COLLECTOR_PROTOCOL_RPC_URL=http://aztec-node:8080
      - AZTEC_COLLECTOR_INTERVAL=30s
      - AZTEC_COLLECTOR_API_LISTEN=0.0.0.0:13285
      - AZTEC_COLLECTOR_REF_SERVICE_ENABLED=true
      - AZTEC_COLLECTOR_REF_SERVICE_USE_EXTERNAL_RPC=true
      - AZTEC_COLLECTOR_REF_SERVICE_EXTERNAL_RPC_URL=http://public-aztec-rpc:8080
    ports:
      - "13285:13285"
```

## API Endpoints

### GET /metrics

Prometheus metrics endpoint. Returns metrics in Prometheus text format.

### GET /state

Returns the full collected state as JSON.

### GET /health

Returns health status. Returns HTTP 200 if healthy, 503 if unhealthy.

### GET /ready

Returns readiness status. Returns HTTP 200 if ready, 503 if not ready.

### GET /config

Returns the current configuration (sanitized).

## Prometheus Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `aztec_block_height` | Gauge | Current block height |
| `aztec_proven_block_height` | Gauge | Current proven block height |
| `aztec_pending_block_height` | Gauge | Current pending block height |
| `aztec_reference_height` | Gauge | Reference block height from external source |
| `aztec_blocks_behind` | Gauge | Number of blocks behind reference |
| `aztec_blocks_ahead` | Gauge | Number of blocks ahead of reference |
| `aztec_proof_lag` | Gauge | Gap between latest and proven block height |
| `aztec_pending_tx_count` | Gauge | Number of pending transactions |
| `aztec_is_synced` | Gauge | Whether the node is synced (1/0) |
| `aztec_is_syncing` | Gauge | Whether the node is syncing (1/0) |
| `aztec_is_ready` | Gauge | Whether the node is ready (1/0) |
| `aztec_num_peers` | Gauge | Number of connected peers |
| `aztec_base_fee_per_da_gas` | Gauge | Current base fee per DA gas |
| `aztec_base_fee_per_l2_gas` | Gauge | Current base fee per L2 gas |
| `aztec_health_status` | Gauge | Health status (0=ok, 1=warn, 2=error, 3=fatal) |
| `aztec_collection_duration_seconds` | Histogram | Time taken to collect metrics |
| `aztec_collection_errors_total` | Counter | Total collection errors by method |
| `aztec_last_collection_timestamp` | Gauge | Unix timestamp of last collection |

## Aztec Node API Methods Used

The collector queries the following Aztec node JSON-RPC methods:

- `node_getL2Tips` - Get chain tips (latest, pending, proven)
- `node_getNodeInfo` - Get node information
- `node_getWorldStateSyncStatus` - Get sync status
- `node_isReady` - Check if node is ready
- `node_getPendingTxCount` - Get pending transaction count
- `node_getCurrentBaseFees` - Get current base fees

For a complete API reference, see the [Aztec Node API documentation](https://docs.aztec.network/the_aztec_network/reference/node_api_reference).

## Development

### Prerequisites

- Go 1.22+
- Make

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linter
make lint
```

### Project Structure

```
aztec-collector/
├── cmd/
│   └── collector/
│       └── main.go           # Entry point
├── pkg/
│   ├── aztec/
│   │   └── models.go         # Aztec data models
│   ├── collector/
│   │   ├── collector.go      # Main collector logic
│   │   ├── config.go         # Configuration
│   │   ├── metrics.go        # Prometheus metrics
│   │   ├── reference.go      # Reference service client
│   │   ├── rpc.go            # Aztec RPC client wrapper
│   │   └── state.go          # Collector state
│   └── jsonrpc/
│       └── client.go         # JSON-RPC client
├── docker-compose.example.yaml  # Docker Compose example
├── example-config.yaml       # Example configuration
├── prometheus.yml            # Prometheus config example
├── Dockerfile                # Docker build file
├── go.mod                    # Go module file
├── Makefile                  # Build automation
└── README.md                 # This file
```

## Alerting

The collector can send alerts when health thresholds are breached. Alerts are sent via webhooks to:

- **Slack** - Using incoming webhooks
- **Discord** - Using Discord webhooks
- **Telegram** - Using a Telegram bot
- **Generic Webhook** - For custom integrations (PagerDuty, Opsgenie, etc.)

### Alert Types

| Alert | Trigger | Severity |
|-------|---------|----------|
| `node_not_ready` | Node is not ready to serve requests | Error |
| `node_not_synced` | Node is currently syncing | Warning |
| `blocks_behind` | Node is behind reference by more than threshold | Error |
| `blocks_ahead` | Node is ahead of reference by more than threshold | Warning |
| `proof_lag` | Large gap between latest and proven blocks | Warning |
| `recovered` | A previous alert condition has been resolved | Info |

### Alerting Configuration Example

```yaml
alerting:
  enabled: true
  cooldown: 5m
  node_name: "prod-aztec-01"
  
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/TXXXXX/BXXXXX/your-token"
    channel: "#alerts"
    
  discord:
    enabled: true
    webhook_url: "https://discord.com/api/webhooks/your-id/your-token"
    
  telegram:
    enabled: true
    bot_token: "123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
    chat_id: "-1001234567890"
```

### Getting Webhook URLs

**Slack:**
1. Go to [Slack API Apps](https://api.slack.com/apps)
2. Create a new app or select existing
3. Enable "Incoming Webhooks"
4. Add a new webhook to your workspace

**Discord:**
1. Open Server Settings > Integrations
2. Create a new webhook
3. Copy the webhook URL

**Telegram:**
1. Message [@BotFather](https://t.me/botfather) to create a bot
2. Get the bot token
3. Add the bot to your group/channel
4. Get the chat ID using `https://api.telegram.org/bot<TOKEN>/getUpdates`

## Tailing Logs

### Docker Compose

```bash
# Follow logs from the collector
docker compose logs -f aztec-collector

# Follow logs with timestamps
docker compose logs -f --timestamps aztec-collector

# Show last 100 lines and follow
docker compose logs -f --tail 100 aztec-collector

# Follow logs from all services
docker compose logs -f
```

### Docker

```bash
# Follow logs from a running container
docker logs -f aztec-collector

# Show last 50 lines and follow
docker logs -f --tail 50 aztec-collector

# With timestamps
docker logs -f --timestamps aztec-collector
```

### Kubernetes

```bash
# Follow logs from a pod
kubectl logs -f deployment/aztec-collector

# Follow logs with timestamps
kubectl logs -f --timestamps deployment/aztec-collector

# Follow logs from all pods with the label
kubectl logs -f -l app=aztec-collector

# Stream logs from a specific container
kubectl logs -f deployment/aztec-collector -c aztec-collector
```

### systemd (if running as a service)

```bash
# Follow journal logs
journalctl -u aztec-collector -f

# Show last 100 lines and follow
journalctl -u aztec-collector -n 100 -f

# With full output (no truncation)
journalctl -u aztec-collector -f --no-pager
```

### JSON Log File

The collector writes state to a JSON log file that can be monitored:

```bash
# Watch the JSON log file for changes
watch -n 5 cat /path/to/aztec-collector.json

# Use jq to format and follow
tail -f /path/to/aztec-collector.json | jq .

# Check specific fields
watch -n 5 'jq ".health, .chain.headHeight" /path/to/aztec-collector.json'
```

## License

MIT License - see LICENSE file for details.
