package dto

// --- Admin DTOs ---

// AdminUserListQuery contains query parameters for admin user listing.
type AdminUserListQuery struct {
	Page    int    `form:"page,default=1" validate:"gte=1"`
	PerPage int    `form:"per_page,default=20" validate:"gte=1,lte=100"`
	Role    string `form:"role" validate:"omitempty,oneof=guest homeowner admin"`
	Search  string `form:"search"`
}

// UpdateUserRoleRequest is the request body for changing a user's role.
type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=guest homeowner admin"`
}

// UpdateUserStatusRequest is the request body for changing a user's active status.
type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// AdminBookingListQuery contains query parameters for admin booking listing.
type AdminBookingListQuery struct {
	Page    int    `form:"page,default=1" validate:"gte=1"`
	PerPage int    `form:"per_page,default=20" validate:"gte=1,lte=100"`
	Status  string `form:"status" validate:"omitempty,oneof=pending confirmed cancelled completed"`
}

// AdminPropertyListQuery contains query parameters for admin property listing.
type AdminPropertyListQuery struct {
	Page    int    `form:"page,default=1" validate:"gte=1"`
	PerPage int    `form:"per_page,default=20" validate:"gte=1,lte=100"`
	Status  string `form:"status" validate:"omitempty,oneof=draft published archived"`
}

// TicketListQuery contains query parameters for listing support tickets.
type TicketListQuery struct {
	Page     int    `form:"page,default=1" validate:"gte=1"`
	PerPage  int    `form:"per_page,default=20" validate:"gte=1,lte=100"`
	Status   string `form:"status" validate:"omitempty,oneof=open in_progress resolved closed"`
	Priority string `form:"priority" validate:"omitempty,oneof=low medium high urgent"`
}

// TicketReplyRequest is the request body for replying to a ticket.
type TicketReplyRequest struct {
	Content    string `json:"content" validate:"required,min=1"`
	IsInternal bool   `json:"is_internal"`
}

// UpdateTicketStatusRequest is the request for changing ticket status.
type UpdateTicketStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=open in_progress resolved closed"`
}

// TicketResponse is the public representation of a support ticket.
type TicketResponse struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Subject   string                 `json:"subject"`
	Status    string                 `json:"status"`
	Priority  string                 `json:"priority"`
	User      *UserResponse          `json:"user,omitempty"`
	Messages  []TicketMessageResponse `json:"messages,omitempty"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// TicketMessageResponse is the public representation of a ticket message.
type TicketMessageResponse struct {
	ID         string        `json:"id"`
	TicketID   string        `json:"ticket_id"`
	SenderID   string        `json:"sender_id"`
	Content    string        `json:"content"`
	IsInternal bool          `json:"is_internal"`
	Sender     *UserResponse `json:"sender,omitempty"`
	CreatedAt  string        `json:"created_at"`
}

// UpdatePermissionsRequest is the request for updating the permission matrix.
type UpdatePermissionsRequest struct {
	Permissions []PermissionEntry `json:"permissions" validate:"required,dive"`
}

// PermissionEntry is a single permission entry.
type PermissionEntry struct {
	Role       string `json:"role" validate:"required,oneof=guest homeowner admin"`
	Module     string `json:"module" validate:"required"`
	CanRead    bool   `json:"can_read"`
	CanWrite   bool   `json:"can_write"`
	CanDelete  bool   `json:"can_delete"`
	CanApprove bool   `json:"can_approve"`
}

// PermissionResponse is the response for the permission matrix.
type PermissionResponse struct {
	ID         string `json:"id"`
	Role       string `json:"role"`
	Module     string `json:"module"`
	CanRead    bool   `json:"can_read"`
	CanWrite   bool   `json:"can_write"`
	CanDelete  bool   `json:"can_delete"`
	CanApprove bool   `json:"can_approve"`
}

// UpdateSiteConfigRequest is the request for updating a site config variable.
type UpdateSiteConfigRequest struct {
	Value       string  `json:"value" validate:"required"`
	Description *string `json:"description"`
}

// SiteConfigResponse is the public representation of a site config variable.
type SiteConfigResponse struct {
	ID          string  `json:"id"`
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	Description *string `json:"description,omitempty"`
	UpdatedAt   string  `json:"updated_at"`
}

// AuditLogQuery contains query parameters for querying audit logs.
type AuditLogQuery struct {
	Page         int    `form:"page,default=1" validate:"gte=1"`
	PerPage      int    `form:"per_page,default=20" validate:"gte=1,lte=100"`
	UserID       string `form:"user_id" validate:"omitempty,uuid"`
	Action       string `form:"action"`
	ResourceType string `form:"resource_type"`
}

// AuditLogResponse is the public representation of an audit log entry.
type AuditLogResponse struct {
	ID           string  `json:"id"`
	UserID       *string `json:"user_id,omitempty"`
	Action       string  `json:"action"`
	ResourceType string  `json:"resource_type"`
	ResourceID   *string `json:"resource_id,omitempty"`
	Metadata     any     `json:"metadata,omitempty"`
	IPAddress    *string `json:"ip_address,omitempty"`
	CreatedAt    string  `json:"created_at"`
}
