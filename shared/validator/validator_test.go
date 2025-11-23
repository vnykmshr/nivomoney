package validator

import (
	"fmt"
	"testing"

	"github.com/vnykmshr/nivo/shared/errors"
)

func TestParseAndValidate(t *testing.T) {
	t.Run("valid struct passes validation", func(t *testing.T) {
		type User struct {
			Name  string `json:"name" validate:"required,min=3"`
			Email string `json:"email" validate:"required,email"`
			Age   int    `json:"age" validate:"required,gte=18"`
		}

		data := []byte(`{
			"name": "John Doe",
			"email": "john@example.com",
			"age": 25
		}`)

		user, err := ParseAndValidate[User](data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if user.Name != "John Doe" {
			t.Errorf("expected name 'John Doe', got %s", user.Name)
		}
	})

	t.Run("type coercion works", func(t *testing.T) {
		type User struct {
			Age int `json:"age" validate:"required,gte=18"`
		}

		data := []byte(`{"age": "25"}`) // String that should coerce to int

		user, err := ParseAndValidate[User](data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if user.Age != 25 {
			t.Errorf("expected age 25, got %d", user.Age)
		}
	})

	t.Run("returns validation error", func(t *testing.T) {
		type User struct {
			Email string `json:"email" validate:"required,email"`
		}

		data := []byte(`{"email": "invalid-email"}`)

		_, err := ParseAndValidate[User](data)
		if err == nil {
			t.Fatal("expected validation error")
		}

		validationErr, ok := err.(*errors.Error)
		if !ok {
			t.Fatal("expected errors.Error type")
		}

		if validationErr.Code != errors.ErrCodeValidation {
			t.Error("expected validation error code")
		}
	})
}

func TestStandardValidators(t *testing.T) {
	t.Run("gt validator", func(t *testing.T) {
		type Product struct {
			Price float64 `json:"price" validate:"gt=0"`
		}

		// Valid
		data := []byte(`{"price": 10.5}`)
		_, err := ParseAndValidate[Product](data)
		if err != nil {
			t.Errorf("expected valid, got error: %v", err)
		}

		// Invalid
		data = []byte(`{"price": 0}`)
		_, err = ParseAndValidate[Product](data)
		if err == nil {
			t.Error("expected validation error for price=0")
		}
	})

	t.Run("gte validator", func(t *testing.T) {
		type User struct {
			Age int `json:"age" validate:"gte=18"`
		}

		// Valid - equal
		data := []byte(`{"age": 18}`)
		_, err := ParseAndValidate[User](data)
		if err != nil {
			t.Errorf("expected valid, got error: %v", err)
		}

		// Valid - greater
		data = []byte(`{"age": 25}`)
		_, err = ParseAndValidate[User](data)
		if err != nil {
			t.Errorf("expected valid, got error: %v", err)
		}

		// Invalid
		data = []byte(`{"age": 17}`)
		_, err = ParseAndValidate[User](data)
		if err == nil {
			t.Error("expected validation error for age=17")
		}
	})

	t.Run("lt validator", func(t *testing.T) {
		type Limit struct {
			Value int `json:"value" validate:"lt=100"`
		}

		// Valid
		data := []byte(`{"value": 99}`)
		_, err := ParseAndValidate[Limit](data)
		if err != nil {
			t.Errorf("expected valid, got error: %v", err)
		}

		// Invalid
		data = []byte(`{"value": 100}`)
		_, err = ParseAndValidate[Limit](data)
		if err == nil {
			t.Error("expected validation error for value=100")
		}
	})

	t.Run("lte validator", func(t *testing.T) {
		type Limit struct {
			Value int `json:"value" validate:"lte=100"`
		}

		// Valid - equal
		data := []byte(`{"value": 100}`)
		_, err := ParseAndValidate[Limit](data)
		if err != nil {
			t.Errorf("expected valid, got error: %v", err)
		}

		// Valid - less
		data = []byte(`{"value": 50}`)
		_, err = ParseAndValidate[Limit](data)
		if err != nil {
			t.Errorf("expected valid, got error: %v", err)
		}

		// Invalid
		data = []byte(`{"value": 101}`)
		_, err = ParseAndValidate[Limit](data)
		if err == nil {
			t.Error("expected validation error for value=101")
		}
	})

	t.Run("oneof validator", func(t *testing.T) {
		type Order struct {
			Status string `json:"status" validate:"oneof=pending processing completed"`
		}

		validStatuses := []string{"pending", "processing", "completed"}
		for _, status := range validStatuses {
			data := []byte(`{"status": "` + status + `"}`)
			_, err := ParseAndValidate[Order](data)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", status, err)
			}
		}

		// Invalid
		data := []byte(`{"status": "invalid"}`)
		_, err := ParseAndValidate[Order](data)
		if err == nil {
			t.Error("expected validation error for invalid status")
		}
	})

	t.Run("uuid validator", func(t *testing.T) {
		type Resource struct {
			ID string `json:"id" validate:"uuid"`
		}

		// Valid UUIDs
		validUUIDs := []string{
			"550e8400-e29b-41d4-a716-446655440000",
			"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		}

		for _, uuid := range validUUIDs {
			data := []byte(`{"id": "` + uuid + `"}`)
			_, err := ParseAndValidate[Resource](data)
			if err != nil {
				t.Errorf("expected %s to be valid UUID, got error: %v", uuid, err)
			}
		}

		// Invalid UUIDs
		invalidUUIDs := []string{
			"not-a-uuid",
			"550e8400-e29b-41d4-a716",
			"550e8400-e29b-41d4-a716-446655440000-extra",
		}

		for _, uuid := range invalidUUIDs {
			data := []byte(`{"id": "` + uuid + `"}`)
			_, err := ParseAndValidate[Resource](data)
			if err == nil {
				t.Errorf("expected %s to be invalid UUID", uuid)
			}
		}
	})

	t.Run("url validator", func(t *testing.T) {
		type Website struct {
			URL string `json:"url" validate:"url"`
		}

		// Valid URLs
		validURLs := []string{
			"https://example.com",
			"http://example.com/path",
			"https://sub.example.com/path?query=1",
		}

		for _, url := range validURLs {
			data := []byte(`{"url": "` + url + `"}`)
			_, err := ParseAndValidate[Website](data)
			if err != nil {
				t.Errorf("expected %s to be valid URL, got error: %v", url, err)
			}
		}

		// Invalid URLs
		invalidURLs := []string{
			"not-a-url",
			"ftp://example.com",
			"example.com",
		}

		for _, url := range invalidURLs {
			data := []byte(`{"url": "` + url + `"}`)
			_, err := ParseAndValidate[Website](data)
			if err == nil {
				t.Errorf("expected %s to be invalid URL", url)
			}
		}
	})

	t.Run("numeric validator", func(t *testing.T) {
		type Code struct {
			Value string `json:"value" validate:"numeric"`
		}

		// Valid
		data := []byte(`{"value": "12345"}`)
		_, err := ParseAndValidate[Code](data)
		if err != nil {
			t.Errorf("expected valid, got error: %v", err)
		}

		// Invalid
		data = []byte(`{"value": "12a45"}`)
		_, err = ParseAndValidate[Code](data)
		if err == nil {
			t.Error("expected validation error for alphanumeric string")
		}
	})
}

func TestFintechValidators(t *testing.T) {
	t.Run("currency validator", func(t *testing.T) {
		type Transfer struct {
			Currency string `json:"currency" validate:"currency"`
		}

		// Valid currencies
		validCurrencies := []string{"USD", "EUR", "GBP", "JPY"}
		for _, curr := range validCurrencies {
			data := []byte(`{"currency": "` + curr + `"}`)
			_, err := ParseAndValidate[Transfer](data)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", curr, err)
			}
		}

		// Case insensitive
		data := []byte(`{"currency": "usd"}`)
		transfer, err := ParseAndValidate[Transfer](data)
		if err != nil {
			t.Errorf("expected lowercase currency to be valid, got error: %v", err)
		}
		if transfer.Currency != "usd" {
			t.Error("currency should preserve original case")
		}

		// Invalid currencies
		invalidCurrencies := []string{"XXX", "INVALID", "US"}
		for _, curr := range invalidCurrencies {
			data := []byte(`{"currency": "` + curr + `"}`)
			_, err := ParseAndValidate[Transfer](data)
			if err == nil {
				t.Errorf("expected %s to be invalid", curr)
			}
		}
	})

	t.Run("money_amount validator", func(t *testing.T) {
		type Payment struct {
			Amount int64 `json:"amount" validate:"money_amount"`
		}

		// Valid amounts
		validAmounts := []int64{1, 100, 1000, 999999}
		for _, amount := range validAmounts {
			data := []byte(fmt.Sprintf(`{"amount": %d}`, amount))
			_, err := ParseAndValidate[Payment](data)
			if err != nil {
				t.Errorf("expected %d to be valid, got error: %v", amount, err)
			}
		}

		// Invalid amounts
		invalidAmounts := []int64{0, -1, -100}
		for _, amount := range invalidAmounts {
			data := []byte(fmt.Sprintf(`{"amount": %d}`, amount))
			_, err := ParseAndValidate[Payment](data)
			if err == nil {
				t.Errorf("expected %d to be invalid", amount)
			}
		}
	})

	t.Run("account_number validator", func(t *testing.T) {
		type Account struct {
			Number string `json:"number" validate:"account_number"`
		}

		// Valid account numbers
		validNumbers := []string{
			"1234567890",
			"ABC1234567890",
			"1234567890ABCDEFGHIJ",
		}

		for _, num := range validNumbers {
			data := []byte(`{"number": "` + num + `"}`)
			_, err := ParseAndValidate[Account](data)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", num, err)
			}
		}

		// Invalid account numbers
		invalidNumbers := []string{
			"123",                     // Too short
			"12345678901234567890123", // Too long
			"123456789a",              // Lowercase
			"12345-6789",              // Special characters
		}

		for _, num := range invalidNumbers {
			data := []byte(`{"number": "` + num + `"}`)
			_, err := ParseAndValidate[Account](data)
			if err == nil {
				t.Errorf("expected %s to be invalid", num)
			}
		}
	})

	t.Run("iban validator", func(t *testing.T) {
		type BankAccount struct {
			IBAN string `json:"iban" validate:"iban"`
		}

		// Valid IBANs
		validIBANs := []string{
			"GB82WEST12345698765432",
			"DE89370400440532013000",
			"FR1420041010050500013M02606",
		}

		for _, iban := range validIBANs {
			data := []byte(`{"iban": "` + iban + `"}`)
			_, err := ParseAndValidate[BankAccount](data)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", iban, err)
			}
		}

		// Invalid IBANs
		invalidIBANs := []string{
			"GB",         // Too short
			"1234567890", // No country code
			"GBAA1234",   // Invalid check digits
		}

		for _, iban := range invalidIBANs {
			data := []byte(`{"iban": "` + iban + `"}`)
			_, err := ParseAndValidate[BankAccount](data)
			if err == nil {
				t.Errorf("expected %s to be invalid", iban)
			}
		}
	})

	t.Run("sort_code validator", func(t *testing.T) {
		type UKAccount struct {
			SortCode string `json:"sort_code" validate:"sort_code"`
		}

		// Valid sort codes
		validCodes := []string{
			"123456",
			"12-34-56",
		}

		for _, code := range validCodes {
			data := []byte(`{"sort_code": "` + code + `"}`)
			_, err := ParseAndValidate[UKAccount](data)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", code, err)
			}
		}

		// Invalid sort codes
		invalidCodes := []string{
			"12345",    // Too short
			"1234567",  // Too long
			"12-34-5A", // Contains letter
		}

		for _, code := range invalidCodes {
			data := []byte(`{"sort_code": "` + code + `"}`)
			_, err := ParseAndValidate[UKAccount](data)
			if err == nil {
				t.Errorf("expected %s to be invalid", code)
			}
		}
	})

	t.Run("routing_number validator", func(t *testing.T) {
		type USAccount struct {
			RoutingNumber string `json:"routing_number" validate:"routing_number"`
		}

		// Valid routing numbers
		validNumbers := []string{
			"021000021",
			"111000025",
			"026009593",
		}

		for _, num := range validNumbers {
			data := []byte(`{"routing_number": "` + num + `"}`)
			_, err := ParseAndValidate[USAccount](data)
			if err != nil {
				t.Errorf("expected %s to be valid, got error: %v", num, err)
			}
		}

		// Invalid routing numbers
		invalidNumbers := []string{
			"12345678",   // Too short
			"1234567890", // Too long
			"02100002A",  // Contains letter
		}

		for _, num := range invalidNumbers {
			data := []byte(`{"routing_number": "` + num + `"}`)
			_, err := ParseAndValidate[USAccount](data)
			if err == nil {
				t.Errorf("expected %s to be invalid", num)
			}
		}
	})
}

