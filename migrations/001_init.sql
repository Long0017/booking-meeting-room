-- Enable btree_gist for exclusion constraint (prevents double-booking)
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS rooms (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(100)  NOT NULL,
    capacity   INT           NOT NULL DEFAULT 20,
    location   VARCHAR(200),
    is_active  BOOLEAN       NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bookings (
    id              SERIAL PRIMARY KEY,
    room_id         INT          NOT NULL REFERENCES rooms(id),
    employee_id     VARCHAR(20)  NOT NULL,
    department      VARCHAR(100) NOT NULL,
    title           VARCHAR(200) NOT NULL,
    start_time      TIMESTAMPTZ  NOT NULL,
    end_time        TIMESTAMPTZ  NOT NULL,
    attendees_count INT          NOT NULL DEFAULT 1,
    notes           TEXT,
    status          VARCHAR(20)  NOT NULL DEFAULT 'confirmed',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT bookings_time_check CHECK (end_time > start_time),
    EXCLUDE USING gist (
        room_id WITH =,
        tstzrange(start_time, end_time, '[)') WITH &&
    ) WHERE (status = 'confirmed')
);

INSERT INTO rooms (name, capacity, location)
VALUES ('ห้องประชุมใหญ่', 20, 'ชั้น 3')
ON CONFLICT DO NOTHING;
