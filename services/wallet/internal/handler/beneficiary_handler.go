package handler

import (
	"io"
	"net/http"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/services/wallet/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// BeneficiaryHandler handles HTTP requests for beneficiary operations.
type BeneficiaryHandler struct {
	beneficiaryService *service.BeneficiaryService
}

// NewBeneficiaryHandler creates a new beneficiary handler.
func NewBeneficiaryHandler(beneficiaryService *service.BeneficiaryService) *BeneficiaryHandler {
	return &BeneficiaryHandler{
		beneficiaryService: beneficiaryService,
	}
}

// AddBeneficiary handles POST /api/v1/beneficiaries
func (h *BeneficiaryHandler) AddBeneficiary(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.AddBeneficiaryRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	beneficiary, createErr := h.beneficiaryService.AddBeneficiary(r.Context(), userID.(string), &req)
	if createErr != nil {
		response.Error(w, createErr)
		return
	}

	response.Created(w, models.ToBeneficiaryResponse(beneficiary))
}

// GetBeneficiary handles GET /api/v1/beneficiaries/:id
func (h *BeneficiaryHandler) GetBeneficiary(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	beneficiaryID := r.PathValue("id")
	if beneficiaryID == "" {
		response.Error(w, errors.BadRequest("beneficiary ID is required"))
		return
	}

	beneficiary, err := h.beneficiaryService.GetBeneficiary(r.Context(), userID.(string), beneficiaryID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, models.ToBeneficiaryResponse(beneficiary))
}

// ListBeneficiaries handles GET /api/v1/beneficiaries
func (h *BeneficiaryHandler) ListBeneficiaries(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	beneficiaries, err := h.beneficiaryService.ListBeneficiaries(r.Context(), userID.(string))
	if err != nil {
		response.Error(w, err)
		return
	}

	// Convert to response format
	responses := make([]*models.BeneficiaryResponse, len(beneficiaries))
	for i, b := range beneficiaries {
		responses[i] = models.ToBeneficiaryResponse(b)
	}

	response.OK(w, responses)
}

// UpdateBeneficiary handles PUT /api/v1/beneficiaries/:id
func (h *BeneficiaryHandler) UpdateBeneficiary(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	beneficiaryID := r.PathValue("id")
	if beneficiaryID == "" {
		response.Error(w, errors.BadRequest("beneficiary ID is required"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.UpdateBeneficiaryRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	beneficiary, updateErr := h.beneficiaryService.UpdateBeneficiary(r.Context(), userID.(string), beneficiaryID, &req)
	if updateErr != nil {
		response.Error(w, updateErr)
		return
	}

	response.OK(w, models.ToBeneficiaryResponse(beneficiary))
}

// DeleteBeneficiary handles DELETE /api/v1/beneficiaries/:id
func (h *BeneficiaryHandler) DeleteBeneficiary(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	beneficiaryID := r.PathValue("id")
	if beneficiaryID == "" {
		response.Error(w, errors.BadRequest("beneficiary ID is required"))
		return
	}

	if err := h.beneficiaryService.DeleteBeneficiary(r.Context(), userID.(string), beneficiaryID); err != nil {
		response.Error(w, err)
		return
	}

	response.NoContent(w)
}
