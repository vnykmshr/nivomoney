package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vnykmshr/nivo/shared/errors"
)

func TestJSON(t *testing.T) {
	t.Run("writes JSON response", func(t *testing.T) {
		rec := httptest.NewRecorder()

		data := map[string]string{"message": "hello"}
		err := JSON(rec, http.StatusOK, data)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if rec.Header().Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type application/json")
		}

		var result map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if result["message"] != "hello" {
			t.Errorf("expected message 'hello', got %s", result["message"])
		}
	})

	t.Run("handles different status codes", func(t *testing.T) {
		codes := []int{200, 201, 400, 404, 500}

		for _, code := range codes {
			rec := httptest.NewRecorder()
			JSON(rec, code, map[string]string{"test": "data"})

			if rec.Code != code {
				t.Errorf("expected status %d, got %d", code, rec.Code)
			}
		}
	})
}

func TestSuccess(t *testing.T) {
	t.Run("writes success response", func(t *testing.T) {
		rec := httptest.NewRecorder()

		data := map[string]string{"key": "value"}
		err := Success(rec, http.StatusOK, data)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var response Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if !response.Success {
			t.Error("expected success to be true")
		}

		if response.Error != nil {
			t.Error("expected error to be nil")
		}

		dataMap := response.Data.(map[string]interface{})
		if dataMap["key"] != "value" {
			t.Error("expected data to be preserved")
		}
	})
}

func TestSuccessWithMeta(t *testing.T) {
	t.Run("includes metadata", func(t *testing.T) {
		rec := httptest.NewRecorder()

		data := map[string]string{"key": "value"}
		meta := &Meta{
			RequestID: "test-123",
			Timestamp: "2024-01-01T00:00:00Z",
		}

		err := SuccessWithMeta(rec, http.StatusOK, data, meta)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var response Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if !response.Success {
			t.Error("expected success to be true")
		}

		if response.Meta == nil {
			t.Fatal("expected meta to be set")
		}

		if response.Meta.RequestID != "test-123" {
			t.Errorf("expected request ID 'test-123', got %s", response.Meta.RequestID)
		}
	})
}

func TestError(t *testing.T) {
	t.Run("writes error response", func(t *testing.T) {
		rec := httptest.NewRecorder()

		err := errors.NotFound("user")
		Error(rec, err)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}

		var response Response
		if jsonErr := json.Unmarshal(rec.Body.Bytes(), &response); jsonErr != nil {
			t.Fatalf("failed to unmarshal response: %v", jsonErr)
		}

		if response.Success {
			t.Error("expected success to be false")
		}

		if response.Error == nil {
			t.Fatal("expected error to be set")
		}

		if response.Error.Code != string(errors.ErrCodeNotFound) {
			t.Errorf("expected code %s, got %s", errors.ErrCodeNotFound, response.Error.Code)
		}

		if response.Error.Message != "user not found" {
			t.Errorf("expected message 'user not found', got %s", response.Error.Message)
		}
	})

	t.Run("includes error details", func(t *testing.T) {
		rec := httptest.NewRecorder()

		err := errors.BadRequest("invalid input")
		err.Details = map[string]interface{}{
			"field": "email",
			"error": "invalid format",
		}

		Error(rec, err)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response.Error.Details == nil {
			t.Fatal("expected details to be set")
		}

		if response.Error.Details["field"] != "email" {
			t.Error("expected details to be preserved")
		}
	})
}

func TestErrorWithMeta(t *testing.T) {
	t.Run("includes metadata with error", func(t *testing.T) {
		rec := httptest.NewRecorder()

		err := errors.Internal("database error")
		meta := &Meta{
			RequestID: "error-123",
		}

		ErrorWithMeta(rec, err, meta)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response.Success {
			t.Error("expected success to be false")
		}

		if response.Error == nil {
			t.Fatal("expected error to be set")
		}

		if response.Meta == nil {
			t.Fatal("expected meta to be set")
		}

		if response.Meta.RequestID != "error-123" {
			t.Error("expected meta to be preserved")
		}
	})
}

func TestBadRequest(t *testing.T) {
	rec := httptest.NewRecorder()

	BadRequest(rec, "invalid request")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var response Response
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response.Success {
		t.Error("expected success to be false")
	}

	if response.Error.Message != "invalid request" {
		t.Error("expected message to match")
	}
}

func TestUnauthorized(t *testing.T) {
	rec := httptest.NewRecorder()

	Unauthorized(rec, "invalid token")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}

	var response Response
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response.Error.Code != string(errors.ErrCodeUnauthorized) {
		t.Error("expected unauthorized code")
	}
}

func TestForbidden(t *testing.T) {
	rec := httptest.NewRecorder()

	Forbidden(rec, "insufficient permissions")

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}

	var response Response
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response.Error.Code != string(errors.ErrCodeForbidden) {
		t.Error("expected forbidden code")
	}
}

func TestNotFound(t *testing.T) {
	rec := httptest.NewRecorder()

	NotFound(rec, "account")

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	var response Response
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response.Error.Message != "account not found" {
		t.Errorf("expected message 'account not found', got %s", response.Error.Message)
	}
}

func TestConflict(t *testing.T) {
	rec := httptest.NewRecorder()

	Conflict(rec, "email already exists")

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}

	var response Response
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response.Error.Code != string(errors.ErrCodeConflict) {
		t.Error("expected conflict code")
	}
}

func TestInternalError(t *testing.T) {
	rec := httptest.NewRecorder()

	InternalError(rec, "database connection failed")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var response Response
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response.Error.Code != string(errors.ErrCodeInternal) {
		t.Error("expected internal error code")
	}
}

