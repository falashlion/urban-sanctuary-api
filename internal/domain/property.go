package domain

import (
	"time"

	"github.com/google/uuid"
)

// Neighborhood represents a neighborhood in Douala.
type Neighborhood string

const (
	NeighborhoodAkwa      Neighborhood = "akwa"
	NeighborhoodBonapriso Neighborhood = "bonapriso"
	NeighborhoodBonanjo   Neighborhood = "bonanjo"
	NeighborhoodKotto     Neighborhood = "kotto"
	NeighborhoodOther     Neighborhood = "other"
)

// PropertyStatus represents the publication status of a property.
type PropertyStatus string

const (
	PropertyStatusDraft     PropertyStatus = "draft"
	PropertyStatusPublished PropertyStatus = "published"
	PropertyStatusArchived  PropertyStatus = "archived"
)

// Property represents an apartment listing.
type Property struct {
	ID            uuid.UUID      `json:"id"`
	OwnerID       uuid.UUID      `json:"owner_id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Neighborhood  Neighborhood   `json:"neighborhood"`
	Address       string         `json:"address"`
	Latitude      *float64       `json:"latitude,omitempty"`
	Longitude     *float64       `json:"longitude,omitempty"`
	PricePerNight float64        `json:"price_per_night"`
	Bedrooms      int            `json:"bedrooms"`
	Bathrooms     int            `json:"bathrooms"`
	MaxGuests     int            `json:"max_guests"`
	Amenities     []string       `json:"amenities"`
	Images        []PropertyImage `json:"images"`
	Status        PropertyStatus `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	// Joined data (not always populated)
	Owner         *User          `json:"owner,omitempty"`
	AverageRating *float64       `json:"average_rating,omitempty"`
	ReviewCount   *int           `json:"review_count,omitempty"`
}

// PropertyImage represents an image associated with a property.
type PropertyImage struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// DateRange represents a range of unavailable dates.
type DateRange struct {
	CheckIn  time.Time `json:"check_in"`
	CheckOut time.Time `json:"check_out"`
}
