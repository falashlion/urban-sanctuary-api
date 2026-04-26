-- Neighborhood enum
DO $$ BEGIN
    CREATE TYPE neighborhood AS ENUM ('akwa', 'bonapriso', 'bonanjo', 'kotto', 'other');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Property status enum
DO $$ BEGIN
    CREATE TYPE property_status AS ENUM ('draft', 'published', 'archived');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Properties table
CREATE TABLE IF NOT EXISTS properties (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL,
    neighborhood    neighborhood NOT NULL,
    address         TEXT NOT NULL,
    latitude        DECIMAL(10,8),
    longitude       DECIMAL(11,8),
    price_per_night DECIMAL(12,2) NOT NULL,
    bedrooms        SMALLINT NOT NULL DEFAULT 1,
    bathrooms       SMALLINT NOT NULL DEFAULT 1,
    max_guests      SMALLINT NOT NULL DEFAULT 2,
    amenities       JSONB NOT NULL DEFAULT '[]',
    images          JSONB NOT NULL DEFAULT '[]',
    status          property_status NOT NULL DEFAULT 'draft',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Price must be positive
    CONSTRAINT properties_price_positive CHECK (price_per_night > 0),
    CONSTRAINT properties_bedrooms_positive CHECK (bedrooms > 0),
    CONSTRAINT properties_bathrooms_positive CHECK (bathrooms > 0),
    CONSTRAINT properties_max_guests_positive CHECK (max_guests > 0)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_properties_owner_id ON properties(owner_id);
CREATE INDEX IF NOT EXISTS idx_properties_status ON properties(status);
CREATE INDEX IF NOT EXISTS idx_properties_neighborhood ON properties(neighborhood);
CREATE INDEX IF NOT EXISTS idx_properties_price ON properties(price_per_night);
CREATE INDEX IF NOT EXISTS idx_properties_created_at ON properties(created_at DESC);

DROP TRIGGER IF EXISTS update_properties_updated_at ON properties;
CREATE TRIGGER update_properties_updated_at
    BEFORE UPDATE ON properties
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
