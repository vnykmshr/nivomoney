---
layout: default
title: SSE Integration
nav_order: 7
description: "Real-time event streaming with Server-Sent Events"
permalink: /sse-integration
---

# SSE (Server-Sent Events) Integration

## Overview

The Nivo platform now has real-time event streaming capabilities using Server-Sent Events (SSE). Services can broadcast events that flow through the Gateway's SSE broker to connected clients in real-time.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Event Flow                                │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Service (Transaction/Wallet/Identity)                        │
│        ↓                                                      │
│  Event Publisher (shared/events/publisher.go)                │
│        ↓ HTTP POST                                            │
│  Gateway /api/v1/events/broadcast                            │
│        ↓                                                      │
│  SSE Broker (shared/events/broker.go)                        │
│        ↓                                                      │
│  Connected Clients (GET /api/v1/events)                      │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. Event Broker (`shared/events/broker.go`)
- Manages SSE client connections
- Handles topic-based subscriptions
- Broadcasts events to subscribed clients
- Automatic heartbeat every 30 seconds

### 2. Event Publisher (`shared/events/publisher.go`)
- Shared library for services to publish events
- HTTP-based communication with Gateway
- Async publishing (fire-and-forget)
- Helper methods for different event types

### 3. Gateway SSE Handler (`gateway/internal/handler/sse.go`)
- **GET /api/v1/events** - Subscribe to event stream
- **POST /api/v1/events/broadcast** - Publish events (internal)
- **GET /api/v1/events/stats** - Broker statistics

## Events Published

### Transaction Service
| Event Type | Topic | Trigger |
|------------|-------|---------|
| `transaction.created` | `transactions` | Transfer/Deposit/Withdrawal created |

**Event Data:**
- transaction_id
- type (transfer/deposit/withdrawal)
- status
- amount
- currency
- source_wallet_id (if applicable)
- destination_wallet_id (if applicable)
- description

### Wallet Service
| Event Type | Topic | Trigger |
|------------|-------|---------|
| `wallet.created` | `wallets` | New wallet created |
| `wallet.status_changed` | `wallets` | Wallet activated/frozen/unfrozen/closed |

**Event Data:**
- wallet_id
- user_id
- type
- currency
- status
- balance
- available_balance
- action (activated/frozen/unfrozen/closed)
- old_status / new_status (for status changes)

### Identity Service
| Event Type | Topic | Trigger |
|------------|-------|---------|
| `user.registered` | `users` | New user signs up |
| `user.kyc_updated` | `users` | KYC submitted/verified/rejected |
| `user.status_changed` | `users` | User status changes (pending→active) |

**Event Data:**
- user_id
- email
- phone
- full_name
- status
- kyc_status
- rejection_reason (if applicable)

## Usage

### Subscribing to Events (Client Side)

```javascript
// Subscribe to all events
const eventSource = new EventSource('http://localhost:8000/api/v1/events');

// Subscribe to specific topic
const eventSource = new EventSource('http://localhost:8000/api/v1/events?topics=transactions');

// Handle events
eventSource.addEventListener('transaction.created', (e) => {
    const data = JSON.parse(e.data);
    console.log('New transaction:', data);
});

eventSource.addEventListener('wallet.created', (e) => {
    const data = JSON.parse(e.data);
    console.log('New wallet:', data);
});

eventSource.addEventListener('user.registered', (e) => {
    const data = JSON.parse(e.data);
    console.log('New user:', data);
});
```

### Publishing Events (Service Side)

```go
// Initialize publisher (in main.go)
eventPublisher := events.NewPublisher(events.PublishConfig{
    GatewayURL:  "http://gateway:8000",
    ServiceName: "transaction",
})

// Publish event
eventPublisher.PublishTransactionEvent("transaction.created", txnID, map[string]interface{}{
    "type":   "transfer",
    "amount": 10000,
    "status": "pending",
})
```

## Configuration

Services use the `GATEWAY_URL` environment variable to connect to the Gateway:

```bash
GATEWAY_URL=http://gateway:8000
```

Default: `http://gateway:8000`

## Topics

| Topic | Description |
|-------|-------------|
| `transactions` | Transaction-related events |
| `wallets` | Wallet-related events |
| `users` | User/Identity-related events |
| `risk` | Risk alerts and events |
| `all` | Special topic - receives all events |

## Testing

### 1. Start an SSE listener:
```bash
curl -N http://localhost:8000/api/v1/events?topics=transactions
```

### 2. Create a transaction:
```bash
curl -X POST http://localhost:8000/api/v1/transactions/deposit \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_id": "wallet-id",
    "amount_paise": 10000,
    "description": "Test deposit"
  }'
```

### 3. Observe the event in the SSE stream

## Future Enhancements

- [ ] Add `transaction.completed` and `transaction.failed` events
- [ ] Add `wallet.balance_updated` events
- [ ] Add Risk Service events
- [ ] Add event replay/history capabilities
- [ ] Add event filtering by user_id
- [ ] Add authentication for SSE connections
- [ ] Add rate limiting for event publishing
- [ ] Add metrics for event throughput

## Troubleshooting

### Events not appearing?

1. Check Gateway logs:
   ```bash
   docker logs nivo-gateway
   ```

2. Check service logs:
   ```bash
   docker logs nivo-transaction
   ```

3. Verify Gateway URL in service configuration

4. Check SSE broker stats:
   ```bash
   curl http://localhost:8000/api/v1/events/stats
   ```

### Connection drops?

- SSE connections send heartbeat every 30 seconds
- Clients should auto-reconnect on connection loss
- Check for proxy/load balancer timeouts
