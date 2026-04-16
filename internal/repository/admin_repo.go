package repository

import (
	"context"
	"fmt"

	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminRepository handles database operations for admin-specific tables.
type AdminRepository struct {
	pool *pgxpool.Pool
}

// NewAdminRepository creates a new AdminRepository.
func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

// --- Support Tickets ---

// ListTickets retrieves support tickets with filters and pagination.
func (r *AdminRepository) ListTickets(ctx context.Context, status, priority string, page, perPage int) ([]domain.SupportTicket, int64, error) {
	countQuery := `SELECT COUNT(*) FROM support_tickets WHERE 1=1`
	dataQuery := `
		SELECT t.id, t.user_id, t.subject, t.status, t.priority, t.created_at, t.updated_at,
			u.first_name, u.last_name, u.email
		FROM support_tickets t
		JOIN users u ON u.id = t.user_id
		WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if status != "" {
		f := fmt.Sprintf(" AND t.status = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		dataQuery += f
		args = append(args, status)
		argIdx++
	}
	if priority != "" {
		f := fmt.Sprintf(" AND t.priority = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND priority = $%d", argIdx)
		dataQuery += f
		args = append(args, priority)
		argIdx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery += fmt.Sprintf(" ORDER BY t.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []domain.SupportTicket
	for rows.Next() {
		var t domain.SupportTicket
		var firstName, lastName string
		var email *string
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Subject, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt,
			&firstName, &lastName, &email,
		); err != nil {
			return nil, 0, err
		}
		t.User = &domain.User{ID: t.UserID, FirstName: firstName, LastName: lastName, Email: email}
		tickets = append(tickets, t)
	}
	return tickets, total, nil
}

// GetTicketByID retrieves a ticket with its messages.
func (r *AdminRepository) GetTicketByID(ctx context.Context, id uuid.UUID) (*domain.SupportTicket, error) {
	query := `
		SELECT t.id, t.user_id, t.subject, t.status, t.priority, t.created_at, t.updated_at
		FROM support_tickets t WHERE t.id = $1`

	t := &domain.SupportTicket{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.UserID, &t.Subject, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load messages
	msgQuery := `
		SELECT m.id, m.ticket_id, m.sender_id, m.content, m.is_internal, m.created_at,
			u.first_name, u.last_name
		FROM ticket_messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.ticket_id = $1 ORDER BY m.created_at ASC`

	rows, err := r.pool.Query(ctx, msgQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m domain.TicketMessage
		var senderFirst, senderLast string
		if err := rows.Scan(
			&m.ID, &m.TicketID, &m.SenderID, &m.Content, &m.IsInternal, &m.CreatedAt,
			&senderFirst, &senderLast,
		); err != nil {
			return nil, err
		}
		m.Sender = &domain.User{ID: m.SenderID, FirstName: senderFirst, LastName: senderLast}
		t.Messages = append(t.Messages, m)
	}

	return t, nil
}

// CreateTicketMessage adds a message to a support ticket.
func (r *AdminRepository) CreateTicketMessage(ctx context.Context, msg *domain.TicketMessage) error {
	query := `
		INSERT INTO ticket_messages (id, ticket_id, sender_id, content, is_internal)
		VALUES ($1, $2, $3, $4, $5) RETURNING created_at`

	msg.ID = uuid.New()
	return r.pool.QueryRow(ctx, query,
		msg.ID, msg.TicketID, msg.SenderID, msg.Content, msg.IsInternal,
	).Scan(&msg.CreatedAt)
}

// UpdateTicketStatus updates a ticket's status.
func (r *AdminRepository) UpdateTicketStatus(ctx context.Context, id uuid.UUID, status domain.TicketStatus) error {
	query := `UPDATE support_tickets SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, status)
	return err
}

// --- Permissions ---

