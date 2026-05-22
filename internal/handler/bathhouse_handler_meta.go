package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/seo"
)

// GetMeta godoc
//
//	@Summary		Get bathhouse SEO meta
//	@Description	Return generated SEO meta tags for a bathhouse.
//	@Tags			bathhouses
//	@Produce		json
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=seo.MetaTags}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bathhouses/{id}/meta [get]
func (h *BathhouseHandler) GetMeta(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	bh, err := h.bathhouseService.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	meta := h.buildMeta(r.Context(), bh)
	if meta == nil {
		meta = &seo.MetaTags{}
	}

	writeJSON(w, http.StatusOK, meta)
}

func (h *BathhouseHandler) buildMeta(ctx context.Context, bh *domain.Bathhouse) *seo.MetaTags {
	var cityName, citySlug string
	if h.cityService != nil && bh.CityID > 0 {
		city, err := h.cityService.GetByID(ctx, bh.CityID)
		if err == nil && city != nil {
			cityName = city.Name
			citySlug = city.Slug
		}
	}

	meta := seo.GenerateMetaTags(seo.MetaInput{
		Name:         bh.Name,
		CityName:     cityName,
		CitySlug:     citySlug,
		Slug:         bh.Slug,
		Description:  bh.Description,
		PricePerHour: bh.PricePerHour,
		Rating:       bh.Rating,
		ReviewCount:  bh.ReviewCount,
		Images:       bh.Images,
		HasPool:      bh.HasPool,
		HasSauna:     bh.HasSauna,
		HasSteamRoom: bh.HasSteamRoom,
		HasHotTub:    bh.HasHotTub,
		HasBBQ:       bh.HasBBQ,
		HasKaraoke:   bh.HasKaraoke,
		BaseURL:      h.baseURL,
	})
	return &meta
}

//	@Summary		Check listing completeness
//	@Description	Check how complete a bathhouse listing is before submitting for moderation.
//	@Tags			bathhouses
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=service.CompletenessResult}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/completeness [get]

func (h *BathhouseHandler) recordBathhouseView(r *http.Request, bathhouseID uuid.UUID, userID uuid.UUID) {
	ctx := r.Context()

	if err := h.bathhouseService.IncrementViewCount(ctx, bathhouseID); err != nil {
		h.log.Warn("failed to increment view count", "bathhouse_id", bathhouseID, "error", err)
	}

	if userID != uuid.Nil && h.recommendationService != nil {
		if err := h.recommendationService.RecordView(ctx, userID, bathhouseID); err != nil {
			h.log.Warn("failed to record recommendation view", "bathhouse_id", bathhouseID, "error", err)
		}
	}

	if h.analyticsService != nil {
		ipHash := getIPHash(r)
		source := domain.ViewSourceDirect
		var viewerID *uuid.UUID
		if userID != uuid.Nil {
			viewerID = &userID
		}
		if err := h.analyticsService.RecordView(ctx, bathhouseID, viewerID, source, ipHash); err != nil {
			h.log.Warn("failed to record analytics view", "bathhouse_id", bathhouseID, "error", err)
		}
	}

	if h.promotionService != nil {
		if err := h.promotionService.RecordClick(ctx, bathhouseID); err != nil {
			h.log.Warn("failed to record promotion click", "bathhouse_id", bathhouseID, "error", err)
		}
	}

	if userID != uuid.Nil && h.savedSearchService != nil {
		if err := h.savedSearchService.RecordView(ctx, userID, bathhouseID); err != nil {
			h.log.Warn("failed to record recently viewed", "bathhouse_id", bathhouseID, "error", err)
		}
	}
}

func (h *BathhouseHandler) buildOwnerProfile(ctx context.Context, ownerID uuid.UUID) *ownerProfileResponse {
	if h.userService == nil {
		return nil
	}

	profile, err := h.userService.GetPublicProfile(ctx, ownerID)
	if err != nil || profile == nil {
		return nil
	}

	objectCount := 0
	if h.bathhouseService != nil {
		result, err := h.bathhouseService.ListByOwner(ctx, ownerID, 1, 1)
		if err == nil && result != nil {
			objectCount = int(result.TotalCount)
		}
	}

	return &ownerProfileResponse{
		ID:          profile.ID.String(),
		Name:        profile.Name,
		AvatarURL:   profile.AvatarURL,
		Rating:      profile.AvgRating,
		ObjectCount: objectCount,
		MemberSince: profile.MemberSince,
	}
}

const maxPageSize = 100

func getPage(s string) int {
	if s == "" {
		return 1
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return 1
	}
	return v
}

func getPageSize(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return defaultVal
	}
	if v > maxPageSize {
		return maxPageSize
	}
	return v
}
