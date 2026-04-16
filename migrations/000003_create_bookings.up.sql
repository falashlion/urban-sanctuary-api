-- Booking status enum
DO $$ BEGIN
    CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'cancelled', 'completed');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Bookings table
CREATE TABLE IF NOT EXISTS bookings (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id          UUID NOT NULL REFERENCES properties(id) ON DELETE RESTRICT,
    guest_id             UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    check_in             DATE NOT NULL,
    check_out            DATE NOT NULL,
    nights               SMALLINT GENERATED ALWAYS AS (check_out - check_in) STORED,
    guests_count         SMALLINT NOT NULL DEFAULT 1,
    base_amount          DECIMAL(12,2) NOT NULL,
    discount_amount      DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_amount         DECIMAL(12,2) NOT NULL,
    status               booking_status NOT NULL DEFAULT 'pending',
    loyalty_points_earned INTEGER NOT NULL DEFAULT 0,
    loyalty_points_used  INTEGER NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT bookings_checkout_after_checkin CHECK (check_out > check_in),
    CONSTRAINT bookings_amount_positive CHECK (base_amount > 0),
    CONSTRAINT bookings_total_positive CHECK (total_amount >= 0),
    CONSTRAINT bookings_guests_positive CHECK (guests_count > 0)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_bookings_property_id ON bookings(property_id);
CREATE INDEX IF NOT EXISTS idx_bookings_guest_id ON bookings(guest_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
CREATE INDEX IF NOT EXISTS idx_bookings_dates ON bookings(property_id, check_in, check_out);
CREATE INDEX IF NOT EXISTS idx_bookings_created_at ON bookings(created_at DESC);

-- Prevent overlapping confirmed/pending bookings for the same property
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_no_overlap
    ON bookings (property_id, check_in, check_out)
    WHERE status IN ('pending', 'confirmed');

CREATE TRIGGER update_bookings_updated_at
    BEFORE UPDATE ON bookings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