// ListPermissions retrieves all permission entries.
func (r *AdminRepository) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	query := `SELECT id, role, module, can_read, can_write, can_delete, can_approve, created_at, updated_at
		FROM permissions ORDER BY role, module`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(
			&p.ID, &p.Role, &p.Module, &p.CanRead, &p.CanWrite, &p.CanDelete, &p.CanApprove,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// GetPermission retrieves a specific permission by role and module.
func (r *AdminRepository) GetPermission(ctx context.Context, role domain.Role, module string) (*domain.Permission, error) {
	query := `SELECT id, role, module, can_read, can_write, can_delete, can_approve, created_at, updated_at
		FROM permissions WHERE role = $1 AND module = $2`

	p := &domain.Permission{}
	err := r.pool.QueryRow(ctx, query, role, module).Scan(
		&p.ID, &p.Role, &p.Module, &p.CanRead, &p.CanWrite, &p.CanDelete, &p.CanApprove,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return p, err
}

// UpsertPermission creates or updates a permission entry.
func (r *AdminRepository) UpsertPermission(ctx context.Context, p *domain.Permission) error {
	query := `
		INSERT INTO permissions (id, role, module, can_read, can_write, can_delete, can_approve)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (role, module)
		DO UPDATE SET can_read = $4, can_write = $5, can_delete = $6, can_approve = $7, updated_at = NOW()`

	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.Role, p.Module, p.CanRead, p.CanWrite, p.CanDelete, p.CanApprove,
	)
	return err
}

// --- Site Config ---

// ListSiteConfig retrieves all site configuration variables.
func (r *AdminRepository) ListSiteConfig(ctx context.Context) ([]domain.SiteConfig, error) {
	query := `SELECT id, key, value, description, updated_by, updated_at, created_at
		FROM site_config ORDER BY key`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.SiteConfig
	for rows.Next() {
		var c domain.SiteConfig
		if err := rows.Scan(
			&c.ID, &c.Key, &c.Value, &c.Description, &c.UpdatedBy, &c.UpdatedAt, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, nil
}

// GetSiteConfig retrieves a site config by key.
func (r *AdminRepository) GetSiteConfig(ctx context.Context, key string) (*domain.SiteConfig, error) {
	query := `SELECT id, key, value, description, updated_by, updated_at, created_at
		FROM site_config WHERE key = $1`

	c := &domain.SiteConfig{}
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&c.ID, &c.Key, &c.Value, &c.Description, &c.UpdatedBy, &c.UpdatedAt, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return c, err
}

// UpdateSiteConfig updates a site config variable.
func (r *AdminRepository) UpdateSiteConfig(ctx context.Context, key, value string, description *string, updatedBy uuid.UUID) error {
	query := `UPDATE site_config SET value = $2, description = COALESCE($3, description), updated_by = $4, updated_at = NOW() WHERE key = $1`
	result, err := r.pool.Exec(ctx, query, key, value, description, updatedBy)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("config key not found: %s", key)
	}
	return nil
}

// --- Audit Logs ---

// CreateAuditLog inserts an audit log entry.
func (r *AdminRepository) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, metadata, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	log.ID = uuid.New()
	return r.pool.QueryRow(ctx, query,
		log.ID, log.UserID, log.Action, log.ResourceType, log.ResourceID,
		log.Metadata, log.IPAddress, log.UserAgent,
	).Scan(&log.CreatedAt)
}

// ListAuditLogs retrieves audit logs with filters and pagination.
func (r *AdminRepository) ListAuditLogs(ctx context.Context, userID, action, resourceType string, page, perPage int) ([]domain.AuditLog, int64, error) {
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`
	dataQuery := `
		SELECT id, user_id, action, resource_type, resource_id, metadata, ip_address, user_agent, created_at
		FROM audit_logs WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if userID != "" {
		f := fmt.Sprintf(" AND user_id = $%d", argIdx)
		countQuery += f
		dataQuery += f
		args = append(args, userID)
		argIdx++
	}
	if action != "" {
		f := fmt.Sprintf(" AND action = $%d", argIdx)
		countQuery += f
		dataQuery += f
		args = append(args, action)
		argIdx++
	}
	if resourceType != "" {
		f := fmt.Sprintf(" AND resource_type = $%d", argIdx)
		countQuery += f
		dataQuery += f
		args = append(args, resourceType)
		argIdx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.Action, &l.ResourceType, &l.ResourceID,
			&l.Metadata, &l.IPAddress, &l.UserAgent, &l.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}

// --- Notifications ---

// ListNotifications retrieves notifications for a user.
func (r *AdminRepository) ListNotifications(ctx context.Context, userID uuid.UUID, page, perPage int) ([]domain.Notification, int64, error) {
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	dataQuery := `
		SELECT id, user_id, type, channel, title, content, sent_at, read_at, created_at
		FROM notifications WHERE user_id = $1 ORDER BY created_at DESC`

	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery += fmt.Sprintf(" LIMIT $2 OFFSET $3")
	rows, err := r.pool.Query(ctx, dataQuery, userID, perPage, (page-1)*perPage)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Channel, &n.Title, &n.Content,
			&n.SentAt, &n.ReadAt, &n.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, n)
	}
	return notifications, total, nil
}

// CreateNotification inserts a new notification.
func (r *AdminRepository) CreateNotification(ctx context.Context, n *domain.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, channel, title, content, sent_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`

	n.ID = uuid.New()
	return r.pool.QueryRow(ctx, query,
		n.ID, n.UserID, n.Type, n.Channel, n.Title, n.Content, n.SentAt,
	).Scan(&n.CreatedAt)
}
