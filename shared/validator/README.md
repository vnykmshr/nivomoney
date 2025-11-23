# Validator Package

The validator package provides input validation for Nivo services using [gopantic](https://github.com/vnykmshr/gopantic), combining JSON/YAML parsing with comprehensive validation in a single step.

## Features

- **Parse + Validate**: Single-step JSON/YAML parsing and validation using generics
- **Type Coercion**: Automatic conversion of strings to numbers, booleans, etc.
- **Custom Validators**: Extended with fintech-specific validators
- **Error Integration**: Converts gopantic errors to Nivo's `shared/errors` format
- **Banking Validators**: IBAN, sort codes, routing numbers, account numbers
- **Currency Validation**: ISO 4217 currency code support

## Usage

### Basic Validation

```go
import "github.com/vnykmshr/nivo/shared/validator"

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,gte=18,lte=120"`
}

func HandleCreateUser(data []byte) error {
    user, err := validator.ParseAndValidate[CreateUserRequest](data)
    if err != nil {
        // err is already a *errors.Error with validation details
        return err
    }

    // user is fully parsed and validated
    processUser(user)
    return nil
}
```

### HTTP Handler Integration

```go
func createUser(w http.ResponseWriter, r *http.Request) {
    // Read request body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        response.Error(w, errors.BadRequest("failed to read request body"))
        return
    }

    // Parse and validate in one step
    user, err := validator.ParseAndValidate[CreateUserRequest](body)
    if err != nil {
        response.Error(w, err) // err is already *errors.Error
        return
    }

    // Process valid request...
}
```

### Fintech Validation

```go
type TransferRequest struct {
    FromAccount string `json:"from_account" validate:"required,account_number"`
    ToAccount   string `json:"to_account" validate:"required,account_number"`
    Amount      int64  `json:"amount" validate:"required,money_amount"`
    Currency    string `json:"currency" validate:"required,currency"`
}

transfer, err := validator.ParseAndValidate[TransferRequest](requestBody)
```

## Available Validators

### Built-in (from gopantic)

- `required` - Field must have a non-zero value
- `min=n` - Minimum value (numeric) or length (string/array)
- `max=n` - Maximum value (numeric) or length (string/array)
- `email` - Valid email address format
- `length=n` - Exact length for strings/arrays
- `alpha` - Only alphabetic characters
- `alphanum` - Only alphanumeric characters

### Standard Validators (custom)

- `gt=n` - Greater than n (numeric comparison)
- `gte=n` - Greater than or equal to n
- `lt=n` - Less than n
- `lte=n` - Less than or equal to n
- `oneof=a b c` - Value must be one of the specified options
- `uuid` - Valid UUID format (lowercase, with hyphens)
- `url` - Valid URL (http/https schemes only)
- `numeric` - String contains only digits [0-9]

### Fintech Validators (custom)

- `currency` - ISO 4217 currency code (INR, USD, EUR, GBP, etc.)
- `money_amount` - Positive amount in cents/smallest unit (>0)
- `account_number` - 10-20 alphanumeric characters (uppercase A-Z, 0-9)
- `iban` - International Bank Account Number (15-34 chars, 2-letter country + 2-digit check)
- `sort_code` - UK bank sort code (6 digits, hyphens optional: 123456 or 12-34-56)
- `routing_number` - US bank routing number (exactly 9 digits)

### India-Specific Validators (custom)

**Primary validators for India-centric fintech operations:**

- `ifsc` - IFSC code (11 chars: 4 bank code + 0 + 6 branch code, e.g., SBIN0001234)
- `upi_id` - UPI ID (username@bankcode format, e.g., user@paytm)
- `pan` - PAN card (10 chars: 5 letters + 4 digits + 1 letter, e.g., ABCDE1234F)
- `aadhaar` - Aadhaar number (12 digits, cannot start with 0 or 1)
- `indian_phone` - Indian mobile (+91 + 10 digits starting with 6-9, e.g., +919876543210)
- `pincode` - Indian PIN code (6 digits, cannot start with 0, e.g., 560001)

## Type Coercion

gopantic automatically coerces types when possible:

```go
// Input JSON: {"age": "25", "active": "true", "balance": "100.50"}
// Output struct: User{Age: 25, Active: true, Balance: 100.50}

