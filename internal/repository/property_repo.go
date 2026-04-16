package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PropertyRepository handles database operations for properties.
type PropertyRepository struct {
	pool *pgxpool.Pool
}

// NewPropertyRepository creates a new PropertyRepository.
func NewPropertyRepository(pool *pgxpool.Pool) *PropertyRepository {
	return &PropertyRepository{pool: pool}
}

// Create inserts a new property into the database.
func (r *PropertyRepository) Create(ctx context.Context, p *domain.Property) error {
	amenitiesJSON, _ := json.Marshal(p.Amenities)
	imagesJSON, _ := json.Marshal(p.Images)

	query := `
		INSERT INTO properties (id, owner_id, title, description, neighborhood, address,
			latitude, longitude, price_per_night, bedrooms, bathrooms, max_guests,
			amenities, images, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING created_at, updated_at`

	p.ID = uuid.New()
	return r.pool.QueryRow(ctx, query,
		p.ID, p.OwnerID, p.Title, p.Description, p.Neighborhood, p.Address,
		p.Latitude, p.Longitude, p.PricePerNight, p.Bedrooms, p.Bathrooms, p.MaxGuests,
		amenitiesJSON, imagesJSON, p.Status,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
}

// GetByID retrieves a property by its ID.
func (r *PropertyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Property, error) {
	query := `
		SELECT p.id, p.owner_id, p.title, p.description, p.neighborhood, p.address,
			p.latitude, p.longitude, p.price_per_night, p.bedrooms, p.bathrooms, p.max_guests,
			p.amenities, p.images, p.status, p.created_at, p.updated_at,
			COALESCE(AVG(rv.rating), 0) as avg_rating,
			COUNT(rv.id) as review_count
		FROM properties p
		LEFT JOIN reviews rv ON rv.property_id = p.id
		WHERE p.id = $1
		GROUP BY p.id`

	p := &domain.Property{}
	var amenitiesJSON, imagesJSON []byte
	var avgRating float64
	var reviewCount int

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.OwnerID, &p.Title, &p.Description, &p.Neighborhood, &p.Address,
		&p.Latitude, &p.Longitude, &p.PricePerNight, &p.Bedrooms, &p.Bathrooms, &p.MaxGuests,
		&amenitiesJSON, &imagesJSON, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		&avgRating, &reviewCount,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}

	_ = json.Unmarshal(amenitiesJSON, &p.Amenities)
	_ = json.Unmarshal(imagesJSON, &p.Images)
	if p.Amenities == nil {
		p.Amenities = []string{}
	}
	if p.Images == nil {
		p.Images = []domain.PropertyImage{}
	}
	p.AverageRating = &avgRating
	p.ReviewCount = &reviewCount

	return p, nil
}

// Update updates a property's details.
func (r *PropertyRepository) Update(ctx context.Context, p *domain.Property) error {
	amenitiesJSON, _ := json.Marshal(p.Amenities)

	query := `
		UPDATE properties SET title = $2, description = $3, neighborhood = $4, address = $5,
			latitude = $6, longitude = $7, price_per_night = $8, bedrooms = $9, bathrooms = $10,
			max_guests = $11, amenities = $12, updated_at = NOW()
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query,
		p.ID, p.Title, p.Description, p.Neighborhood, p.Address,
		p.Latitude, p.Longitude, p.PricePerNight, p.Bedrooms, p.Bathrooms,
		p.MaxGuests, amenitiesJSON,
	)
	return err
}

// UpdateStatus updates a property's status.
func (r *PropertyRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PropertyStatus) error {
	query := `UPDATE properties SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, status)
	return err
}

