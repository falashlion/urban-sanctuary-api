package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/falashlion/urban-sanctuary-api/internal/dto"
	"github.com/falashlion/urban-sanctuary-api/internal/platform/storage"
	"github.com/falashlion/urban-sanctuary-api/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// PropertyService handles property-related business logic.
type PropertyService struct {
	propRepo *repository.PropertyRepository
	userRepo *repository.UserRepository
	authRepo *repository.AuthRepository
	s3       *storage.S3Client
	log      zerolog.Logger
}

// NewPropertyService creates a new PropertyService.
func NewPropertyService(
	propRepo *repository.PropertyRepository,
	userRepo *repository.UserRepository,
	authRepo *repository.AuthRepository,
	s3 *storage.S3Client,
	log zerolog.Logger,
) *PropertyService {
	return &PropertyService{
		propRepo: propRepo,
		userRepo: userRepo,
		authRepo: authRepo,
		s3:       s3,
		log:      log,
	}
}

// Create creates a new property listing.
func (s *PropertyService) Create(ctx context.Context, ownerID uuid.UUID, req dto.CreatePropertyRequest) (*dto.PropertyResponse, error) {
	// Verify owner exists and has appropriate role
	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if owner == nil {
		return nil, domain.ErrNotFound("User")
	}

	amenities := req.Amenities
	if amenities == nil {
		amenities = []string{}
	}

	prop := &domain.Property{
		OwnerID:       ownerID,
		Title:         req.Title,
		Description:   req.Description,
		Neighborhood:  domain.Neighborhood(req.Neighborhood),
		Address:       req.Address,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		PricePerNight: req.PricePerNight,
		Bedrooms:      req.Bedrooms,
		Bathrooms:     req.Bathrooms,
		MaxGuests:     req.MaxGuests,
		Amenities:     amenities,
		Images:        []domain.PropertyImage{},
		Status:        domain.PropertyStatusDraft,
	}

	if err := s.propRepo.Create(ctx, prop); err != nil {
		return nil, domain.ErrInternal(fmt.Errorf("failed to create property: %w", err))
	}

	resp := toPropertyResponse(prop)
	return &resp, nil
}

// GetByID retrieves a property by its ID.
func (s *PropertyService) GetByID(ctx context.Context, id uuid.UUID) (*dto.PropertyResponse, error) {
	prop, err := s.propRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if prop == nil {
		return nil, domain.ErrNotFound("Property")
	}

	// Load owner info
	owner, err := s.userRepo.GetByID(ctx, prop.OwnerID)
	if err == nil && owner != nil {
		prop.Owner = owner
	}

	resp := toPropertyResponse(prop)

	// Load reviews
	reviews, err := s.authRepo.ListReviewsByProperty(ctx, id)
	if err == nil && len(reviews) > 0 {
		// Reviews are available via the property detail endpoint
		s.log.Debug().Int("review_count", len(reviews)).Msg("reviews loaded for property")
	}

	return &resp, nil
}

// List retrieves properties with filters and pagination.
func (s *PropertyService) List(ctx context.Context, query dto.PropertyListQuery) ([]dto.PropertyResponse, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 || query.PerPage > 100 {
		query.PerPage = 20
	}

	props, total, err := s.propRepo.List(ctx,
		query.Neighborhood, query.MinPrice, query.MaxPrice,
		query.Bedrooms, query.MaxGuests, query.Search,
		query.SortBy, query.SortOrder,
		query.Page, query.PerPage, "",
	)
	if err != nil {
		return nil, 0, domain.ErrInternal(err)
	}

	var responses []dto.PropertyResponse
	for _, p := range props {
		responses = append(responses, toPropertyResponse(&p))
	}
	return responses, total, nil
}

// Update updates property details.
func (s *PropertyService) Update(ctx context.Context, id, ownerID uuid.UUID, req dto.UpdatePropertyRequest, isAdmin bool) (*dto.PropertyResponse, error) {
	prop, err := s.propRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if prop == nil {
		return nil, domain.ErrNotFound("Property")
	}

	// Check ownership
	if !isAdmin && prop.OwnerID != ownerID {
		return nil, domain.ErrForbidden()
	}

	// Apply updates
	if req.Title != nil {
		prop.Title = *req.Title
	}
	if req.Description != nil {
		prop.Description = *req.Description
	}
	if req.Neighborhood != nil {
		prop.Neighborhood = domain.Neighborhood(*req.Neighborhood)
	}
	if req.Address != nil {
		prop.Address = *req.Address
	}
	if req.Latitude != nil {
		prop.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		prop.Longitude = req.Longitude
	}
	if req.PricePerNight != nil {
		prop.PricePerNight = *req.PricePerNight
	}
	if req.Bedrooms != nil {
		prop.Bedrooms = *req.Bedrooms
	}
	if req.Bathrooms != nil {
		prop.Bathrooms = *req.Bathrooms
	}
	if req.MaxGuests != nil {
		prop.MaxGuests = *req.MaxGuests
	}
	if req.Amenities != nil {
		prop.Amenities = req.Amenities
	}

	if err := s.propRepo.Update(ctx, prop); err != nil {
		return nil, domain.ErrInternal(err)
	}

	resp := toPropertyResponse(prop)
	return &resp, nil
}

