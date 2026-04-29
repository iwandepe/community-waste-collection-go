CREATE TYPE pickup_type AS ENUM ('organic', 'plastic', 'paper', 'electronic');
CREATE TYPE pickup_status AS ENUM ('pending', 'scheduled', 'completed', 'canceled');

CREATE TABLE waste_pickups (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    type         pickup_type NOT NULL,
    status       pickup_status NOT NULL DEFAULT 'pending',
    pickup_date  TIMESTAMPTZ,
    safety_check BOOLEAN,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_waste_pickups_household_id ON waste_pickups(household_id);
CREATE INDEX idx_waste_pickups_status ON waste_pickups(status);
CREATE INDEX idx_waste_pickups_type_status ON waste_pickups(type, status);
