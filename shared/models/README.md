# Models Package

Common domain types for Nivo services with India-centric defaults.

## Overview

The `models` package provides fundamental domain types used across all Nivo services, with a focus on fintech-specific types like Money and Currency that require precision and correctness. **INR (Indian Rupee) is the primary currency** for India-centric operations.

## Features

- **Money**: Precise monetary amounts using integer arithmetic (no float precision issues)
- **Currency**: ISO 4217 currency codes with validation
- **Timestamp**: Custom timestamp type with consistent JSON/database serialization

## Money Type

The `Money` type stores monetary amounts in the smallest currency unit (cents) to avoid floating-point precision issues.

### Basic Usage

```go
import "github.com/vnykmshr/nivo/shared/models"

// Create from paise (smallest unit for INR)
money := models.NewMoney(10050, models.INR) // ₹100.50

// Create from float
money := models.NewMoneyFromFloat(100.50, models.INR)

// String representation
fmt.Println(money) // Output: "100.50 INR"

// Convert to float
amount := money.ToFloat() // 100.50

// Using default currency
defaultMoney := models.NewMoney(50000, models.DefaultCurrency) // ₹500.00 (INR is default)
```

### Arithmetic Operations

```go
price := models.NewMoney(100000, models.INR)  // ₹1,000.00
tax := models.NewMoney(18000, models.INR)     // ₹180.00 (18% GST)

// Addition
total, err := price.Add(tax) // ₹1,180.00
if err != nil {
    // Handle currency mismatch error
}

// Subtraction
discount := models.NewMoney(20000, models.INR)
final, err := total.Subtract(discount) // ₹980.00

// Multiplication
double := price.Multiply(2) // ₹2,000.00

// Division
half := price.Divide(2) // ₹500.00
```

### Comparisons

```go
balance := models.NewMoney(1000000, models.INR) // ₹10,000.00
amount := models.NewMoney(500000, models.INR)   // ₹5,000.00

if balance.GreaterThan(amount) {
    fmt.Println("Sufficient funds")
}

if amount.LessThanOrEqual(balance) {
    // Process transaction
}

if money1.Equal(money2) {
    // Same amount and currency
}
```

### Validation

```go
money := models.NewMoney(100000, models.INR) // ₹1,000.00

if money.IsZero() {
    // Handle zero amount
}

if money.IsPositive() {
    // Process positive amount
}

if money.IsNegative() {
    // Handle negative balance
}

// Validate currency
if err := money.Validate(); err != nil {
    // Handle invalid currency
}
```

### JSON Serialization

```go
money := models.NewMoney(10050, models.INR)

// Marshal to JSON
data, _ := json.Marshal(money)
// {"amount":10050,"currency":"INR"}

// Unmarshal from JSON
var decoded models.Money
json.Unmarshal(data, &decoded)
```

## Currency Type

The `Currency` type represents ISO 4217 currency codes.

### Supported Currencies

**Primary Currency (India-centric):**
- INR - Indian Rupee (₹) - **Default currency**

**International Currencies:**
- USD - US Dollar ($)
- EUR - Euro (€)
- GBP - British Pound (£)
- JPY - Japanese Yen (¥)
- CNY - Chinese Yuan (¥)
- CAD - Canadian Dollar (C$)
- AUD - Australian Dollar (A$)
- CHF - Swiss Franc (CHF)
- SGD - Singapore Dollar (S$)

### Usage

```go
// Using constants (INR is the default/primary currency)
currency := models.INR

// Using default currency constant
currency := models.DefaultCurrency // INR

// Parse from string
currency, err := models.ParseCurrency("INR")
if err != nil {
    // Handle invalid currency
}

// Case-insensitive parsing
currency, _ := models.ParseCurrency("inr") // Returns INR

// Validation
if err := currency.Validate(); err != nil {
    // Handle invalid currency
}

// Check support
if currency.IsSupported() {
    // Currency is valid
}

// Get all supported currencies
currencies := models.GetSupportedCurrencies()
```

### Currency Properties

```go
currency := models.INR

// Get symbol
symbol := currency.GetSymbol() // "₹"

// Get decimal places
places := currency.GetDecimalPlaces() // 2 (paise)
// Note: JPY returns 0 (no decimal places)

// String representation
name := currency.String() // "INR"
```

## Timestamp Type

Custom timestamp type with consistent JSON and database serialization.

### Usage

```go
// Create from time.Time
t := time.Now()
ts := models.NewTimestamp(t)

// Get current timestamp
ts := models.Now()

// String representation (ISO 8601)
str := ts.String() // "2025-01-15T10:30:00Z"
```

### Comparisons

```go
ts1 := models.Now()
time.Sleep(1 * time.Second)
ts2 := models.Now()

if ts2.After(ts1) {
    fmt.Println("ts2 is after ts1")
}

if ts1.Before(ts2) {
    fmt.Println("ts1 is before ts2")
}

if ts1.Equal(ts1) {
    fmt.Println("Equal timestamps")
}
```

### JSON Serialization

```go
ts := models.Now()

// Marshal to JSON (ISO 8601 format)
data, _ := json.Marshal(ts)
// "2025-01-15T10:30:00Z"

// Unmarshal from JSON
var decoded models.Timestamp
json.Unmarshal(data, &decoded)

// Zero timestamps serialize as null
zero := models.Timestamp{}
data, _ := json.Marshal(zero) // null
```

### Database Integration