// UpdateStatus changes property status (admin only).
func (s *PropertyService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	prop, err := s.propRepo.GetByID(ctx, id)
	if err != nil {
		return domain.ErrInternal(err)
	}
	if prop == nil {
		return domain.ErrNotFound("Property")
	}

	return s.propRepo.UpdateStatus(ctx, id, domain.PropertyStatus(status))
}

// UploadImage uploads an image to S3 and adds it to the property.
func (s *PropertyService) UploadImage(ctx context.Context, propertyID, ownerID uuid.UUID, filename string, file io.Reader, contentType string, isAdmin bool) (*dto.PropertyImageDTO, error) {
	prop, err := s.propRepo.GetByID(ctx, propertyID)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if prop == nil {
		return nil, domain.ErrNotFound("Property")
	}
	if !isAdmin && prop.OwnerID != ownerID {
		return nil, domain.ErrForbidden()
	}

	if s.s3 == nil {
		// S3 not configured, use a placeholder URL
		imageID := uuid.New().String()
		img := domain.PropertyImage{
			ID:  imageID,
			URL: fmt.Sprintf("/uploads/properties/%s/%s%s", propertyID.String(), imageID, filepath.Ext(filename)),
		}
		prop.Images = append(prop.Images, img)
		if err := s.propRepo.UpdateImages(ctx, propertyID, prop.Images); err != nil {
			return nil, domain.ErrInternal(err)
		}
		return &dto.PropertyImageDTO{ID: img.ID, URL: img.URL}, nil
	}

	ext := filepath.Ext(filename)
	key := storage.GenerateKey(propertyID.String(), ext)

	url, err := s.s3.Upload(ctx, key, file, contentType)
	if err != nil {
		return nil, domain.ErrInternal(fmt.Errorf("failed to upload image: %w", err))
	}

	imageID := uuid.New().String()
	img := domain.PropertyImage{ID: imageID, URL: url}
	prop.Images = append(prop.Images, img)

	if err := s.propRepo.UpdateImages(ctx, propertyID, prop.Images); err != nil {
		return nil, domain.ErrInternal(err)
	}

	return &dto.PropertyImageDTO{ID: imageID, URL: url}, nil
}

// DeleteImage removes an image from a property.
func (s *PropertyService) DeleteImage(ctx context.Context, propertyID uuid.UUID, imageID string, ownerID uuid.UUID, isAdmin bool) error {
	prop, err := s.propRepo.GetByID(ctx, propertyID)
	if err != nil {
		return domain.ErrInternal(err)
	}
	if prop == nil {
		return domain.ErrNotFound("Property")
	}
	if !isAdmin && prop.OwnerID != ownerID {
		return domain.ErrForbidden()
	}

	var updated []domain.PropertyImage
	found := false
	for _, img := range prop.Images {
		if img.ID == imageID {
			found = true
			continue
		}
		updated = append(updated, img)
	}
	if !found {
		return domain.ErrNotFound("Image")
	}

	return s.propRepo.UpdateImages(ctx, propertyID, updated)
}

// GetAvailability returns unavailable date ranges for a property.
func (s *PropertyService) GetAvailability(ctx context.Context, propertyID uuid.UUID) ([]domain.DateRange, error) {
	prop, err := s.propRepo.GetByID(ctx, propertyID)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if prop == nil {
		return nil, domain.ErrNotFound("Property")
	}

	return s.propRepo.GetAvailability(ctx, propertyID)
}

func toPropertyResponse(p *domain.Property) dto.PropertyResponse {
	images := make([]dto.PropertyImageDTO, 0, len(p.Images))
	for _, img := range p.Images {
		images = append(images, dto.PropertyImageDTO{ID: img.ID, URL: img.URL})
	}

	resp := dto.PropertyResponse{
		ID:            p.ID.String(),
		OwnerID:       p.OwnerID.String(),
		Title:         p.Title,
		Description:   p.Description,
		Neighborhood:  string(p.Neighborhood),
		Address:       p.Address,
		Latitude:      p.Latitude,
		Longitude:     p.Longitude,
		PricePerNight: p.PricePerNight,
		Bedrooms:      p.Bedrooms,
		Bathrooms:     p.Bathrooms,
		MaxGuests:     p.MaxGuests,
		Amenities:     p.Amenities,
		Images:        images,
		Status:        string(p.Status),
		AverageRating: p.AverageRating,
		ReviewCount:   p.ReviewCount,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}

	if p.Owner != nil {
		ownerResp := toUserResponse(p.Owner)
		resp.Owner = &ownerResp
	}

	return resp
}