// UpdateImages updates a property's images.
func (r *PropertyRepository) UpdateImages(ctx context.Context, id uuid.UUID, images []domain.PropertyImage) error {
	imagesJSON, _ := json.Marshal(images)
	query := `UPDATE properties SET images = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, imagesJSON)
	return err
}

// List retrieves properties with filters, sorting, and pagination.
func (r *PropertyRepository) List(ctx context.Context, neighborhood string, minPrice, maxPrice float64,
	bedrooms, maxGuests int, search, sortBy, sortOrder string, page, perPage int, statusFilter string) ([]domain.Property, int64, error) {

	countQuery := `SELECT COUNT(*) FROM properties WHERE 1=1`
	dataQuery := `
		SELECT p.id, p.owner_id, p.title, p.description, p.neighborhood, p.address,
			p.latitude, p.longitude, p.price_per_night, p.bedrooms, p.bathrooms, p.max_guests,
			p.amenities, p.images, p.status, p.created_at, p.updated_at,
			COALESCE(AVG(rv.rating), 0) as avg_rating,
			COUNT(rv.id) as review_count
		FROM properties p
		LEFT JOIN reviews rv ON rv.property_id = p.id
		WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	statusCond := " AND p.status = 'published'"
	if statusFilter != "" {
		statusCond = fmt.Sprintf(" AND p.status = $%d", argIdx)
		args = append(args, statusFilter)
		argIdx++
	}
	countStatusCond := statusCond
	if statusFilter == "" {
		countStatusCond = " AND status = 'published'"
	}
	countQuery += countStatusCond
	dataQuery += statusCond

	if neighborhood != "" {
		f := fmt.Sprintf(" AND p.neighborhood = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND neighborhood = $%d", argIdx)
		dataQuery += f
		args = append(args, neighborhood)
		argIdx++
	}
	if minPrice > 0 {
		f := fmt.Sprintf(" AND p.price_per_night >= $%d", argIdx)
		countQuery += fmt.Sprintf(" AND price_per_night >= $%d", argIdx)
		dataQuery += f
		args = append(args, minPrice)
		argIdx++
	}
	if maxPrice > 0 {
		f := fmt.Sprintf(" AND p.price_per_night <= $%d", argIdx)
		countQuery += fmt.Sprintf(" AND price_per_night <= $%d", argIdx)
		dataQuery += f
		args = append(args, maxPrice)
		argIdx++
	}
	if bedrooms > 0 {
		f := fmt.Sprintf(" AND p.bedrooms >= $%d", argIdx)
		countQuery += fmt.Sprintf(" AND bedrooms >= $%d", argIdx)
		dataQuery += f
		args = append(args, bedrooms)
		argIdx++
	}
	if maxGuests > 0 {
		f := fmt.Sprintf(" AND p.max_guests >= $%d", argIdx)
		countQuery += fmt.Sprintf(" AND max_guests >= $%d", argIdx)
		dataQuery += f
		args = append(args, maxGuests)
		argIdx++
	}
	if search != "" {
		f := fmt.Sprintf(" AND (p.title ILIKE $%d OR p.description ILIKE $%d)", argIdx, argIdx)
		countQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx)
		dataQuery += f
		args = append(args, "%"+search+"%")
		argIdx++
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count properties: %w", err)
	}

	dataQuery += " GROUP BY p.id"

	// Sort
	validSorts := map[string]string{
		"created_at":     "p.created_at",
		"price_per_night": "p.price_per_night",
	}
	sortCol, ok := validSorts[sortBy]
	if !ok {
		sortCol = "p.created_at"
	}
	if sortOrder != "asc" {
		sortOrder = "desc"
	}
	dataQuery += fmt.Sprintf(" ORDER BY %s %s", sortCol, sortOrder)
	dataQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list properties: %w", err)
	}
	defer rows.Close()

	var properties []domain.Property
	for rows.Next() {
		var p domain.Property
		var amenitiesJSON, imagesJSON []byte
		var avgRating float64
		var reviewCount int

		if err := rows.Scan(
			&p.ID, &p.OwnerID, &p.Title, &p.Description, &p.Neighborhood, &p.Address,
			&p.Latitude, &p.Longitude, &p.PricePerNight, &p.Bedrooms, &p.Bathrooms, &p.MaxGuests,
			&amenitiesJSON, &imagesJSON, &p.Status, &p.CreatedAt, &p.UpdatedAt,
			&avgRating, &reviewCount,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan property: %w", err)
		}

		_ = json.Unmarshal(amenitiesJSON, &p.Amenities)
		_ = json.Unmarshal(imagesJSON, &p.Images)
		if p.Amenities == nil {
			p.Amenities = []string{}
		}
		if p.Images == nil {
			p.Images = []domain.PropertyImage{}
		}
		p.AverageRating = &avgRating
		p.ReviewCount = &reviewCount
		properties = append(properties, p)
	}

	return properties, total, nil
}

// GetAvailability gets unavailable date ranges for a property.
func (r *PropertyRepository) GetAvailability(ctx context.Context, propertyID uuid.UUID) ([]domain.DateRange, error) {
	query := `
		SELECT check_in, check_out FROM bookings
		WHERE property_id = $1 AND status IN ('pending', 'confirmed')
		AND check_out > CURRENT_DATE
		ORDER BY check_in`

	rows, err := r.pool.Query(ctx, query, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get availability: %w", err)
	}
	defer rows.Close()

	var ranges []domain.DateRange
	for rows.Next() {
		var dr domain.DateRange
		if err := rows.Scan(&dr.CheckIn, &dr.CheckOut); err != nil {
			return nil, fmt.Errorf("failed to scan date range: %w", err)
		}
		ranges = append(ranges, dr)
	}
	return ranges, nil
}
