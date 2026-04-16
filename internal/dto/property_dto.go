package dto

// --- Property DTOs ---

// CreatePropertyRequest is the request body for creating a property.
type CreatePropertyRequest struct {
	Title         string   `json:"title" validate:"required,min=1,max=255"`
	Description   string   `json:"description" validate:"required,min=10"`
	Neighborhood  string   `json:"neighborhood" validate:"required,oneof=akwa bonapriso bonanjo kotto other"`
	Address       string   `json:"address" validate:"required"`
	Latitude      *float64 `json:"latitude" validate:"omitempty,latitude"`
	Longitude     *float64 `json:"longitude" validate:"omitempty,longitude"`
	PricePerNight float64  `json:"price_per_night" validate:"required,gt=0"`
	Bedrooms      int      `json:"bedrooms" validate:"required,gte=1"`
	Bathrooms     int      `json:"bathrooms" validate:"required,gte=1"`
	MaxGuests     int      `json:"max_guests" validate:"required,gte=1"`
	Amenities     []string `json:"amenities"`
}

// UpdatePropertyRequest is the request body for updating a property.
type UpdatePropertyRequest struct {
	Title         *string  `json:"title" validate:"omitempty,min=1,max=255"`
	Description   *string  `json:"description" validate:"omitempty,min=10"`
	Neighborhood  *string  `json:"neighborhood" validate:"omitempty,oneof=akwa bonapriso bonanjo kotto other"`
	Address       *string  `json:"address"`
	Latitude      *float64 `json:"latitude" validate:"omitempty,latitude"`
	Longitude     *float64 `json:"longitude" validate:"omitempty,longitude"`
	PricePerNight *float64 `json:"price_per_night" validate:"omitempty,gt=0"`
	Bedrooms      *int     `json:"bedrooms" validate:"omitempty,gte=1"`
	Bathrooms     *int     `json:"bathrooms" validate:"omitempty,gte=1"`
	MaxGuests     *int     `json:"max_guests" validate:"omitempty,gte=1"`
	Amenities     []string `json:"amenities"`
}

// UpdatePropertyStatusRequest is the request for changing property status.
type UpdatePropertyStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft published archived"`
}

// PropertyListQuery contains query parameters for listing properties.
type PropertyListQuery struct {
	Page         int     `form:"page,default=1" validate:"gte=1"`
	PerPage      int     `form:"per_page,default=20" validate:"gte=1,lte=100"`
	Neighborhood string  `form:"neighborhood" validate:"omitempty,oneof=akwa bonapriso bonanjo kotto other"`
	MinPrice     float64 `form:"min_price" validate:"omitempty,gte=0"`
	MaxPrice     float64 `form:"max_price" validate:"omitempty,gte=0"`
	Bedrooms     int     `form:"bedrooms" validate:"omitempty,gte=1"`
	MaxGuests    int     `form:"max_guests" validate:"omitempty,gte=1"`
	SortBy       string  `form:"sort_by,default=created_at" validate:"omitempty,oneof=created_at price_per_night"`
	SortOrder    string  `form:"sort_order,default=desc" validate:"omitempty,oneof=asc desc"`
	Search       string  `form:"search"`
}

// PropertyResponse is the public representation of a property.
type PropertyResponse struct {
	ID            string            `json:"id"`
	OwnerID       string            `json:"owner_id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Neighborhood  string            `json:"neighborhood"`
	Address       string            `json:"address"`
	Latitude      *float64          `json:"latitude,omitempty"`
	Longitude     *float64          `json:"longitude,omitempty"`
	PricePerNight float64           `json:"price_per_night"`
	Bedrooms      int               `json:"bedrooms"`
	Bathrooms     int               `json:"bathrooms"`
	MaxGuests     int               `json:"max_guests"`
	Amenities     []string          `json:"amenities"`
	Images        []PropertyImageDTO `json:"images"`
	Status        string            `json:"status"`
	AverageRating *float64          `json:"average_rating,omitempty"`
	ReviewCount   *int              `json:"review_count,omitempty"`
	Owner         *UserResponse     `json:"owner,omitempty"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

// PropertyImageDTO is the public representation of a property image.
type PropertyImageDTO struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