type User struct {
    Age     int     `json:"age"`
    Active  bool    `json:"active"`
    Balance float64 `json:"balance"`
}
```

Supported coercions:
- String → int/int64/float64
- String → bool ("true"/"false")
- Number → string

## Error Handling

Validation errors are returned as `*errors.Error` with structured details:

```go
user, err := validator.ParseAndValidate[User](data)
if err != nil {
    // err is *errors.Error with:
    //   Code: ErrCodeValidation
    //   Message: validation error message
    //   Details: {"field": "fieldName", "value": invalidValue}
    return err
}
```

Example error response:

```json
{
  "code": "VALIDATION_ERROR",
  "message": "must be a valid email address",
  "details": {
    "field": "email",
    "value": "invalid-email"
  }
}
```

## Complete Examples

### India KYC Verification (Primary Use Case)

```go
type KYCRequest struct {
    FullName string `json:"full_name" validate:"required,min=2,max=100"`
    PAN      string `json:"pan" validate:"required,pan"`
    Aadhaar  string `json:"aadhaar" validate:"required,aadhaar"`
    Phone    string `json:"phone" validate:"required,indian_phone"`
    Address  struct {
        Street  string `json:"street" validate:"required,min=5,max=200"`
        City    string `json:"city" validate:"required,min=2,max=100"`
        State   string `json:"state" validate:"required,min=2,max=100"`
        PIN     string `json:"pin" validate:"required,pincode"`
    } `json:"address"`
}

func kycVerificationHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    kyc, err := validator.ParseAndValidate[KYCRequest](body)
    if err != nil {
        response.Error(w, err)
        return
    }

    // All fields are validated:
    // - PAN is valid format (ABCDE1234F)
    // - Aadhaar is 12 digits, doesn't start with 0 or 1
    // - Phone is valid Indian mobile (+919876543210)
    // - PIN is valid 6-digit Indian postal code

    verifyKYC(kyc)
}
```

Example request:
```json
{
    "full_name": "Rajesh Kumar",
    "pan": "ABCDE1234F",
    "aadhaar": "234567890123",
    "phone": "+919876543210",
    "address": {
        "street": "123 MG Road",
        "city": "Bangalore",
        "state": "Karnataka",
        "pin": "560001"
    }
}
```

### UPI Payment Request

```go
type UPIPaymentRequest struct {
    FromUPI string `json:"from_upi" validate:"required,upi_id"`
    ToUPI   string `json:"to_upi" validate:"required,upi_id"`
    Amount  int64  `json:"amount" validate:"required,money_amount"`
}

func upiPaymentHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    payment, err := validator.ParseAndValidate[UPIPaymentRequest](body)
    if err != nil {
        response.Error(w, err)
        return
    }

    // Both UPI IDs are validated (username@bankcode format)
    // Amount is positive (in paise)

    processUPIPayment(payment)
}
```

Example request:
```json
{
    "from_upi": "rajesh@paytm",
    "to_upi": "merchant@okaxis",
    "amount": 50000
}
```
**Note:** Amount is in paise (50000 paise = ₹500)

### Bank Transfer with IFSC

```go
type BankTransferRequest struct {
    FromAccount string `json:"from_account" validate:"required,account_number"`
    ToAccount   string `json:"to_account" validate:"required,account_number"`
    ToIFSC      string `json:"to_ifsc" validate:"required,ifsc"`
    Amount      int64  `json:"amount" validate:"required,money_amount"`
    Currency    string `json:"currency" validate:"required,currency"`
}

func bankTransferHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    transfer, err := validator.ParseAndValidate[BankTransferRequest](body)
    if err != nil {
        response.Error(w, err)
        return
    }

    // IFSC is validated (SBIN0001234 format)
    // Currency defaults to INR for India-centric operations

    processBankTransfer(transfer)
}
```

Example request:
```json
{
    "from_account": "1234567890",
    "to_account": "0987654321",
    "to_ifsc": "SBIN0001234",
    "amount": 100000,
    "currency": "INR"
}
```
**Note:** Amount ₹1,000 (100000 paise = ₹1,000)

### International Transfer Request

```go
type TransferRequest struct {
    FromAccount string `json:"from_account" validate:"required,account_number"`
    ToAccount   string `json:"to_account" validate:"required,account_number"`
    Amount      int64  `json:"amount" validate:"required,money_amount"`
    Currency    string `json:"currency" validate:"required,currency"`
    Reference   string `json:"reference" validate:"max=100"`
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    transfer, err := validator.ParseAndValidate[TransferRequest](body)
    if err != nil {
        response.Error(w, err)
        return
    }

    // transfer.Amount is guaranteed to be > 0
    // transfer.Currency is guaranteed to be supported
    // transfer.FromAccount/ToAccount are valid account numbers

    processTransfer(transfer)
}
```

### User Registration

```go
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=100"`
    FullName string `json:"full_name" validate:"required,min=2,max=100"`
    Age      int    `json:"age" validate:"required,gte=18,lte=120"`
    Country  string `json:"country" validate:"required,length=2,alpha"`
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    req, err := validator.ParseAndValidate[RegisterRequest](body)
    if err != nil {
        response.Error(w, err)
        return
    }

    createUser(req)
}
```

### Payment Method

```go
type PaymentMethod struct {
    Type          string `json:"type" validate:"required,oneof=card bank_transfer"`
    IBAN          string `json:"iban,omitempty" validate:"iban"`
    SortCode      string `json:"sort_code,omitempty" validate:"sort_code"`
    RoutingNumber string `json:"routing_number,omitempty" validate:"routing_number"`
}

func addPaymentMethod(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    pm, err := validator.ParseAndValidate[PaymentMethod](body)
    if err != nil {
        response.Error(w, err)
        return
    }

    savePaymentMethod(pm)
}
```

## Validator Details

### Currency Validator

Validates ISO 4217 currency codes using `models.Currency.IsSupported()`:

```go
type Payment struct {
    Currency string `json:"currency" validate:"currency"`
}

// Valid: USD, EUR, GBP, JPY, CNY, INR, CAD, AUD, CHF, SGD
// Invalid: XXX, INVALID, US
// Case-insensitive but preserves original case
```

### Money Amount Validator

Validates positive monetary amounts in cents/smallest currency unit:

```go
type Payment struct {
    Amount int64 `json:"amount" validate:"money_amount"`
}

// Valid: 1, 100, 1000, 999999
// Invalid: 0, -1, -100
// Error message: "amount must be positive (in cents)"
```

### Account Number Validator

Validates bank account numbers (10-20 uppercase alphanumeric):

```go
type Account struct {
    Number string `json:"number" validate:"account_number"`
}

// Valid: "1234567890", "ABC1234567890", "1234567890ABCDEFGHIJ"
// Invalid: "123" (too short), "12345678901234567890123" (too long)
//         "123456789a" (lowercase), "12345-6789" (special chars)
```

### IBAN Validator

Validates International Bank Account Number format:

```go
type BankAccount struct {
    IBAN string `json:"iban" validate:"iban"`
}

// Valid: "GB82WEST12345698765432", "DE89370400440532013000"
// Invalid: "GB" (too short), "1234567890" (no country), "GBAA1234" (invalid check)
// Format: 2 letters (country) + 2 digits (check) + up to 30 alphanumeric
// Note: Format validation only, checksum not validated
```

### Sort Code Validator (UK)

Validates UK bank sort codes (6 digits, hyphens optional):

```go
type UKAccount struct {
    SortCode string `json:"sort_code" validate:"sort_code"`
}

// Valid: "123456", "12-34-56"
// Invalid: "12345" (too short), "1234567" (too long), "12-34-5A" (letter)
```

### Routing Number Validator (US)

Validates US bank routing numbers (exactly 9 digits):

```go
type USAccount struct {
    RoutingNumber string `json:"routing_number" validate:"routing_number"`
}

// Valid: "021000021", "111000025", "026009593"
// Invalid: "12345678" (too short), "1234567890" (too long), "02100002A" (letter)
```

### IFSC Code Validator (India - Primary)

Validates Indian Financial System Code (11 characters):

```go
type BankAccount struct {
    IFSC string `json:"ifsc" validate:"ifsc"`
}

// Valid: "SBIN0001234", "HDFC0000001", "ICIC0001234", "AXIS0000001"
// Invalid: "SBIN001234" (too short), "SBIN1001234" (5th char not 0)
//          "123400012345" (starts with digits), "SBIN-001234" (special char)
// Format: 4 bank code (letters) + 0 (reserved) + 6 branch code (alphanumeric)
```

### UPI ID Validator (India - Primary)

Validates Unified Payments Interface ID:

```go
type Payment struct {
    UPI string `json:"upi" validate:"upi_id"`
}