func TestCreated(t *testing.T) {
	rec := httptest.NewRecorder()

	data := map[string]interface{}{
		"id":   "123",
		"name": "John",
	}

	Created(rec, data)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var response Response
	json.Unmarshal(rec.Body.Bytes(), &response)

	if !response.Success {
		t.Error("expected success to be true")
	}
}

func TestOK(t *testing.T) {
	rec := httptest.NewRecorder()

	data := map[string]string{"status": "ok"}

	OK(rec, data)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response Response
	json.Unmarshal(rec.Body.Bytes(), &response)

	if !response.Success {
		t.Error("expected success to be true")
	}
}

func TestNoContent(t *testing.T) {
	rec := httptest.NewRecorder()

	NoContent(rec)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}

	if rec.Body.Len() != 0 {
		t.Error("expected empty body")
	}
}

func TestPaginated(t *testing.T) {
	t.Run("calculates pagination correctly", func(t *testing.T) {
		rec := httptest.NewRecorder()

		data := []map[string]string{
			{"id": "1"},
			{"id": "2"},
		}

		Paginated(rec, data, 2, 10, 35)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		if !response.Success {
			t.Error("expected success to be true")
		}

		if response.Meta == nil || response.Meta.Pagination == nil {
			t.Fatal("expected pagination to be set")
		}

		p := response.Meta.Pagination

		if p.Page != 2 {
			t.Errorf("expected page 2, got %d", p.Page)
		}

		if p.PageSize != 10 {
			t.Errorf("expected page size 10, got %d", p.PageSize)
		}

		if p.TotalItems != 35 {
			t.Errorf("expected total items 35, got %d", p.TotalItems)
		}

		if p.TotalPages != 4 {
			t.Errorf("expected total pages 4, got %d", p.TotalPages)
		}

		if !p.HasNext {
			t.Error("expected has_next to be true")
		}

		if !p.HasPrev {
			t.Error("expected has_prev to be true")
		}
	})

	t.Run("first page", func(t *testing.T) {
		rec := httptest.NewRecorder()

		Paginated(rec, []int{1, 2}, 1, 10, 25)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		p := response.Meta.Pagination

		if p.HasPrev {
			t.Error("expected has_prev to be false on first page")
		}

		if !p.HasNext {
			t.Error("expected has_next to be true")
		}
	})

	t.Run("last page", func(t *testing.T) {
		rec := httptest.NewRecorder()

		Paginated(rec, []int{21, 22}, 3, 10, 25)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		p := response.Meta.Pagination

		if !p.HasPrev {
			t.Error("expected has_prev to be true")
		}

		if p.HasNext {
			t.Error("expected has_next to be false on last page")
		}
	})

	t.Run("single page", func(t *testing.T) {
		rec := httptest.NewRecorder()

		Paginated(rec, []int{1, 2, 3}, 1, 10, 3)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		p := response.Meta.Pagination

		if p.TotalPages != 1 {
			t.Errorf("expected total pages 1, got %d", p.TotalPages)
		}

		if p.HasPrev || p.HasNext {
			t.Error("expected no prev/next on single page")
		}
	})

	t.Run("empty results", func(t *testing.T) {
		rec := httptest.NewRecorder()

		Paginated(rec, []int{}, 1, 10, 0)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		p := response.Meta.Pagination

		if p.TotalPages != 1 {
			t.Errorf("expected total pages 1 for empty results, got %d", p.TotalPages)
		}

		if p.TotalItems != 0 {
			t.Errorf("expected total items 0, got %d", p.TotalItems)
		}
	})

	t.Run("exact page boundary", func(t *testing.T) {
		rec := httptest.NewRecorder()

		Paginated(rec, []int{1, 2}, 1, 10, 20)

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		p := response.Meta.Pagination

		if p.TotalPages != 2 {
			t.Errorf("expected total pages 2, got %d", p.TotalPages)
		}
	})
}

func TestValidationError(t *testing.T) {
	t.Run("writes validation error with details", func(t *testing.T) {
		rec := httptest.NewRecorder()

		details := map[string]interface{}{
			"email":    "invalid email format",
			"password": "must be at least 8 characters",
		}

		ValidationError(rec, details)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response.Success {
			t.Error("expected success to be false")
		}

		if response.Error == nil {
			t.Fatal("expected error to be set")
		}

		if response.Error.Code != string(errors.ErrCodeValidation) {
			t.Error("expected validation error code")
		}

		if response.Error.Details == nil {
			t.Fatal("expected details to be set")
		}

		if response.Error.Details["email"] != "invalid email format" {
			t.Error("expected email validation error")
		}

		if response.Error.Details["password"] != "must be at least 8 characters" {
			t.Error("expected password validation error")
		}
	})
}

func TestResponseStructure(t *testing.T) {
	t.Run("success response structure", func(t *testing.T) {
		rec := httptest.NewRecorder()

		Success(rec, http.StatusOK, map[string]string{"key": "value"})

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		if !response.Success {
			t.Error("expected success field to be true")
		}

		if response.Data == nil {
			t.Error("expected data field to be set")
		}

		if response.Error != nil {
			t.Error("expected error field to be nil")
		}
	})

	t.Run("error response structure", func(t *testing.T) {
		rec := httptest.NewRecorder()

		Error(rec, errors.BadRequest("test error"))

		var response Response
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response.Success {
			t.Error("expected success field to be false")
		}

		if response.Data != nil {
			t.Error("expected data field to be nil")
		}

		if response.Error == nil {
			t.Error("expected error field to be set")
		}

		if response.Error.Code == "" {
			t.Error("expected error code to be set")
		}

		if response.Error.Message == "" {
			t.Error("expected error message to be set")
		}
	})
}