The Timestamp type implements `sql.Scanner` and `driver.Valuer` for seamless database integration:

```go
// Scan from database
var ts models.Timestamp
err := db.QueryRow("SELECT created_at FROM users WHERE id = $1", userID).Scan(&ts)

// Store in database
_, err := db.Exec("INSERT INTO users (name, created_at) VALUES ($1, $2)",
    name, models.Now())
```

## Complete Examples

### India UPI Transfer (Primary Use Case)

```go
package main

import (
    "fmt"
    "github.com/vnykmshr/nivo/shared/models"
)

func UPITransfer(senderBalance, receiverBalance, amount models.Money) error {
    // Validate currencies match (should all be INR for UPI)
    if senderBalance.Currency != models.INR || amount.Currency != models.INR {
        return fmt.Errorf("UPI transfers require INR currency")
    }

    // Check sufficient funds
    if senderBalance.LessThan(amount) {
        return fmt.Errorf("insufficient funds")
    }

    // Validate amount is positive
    if !amount.IsPositive() {
        return fmt.Errorf("amount must be positive")
    }

    // Perform transfer
    newSenderBalance, _ := senderBalance.Subtract(amount)
    newReceiverBalance, _ := receiverBalance.Add(amount)

    fmt.Printf("Sender:   %s -> %s\n", senderBalance, newSenderBalance)
    fmt.Printf("Receiver: %s -> %s\n", receiverBalance, newReceiverBalance)

    return nil
}

func main() {
    // UPI transfer amounts in paise (1 rupee = 100 paise)
    sender := models.NewMoney(1000000, models.INR)   // ₹10,000.00
    receiver := models.NewMoney(500000, models.INR)  // ₹5,000.00
    amount := models.NewMoney(250000, models.INR)    // ₹2,500.00

    if err := UPITransfer(sender, receiver, amount); err != nil {
        fmt.Printf("Transfer failed: %v\n", err)
        return
    }

    fmt.Println("UPI transfer successful!")
    // Output:
    // Sender:   10000.00 INR -> 7500.00 INR
    // Receiver: 5000.00 INR -> 7500.00 INR
    // UPI transfer successful!
}
```

### International Transfer

```go
func InternationalTransfer(from, to models.Money, amount models.Money) error {
    // Validate currencies match
    if from.Currency != amount.Currency {
        return fmt.Errorf("currency mismatch")
    }

    // Check sufficient funds
    if from.LessThan(amount) {
        return fmt.Errorf("insufficient funds")
    }

    // Validate amount is positive
    if !amount.IsPositive() {
        return fmt.Errorf("amount must be positive")
    }

    // Perform transfer
    newFrom, _ := from.Subtract(amount)
    newTo, _ := to.Add(amount)

    fmt.Printf("From: %s -> %s\n", from, newFrom)
    fmt.Printf("To:   %s -> %s\n", to, newTo)

    return nil
}

func main() {
    sender := models.NewMoney(1000000, models.INR)   // ₹10,000.00
    receiver := models.NewMoney(500000, models.INR)  // ₹5,000.00
    amount := models.NewMoney(250000, models.INR)    // ₹2,500.00

    if err := InternationalTransfer(sender, receiver, amount); err != nil {
        fmt.Printf("Transfer failed: %v\n", err)
        return
    }

    fmt.Println("Transfer successful!")
}
```

## Best Practices

### Money

1. **Always use integer storage**: Store amounts in paise (for INR) to avoid float precision issues
   ```go
   // Good (paise for INR)
   money := models.NewMoney(100050, models.INR) // ₹1,000.50

   // Avoid
   amount := 1000.50 // float64 has precision issues
   ```

2. **Use default currency (INR)**: For India-centric operations, use default currency constant
   ```go
   // Good
   money := models.NewMoney(50000, models.DefaultCurrency) // ₹500.00 (INR)

   // Explicit
   money := models.NewMoney(50000, models.INR)
   ```

3. **Validate currency compatibility**: Always check currencies match before operations
   ```go
   result, err := money1.Add(money2)
   if err != nil {
       // Handle currency mismatch
   }
   ```

4. **Use comparison methods**: Don't compare amounts directly
   ```go
   // Good
   if balance.GreaterThan(amount) { }

   // Avoid
   if balance.Amount > amount.Amount { } // Ignores currency
   ```

5. **Handle zero division**: Check divisor before division
   ```go
   if divisor != 0 {
       result := money.Divide(divisor)
   }
   ```

### Currency

1. **Use INR as primary**: Default to INR for India-centric operations
   ```go
   // Good (India-centric)
   currency := models.DefaultCurrency // INR

   // Explicit
   currency := models.INR
   ```

2. **Use constants**: Prefer predefined currency constants
   ```go
   // Good
   currency := models.INR

   // Less ideal
   currency := models.Currency("INR")
   ```

3. **Always validate**: Validate currency before use
   ```go
   currency, err := models.ParseCurrency(input)
   if err != nil {
       return err
   }
   ```

### Timestamp

1. **Use UTC**: Always work with UTC timestamps
   ```go
   ts := models.NewTimestamp(time.Now().UTC())
   ```

2. **Check for zero**: Check if timestamp is set before use
   ```go
   if !ts.IsZero() {
       // Use timestamp
   }
   ```

## Testing

```bash
go test ./shared/models/...
go test -cover ./shared/models/...
```

## Related Packages

- [shared/errors](../errors/README.md) - For validation errors
- [shared/database](../database/README.md) - For database persistence
