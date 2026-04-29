CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed');

CREATE TABLE payments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id   UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    waste_id       UUID NOT NULL REFERENCES waste_pickups(id) ON DELETE CASCADE,
    amount         NUMERIC(10, 2) NOT NULL,
    payment_date   TIMESTAMPTZ,
    status         payment_status NOT NULL DEFAULT 'pending',
    proof_file_url TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_household_id ON payments(household_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_payment_date ON payments(payment_date);