// Valid: "user@paytm", "john.doe@okaxis", "9876543210@paytm"
// Invalid: "user" (no @), "@paytm" (no username), "user@PAYTM" (uppercase bank)
// Format: username (alphanumeric, dots, hyphens, underscores) @ bankcode (lowercase)
```

### PAN Card Validator (India - Primary)

Validates Permanent Account Number (mandatory for financial transactions):

```go
type User struct {
    PAN string `json:"pan" validate:"pan"`
}

// Valid: "ABCDE1234F", "AAAAA0000A"
// Invalid: "abcde1234f" (lowercase), "ABCDE1234" (too short), "12345ABCDE" (wrong format)
// Format: 5 letters (uppercase) + 4 digits + 1 letter (uppercase)
// Note: Must be uppercase, case-sensitive validation
```

### Aadhaar Validator (India)

Validates Aadhaar unique identification number:

```go
type User struct {
    Aadhaar string `json:"aadhaar" validate:"aadhaar"`
}

// Valid: "234567890123", "987654321098", "2345 6789 0123" (spaces allowed)
// Invalid: "1234567890123" (starts with 1), "0123456789012" (starts with 0)
//          "ABCD56789012" (contains letters)
// Format: 12 digits, cannot start with 0 or 1, spaces are automatically removed
```

### Indian Phone Number Validator (India)

Validates Indian mobile numbers with country code:

```go
type User struct {
    Phone string `json:"phone" validate:"indian_phone"`
}

// Valid: "+919876543210", "+91-9876543210", "+91 98765 43210"
// Invalid: "9876543210" (no +91), "+911234567890" (starts with 1-5)
//          "+9198765432" (too short), "+91987654321012" (too long)
// Format: +91 + 10 digits starting with 6-9 (hyphens and spaces allowed)
```

### PIN Code Validator (India)

Validates Indian Postal Index Number:

```go
type Address struct {
    PIN string `json:"pin" validate:"pincode"`
}

// Valid: "560001" (Bangalore), "110001" (Delhi), "400001" (Mumbai)
// Invalid: "056001" (starts with 0), "56001" (too short), "56A001" (contains letter)
// Format: 6 digits, cannot start with 0
```

## Testing

```bash
cd shared/validator
go test -v
go test -v -cover
```

Test coverage: 77.2%

## Implementation Notes

### General
- All custom validators (standard, fintech, India) are registered globally in `init()`
- Empty values skip validation (handled by `required` validator)
- Validators return structured `*errors.Error` with field and value details

### India Validators (Primary)
- **IFSC**: Validates 11-character format (BANK0BRANCH), case-insensitive
- **UPI ID**: Username allows alphanumeric + dots/hyphens/underscores, bank code must be lowercase
- **PAN**: Must be uppercase (case-sensitive validation), 10 characters
- **Aadhaar**: Automatically removes spaces, cannot start with 0 or 1
- **Indian Phone**: Accepts various formats with hyphens/spaces, validates 6-9 starting digit
- **PIN Code**: 6 digits, cannot start with 0

### International Validators
- Currency validation is case-insensitive but preserves original case (INR is primary)
- IBAN validation checks format but not checksum
- Account numbers must be uppercase alphanumeric only
- Sort codes (UK) and routing numbers (US) must be numeric only

## Integration with shared/response

Works seamlessly with `shared/response` for consistent error handling:

```go
import (
    "github.com/vnykmshr/nivo/shared/response"
    "github.com/vnykmshr/nivo/shared/validator"
)

func handler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    req, err := validator.ParseAndValidate[Request](body)
    if err != nil {
        response.Error(w, err) // Automatically formatted
        return
    }

    // Process valid request...
}
```

## Related Packages

- [`shared/errors`](../errors/README.md) - Error types and codes
- [`shared/response`](../response/README.md) - API response formats
- [`shared/models`](../models/README.md) - Currency and Money types
- [`gopantic`](https://github.com/vnykmshr/gopantic) - Underlying validation library

## License

Copyright © 2025 Nivo
