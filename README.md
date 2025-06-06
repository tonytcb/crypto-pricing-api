# Crypto Pricing API

A real-time cryptocurrency price streaming service that fetches price data from CoinDesk API and broadcasts it to clients using Server-Sent Events (SSE).

## Features

- Real-time price updates for cryptocurrency pairs (default: BTCUSD)
- Server-Sent Events (SSE) for efficient client streaming
- Configurable polling intervals and retry mechanisms
- In-memory storage with configurable capacity
- Clean architecture with dependency injection

## Running in Development Environment

### Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local development without Docker)

### Using Docker

1. Clone the repository
2. Start the application:
   ```
   make up
   ```
3. Access the API at http://localhost:8080

### Local Development

1. Clone the repository
2. Create a `.env` file in the root directory or use the default configuration in `config/default.env`
3. Run the application:
   ```
   go run cmd/main.go
   ```

### Configuration

The application uses environment variables for configuration. You can set these in:
- `.env` file in the root directory (highest priority)
- `config/default.env` file (default values)

Key configuration parameters:

```
# General configurations
ENV=development
LOG_LEVEL=info
REST_API_PORT=:8080

# Price monitoring configurations
PAIR_PRICE_TO_MONITOR=BTCUSD
STORE_MAX_ITEMS=1000
PRICES_PULLING_INTERVAL=5s
PRICES_CHANNEL_BUFFER_SIZE=100

# SSE configurations
SSE_CLIENTS_BUFFER_SIZE=100
SSE_CLIENTS_CLEAN_UP_INTERVAL=30s

# CoinDesk HTTP Client configuration
COIN_DESK_API_URL=https://min-api.cryptocompare.com/data/price
COIN_DESK_RETRY_MAX_ATTEMPTS=3
COIN_DESK_CLIENT_TIMEOUT=3s
```

## Architecture

The application follows a clean architecture approach with dependency injection for better testability and component replaceability:

### Key Components

1. **Application Layer** (`internal/app`)
   - Orchestrates the application components
   - Manages the lifecycle of services

2. **API Layer** (`internal/api`)
   - HTTP server and handlers
   - Exposes endpoints for price streaming

3. **Domain Layer** (`internal/domain`)
   - Core business entities (Currency, Pair, Price)
   - Business logic independent of external frameworks

4. **Infrastructure Layer** (`internal/infra`)
   - External services integration (CoinDesk API)
   - Event providers and listeners
   - Storage implementations
   - SSE implementation for client streaming

### Dependency Injection

The application uses constructor-based dependency injection:
- Components receive their dependencies through constructors
- No global state or service locators
- Easy to replace implementations for testing or future changes

### Testing

The codebase includes unit and integration tests. Run them with:

```
make tests
```

## Example Client

The repository includes an HTML client example (`client-example.html`) that demonstrates how to connect to the API and receive real-time price updates.

To use it:
1. Start the application
2. Open `client-example.html` in a web browser
3. Click "Connect" to start receiving price updates

This example was built with an AI assistant and is intended to demonstrate the API's capabilities.