func TestComplexValidation(t *testing.T) {
	t.Run("multiple field validation", func(t *testing.T) {
		type CreateUserRequest struct {
			Name   string `json:"name" validate:"required,min=2,max=50"`
			Email  string `json:"email" validate:"required,email"`
			Age    int    `json:"age" validate:"required,gte=18,lte=120"`
			Status string `json:"status" validate:"oneof=active inactive pending"`
		}

		// Valid request
		data := []byte(`{
			"name": "John Doe",
			"email": "john@example.com",
			"age": 25,
			"status": "active"
		}`)

		user, err := ParseAndValidate[CreateUserRequest](data)
		if err != nil {
			t.Fatalf("expected valid, got error: %v", err)
		}

		if user.Name != "John Doe" {
			t.Error("name not parsed correctly")
		}
	})

	t.Run("transfer request validation", func(t *testing.T) {
		type TransferRequest struct {
			FromAccount string `json:"from_account" validate:"required,account_number"`
			ToAccount   string `json:"to_account" validate:"required,account_number"`
			Amount      int64  `json:"amount" validate:"required,money_amount"`
			Currency    string `json:"currency" validate:"required,currency"`
		}

		// Valid transfer
		data := []byte(`{
			"from_account": "1234567890",
			"to_account": "0987654321",
			"amount": 10000,
			"currency": "USD"
		}`)

		transfer, err := ParseAndValidate[TransferRequest](data)
		if err != nil {
			t.Fatalf("expected valid, got error: %v", err)
		}

		if transfer.Amount != 10000 {
			t.Error("amount not parsed correctly")
		}
	})
}

