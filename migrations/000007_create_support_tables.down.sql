DROP TABLE IF EXISTS ticket_messages;
DROP TRIGGER IF EXISTS update_support_tickets_updated_at ON support_tickets;
DROP TABLE IF EXISTS support_tickets;
DROP TYPE IF EXISTS ticket_priority;
DROP TYPE IF EXISTS ticket_status;
