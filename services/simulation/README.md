# Simulation Service

The Simulation Service generates realistic demo traffic for the Nivo platform. It simulates multiple user personas with different transaction patterns to demonstrate the system's capabilities.

## Features

- **User Personas**: Six distinct user behavior patterns
- **Realistic Traffic**: Time-based activity with configurable patterns
- **Auto-Start**: Optional automatic simulation on service start
- **Gateway Integration**: Uses the API Gateway for all operations
- **Controllable**: Start/stop simulation via API

## API Endpoints

### Get Simulation Status
```http
GET /api/v1/simulation/status
```

**Response:**
```json
{
  "running": true,
  "message": "Simulation is running"
}
```

### Start Simulation
```http
POST /api/v1/simulation/start
```

**Response:**
```json
{
  "message": "simulation started"
}
```

### Stop Simulation
```http
POST /api/v1/simulation/stop
```

**Response:**
```json
{
  "message": "simulation stopped"
}
```

### Health Check
```http
GET /health
```

## User Personas

The simulation engine supports six distinct user personas:

### Frequent Trader
High-frequency small transactions, typical day trader behavior.

| Property | Value |
|----------|-------|
| Frequency | Every 2 minutes |
| Amount Range | ₹10 - ₹500 |
| Transaction Mix | 60% transfer, 30% deposit, 10% withdrawal |
| Active Hours | 9 AM - 8 PM |
| Min Balance | ₹1,000 |

### Saver
Conservative user who primarily deposits with rare withdrawals.

| Property | Value |
|----------|-------|
| Frequency | Once per day |
| Amount Range | ₹500 - ₹10,000 |
| Transaction Mix | 80% deposit, 15% transfer, 5% withdrawal |
| Active Hours | 10-11 AM, 2-3 PM, 7-8 PM |
| Min Balance | ₹5,000 |

### Bill Payer
Regular user making scheduled payments.

| Property | Value |
|----------|-------|
| Frequency | Every 6 hours |
| Amount Range | ₹500 - ₹5,000 |
| Transaction Mix | 70% transfer, 25% deposit, 5% withdrawal |
| Active Hours | 9-11 AM, 5-7 PM |
| Min Balance | ₹2,000 |

### Shopper
Frequent payments user, typical e-commerce customer.

| Property | Value |
|----------|-------|
| Frequency | Every 4 hours |
| Amount Range | ₹100 - ₹2,000 |
| Transaction Mix | 75% transfer, 20% deposit, 5% withdrawal |
| Active Hours | 10 AM - 9 PM |
| Min Balance | ₹1,500 |

### Investor
Large deposit user making strategic transfers.

| Property | Value |
|----------|-------|
| Frequency | Every 48 hours |
| Amount Range | ₹5,000 - ₹1,00,000 |
| Transaction Mix | 50% deposit, 45% transfer, 5% withdrawal |
| Active Hours | 10-11 AM, 2-4 PM |
| Min Balance | ₹1,00,000 |

### Casual
Random sporadic activity typical of occasional users.

| Property | Value |
|----------|-------|
| Frequency | Every 12 hours |
| Amount Range | ₹50 - ₹1,000 |
| Transaction Mix | 50% transfer, 40% deposit, 10% withdrawal |
| Active Hours | 9 AM - 10 PM |
| Min Balance | ₹500 |

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | HTTP server port | 8086 |
| `GATEWAY_URL` | API Gateway URL | http://gateway:8000 |
| `ADMIN_TOKEN` | Admin JWT for API calls | (required) |
| `AUTO_START_SIMULATION` | Start simulation on boot | true |
| `DATABASE_PASSWORD` | PostgreSQL password | (required) |
| `JWT_SECRET` | JWT validation secret | (required) |

### Auto-Start Behavior

When `AUTO_START_SIMULATION=true`:
1. Service waits 10 seconds for other services to initialize
2. Loads seeded users from database
3. Assigns random personas to users
4. Begins generating traffic based on persona schedules

## Setup

### Prerequisites

- Go 1.23+
- PostgreSQL 14+
- Running Gateway service
- Seeded user accounts (via Seed Service)
- Valid admin JWT token

### Running the Service

```bash
# Set required environment
export GATEWAY_URL=http://localhost:8000
export ADMIN_TOKEN=eyJhbGciOiJIUzI1NiIs...

# Run
cd services/simulation
go run cmd/server/main.go
```

### Docker Compose

```yaml
simulation:
  build:
    context: .
    dockerfile: services/simulation/Dockerfile
  environment:
    - GATEWAY_URL=http://gateway:8000
    - ADMIN_TOKEN=${ADMIN_TOKEN}
    - AUTO_START_SIMULATION=true
    - DATABASE_PASSWORD=${DATABASE_PASSWORD}
    - JWT_SECRET=${JWT_SECRET}
  depends_on:
    - gateway
    - seed
```

## Architecture

```
services/simulation/
├── cmd/
│   └── server/          # Server entry point
├── internal/
│   ├── handler/         # HTTP handlers
│   │   └── simulation_handler.go
│   ├── service/         # Core logic
│   │   ├── simulator.go
│   │   ├── user_lifecycle.go
│   │   └── gateway_client.go
│   └── personas/        # User behavior patterns
│       └── personas.go
├── Makefile
└── README.md
```

## How It Works

1. **User Selection**: Loads seeded users from database
2. **Persona Assignment**: Randomly assigns a persona to each user
3. **Activity Loop**: For each user:
   - Check if current hour is in persona's active hours
   - Check if enough time passed since last transaction
   - Check if balance meets minimum threshold
   - Generate appropriate transaction based on weights
4. **Auto-Deposit**: If balance drops below threshold, generates a deposit
5. **Gateway Calls**: All operations go through Gateway API

## Monitoring

The simulation generates logs for each action:

```
[simulation] Started simulation with 5 users
[simulation] [user-123] Persona: frequent_trader, Balance: ₹10,000
[simulation] [user-123] Generated transfer: ₹250 to user-456
[simulation] [user-789] Balance low (₹800), generating deposit
[simulation] [user-789] Generated deposit: ₹5,000
```

## Future Enhancements

- [ ] Web UI for simulation control
- [ ] Real-time metrics dashboard
- [ ] Configurable persona parameters via API
- [ ] Load testing mode
- [ ] Scheduled simulation windows