func TestIndiaValidators(t *testing.T) {
	t.Run("ifsc validator", func(t *testing.T) {
		type BankAccount struct {
			IFSC string `json:"ifsc" validate:"ifsc"`
		}

		// Valid IFSC codes
		validCodes := []string{
			"SBIN0001234",
			"HDFC0000001",
			"ICIC0001234",
			"AXIS0000001",
		}

		for _, code := range validCodes {
			data := []byte(fmt.Sprintf(`{"ifsc": "%s"}`, code))
			_, err := ParseAndValidate[BankAccount](data)
			if err != nil {
				t.Errorf("expected %s to be valid IFSC, got error: %v", code, err)
			}
		}

		// Invalid IFSC codes
		invalidCodes := []string{
			"SBIN001234",   // Too short
			"SBIN00012345", // Too long
			"SBIN1001234",  // 5th char not 0
			"123400012345", // Starts with digits
			"SBIN-001234",  // Special character
		}

		for _, code := range invalidCodes {
			data := []byte(fmt.Sprintf(`{"ifsc": "%s"}`, code))
			_, err := ParseAndValidate[BankAccount](data)
			if err == nil {
				t.Errorf("expected %s to be invalid IFSC", code)
			}
		}
	})

	t.Run("upi_id validator", func(t *testing.T) {
		type Payment struct {
			UPI string `json:"upi" validate:"upi_id"`
		}

		// Valid UPI IDs
		validUPIs := []string{
			"user@paytm",
			"john.doe@okaxis",
			"test_user@ybl",
			"9876543210@paytm",
			"user-name@icici",
		}

		for _, upi := range validUPIs {
			data := []byte(fmt.Sprintf(`{"upi": "%s"}`, upi))
			_, err := ParseAndValidate[Payment](data)
			if err != nil {
				t.Errorf("expected %s to be valid UPI ID, got error: %v", upi, err)
			}
		}

		// Invalid UPI IDs
		invalidUPIs := []string{
			"user",        // No @
			"@paytm",      // No username
			"user@",       // No bank code
			"user@@paytm", // Multiple @
			"user@PAYTM",  // Uppercase bank code
			"user@pay tm", // Space in bank code
		}

		for _, upi := range invalidUPIs {
			data := []byte(fmt.Sprintf(`{"upi": "%s"}`, upi))
			_, err := ParseAndValidate[Payment](data)
			if err == nil {
				t.Errorf("expected %s to be invalid UPI ID", upi)
			}
		}
	})

	t.Run("pan validator", func(t *testing.T) {
		type User struct {
			PAN string `json:"pan" validate:"pan"`
		}

		// Valid PAN numbers
		validPANs := []string{
			"ABCDE1234F",
			"AAAAA0000A",
			"ZZZZZ9999Z",
		}

		for _, pan := range validPANs {
			data := []byte(fmt.Sprintf(`{"pan": "%s"}`, pan))
			_, err := ParseAndValidate[User](data)
			if err != nil {
				t.Errorf("expected %s to be valid PAN, got error: %v", pan, err)
			}
		}

		// Invalid PAN numbers
		invalidPANs := []string{
			"ABCDE1234",   // Too short
			"ABCDE12345",  // Too long
			"12345ABCDE",  // Wrong format
			"ABCDE1234F1", // Extra character
			"abcde1234f",  // Lowercase
		}

		for _, pan := range invalidPANs {
			data := []byte(fmt.Sprintf(`{"pan": "%s"}`, pan))
			_, err := ParseAndValidate[User](data)
			if err == nil {
				t.Errorf("expected %s to be invalid PAN", pan)
			}
		}
	})

	t.Run("aadhaar validator", func(t *testing.T) {
		type User struct {
			Aadhaar string `json:"aadhaar" validate:"aadhaar"`
		}

		// Valid Aadhaar numbers
		validAadhaar := []string{
			"234567890123",
			"987654321098",
			"2345 6789 0123", // With spaces
		}

		for _, aadhaar := range validAadhaar {
			data := []byte(fmt.Sprintf(`{"aadhaar": "%s"}`, aadhaar))
			_, err := ParseAndValidate[User](data)
			if err != nil {
				t.Errorf("expected %s to be valid Aadhaar, got error: %v", aadhaar, err)
			}
		}

		// Invalid Aadhaar numbers
		invalidAadhaar := []string{
			"12345678901",   // Too short
			"1234567890123", // Starts with 1
			"0123456789012", // Starts with 0
			"ABCD56789012",  // Contains letters
			"12345678901A",  // Contains letter
		}

		for _, aadhaar := range invalidAadhaar {
			data := []byte(fmt.Sprintf(`{"aadhaar": "%s"}`, aadhaar))
			_, err := ParseAndValidate[User](data)
			if err == nil {
				t.Errorf("expected %s to be invalid Aadhaar", aadhaar)
			}
		}
	})

	t.Run("indian_phone validator", func(t *testing.T) {
		type User struct {
			Phone string `json:"phone" validate:"indian_phone"`
		}

		// Valid phone numbers
		validPhones := []string{
			"+919876543210",
			"+91-9876543210",
			"+91 98765 43210",
			"+916789012345",
			"+917890123456",
			"+918901234567",
		}

		for _, phone := range validPhones {
			data := []byte(fmt.Sprintf(`{"phone": "%s"}`, phone))
			_, err := ParseAndValidate[User](data)
			if err != nil {
				t.Errorf("expected %s to be valid Indian phone, got error: %v", phone, err)
			}
		}

		// Invalid phone numbers
		invalidPhones := []string{
			"9876543210",      // No +91
			"+911234567890",   // Starts with 1
			"+915678901234",   // Starts with 5
			"+9198765432",     // Too short
			"+91987654321012", // Too long
			"+91-ABCD567890",  // Contains letters
		}

		for _, phone := range invalidPhones {
			data := []byte(fmt.Sprintf(`{"phone": "%s"}`, phone))
			_, err := ParseAndValidate[User](data)
			if err == nil {
				t.Errorf("expected %s to be invalid Indian phone", phone)
			}
		}
	})

	t.Run("pincode validator", func(t *testing.T) {
		type Address struct {
			PIN string `json:"pin" validate:"pincode"`
		}

		// Valid PIN codes
		validPINs := []string{
			"560001", // Bangalore
			"110001", // Delhi
			"400001", // Mumbai
			"700001", // Kolkata
			"600001", // Chennai
		}

		for _, pin := range validPINs {
			data := []byte(fmt.Sprintf(`{"pin": "%s"}`, pin))
			_, err := ParseAndValidate[Address](data)
			if err != nil {
				t.Errorf("expected %s to be valid PIN code, got error: %v", pin, err)
			}
		}

		// Invalid PIN codes
		invalidPINs := []string{
			"56001",   // Too short
			"5600012", // Too long
			"056001",  // Starts with 0
			"56A001",  // Contains letter
			"5600-1",  // Contains special char
		}

		for _, pin := range invalidPINs {
			data := []byte(fmt.Sprintf(`{"pin": "%s"}`, pin))
			_, err := ParseAndValidate[Address](data)
			if err == nil {
				t.Errorf("expected %s to be invalid PIN code", pin)
			}
		}
	})

	t.Run("complete India KYC validation", func(t *testing.T) {
		type KYCRequest struct {
			FullName string `json:"full_name" validate:"required,min=2,max=100"`
			PAN      string `json:"pan" validate:"required,pan"`
			Aadhaar  string `json:"aadhaar" validate:"required,aadhaar"`
			Phone    string `json:"phone" validate:"required,indian_phone"`
			PIN      string `json:"pin" validate:"required,pincode"`
		}

		// Valid KYC data
		data := []byte(`{
			"full_name": "Rajesh Kumar",
			"pan": "ABCDE1234F",
			"aadhaar": "234567890123",
			"phone": "+919876543210",
			"pin": "560001"
		}`)

		kyc, err := ParseAndValidate[KYCRequest](data)
		if err != nil {
			t.Fatalf("expected valid KYC, got error: %v", err)
		}

		if kyc.FullName != "Rajesh Kumar" {
			t.Error("full_name not parsed correctly")
		}

		if kyc.PAN != "ABCDE1234F" {
			t.Error("PAN not parsed correctly")
		}
	})
}
