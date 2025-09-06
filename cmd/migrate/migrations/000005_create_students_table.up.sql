CREATE TABLE students (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(72) NOT NULL,
    last_name VARCHAR(72) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password bytea NOT NULL,
    phone_number VARCHAR(20),
    class VARCHAR(10) NOT NULL,
    birth_date DATE NOT NULL,
    address TEXT NOT NULL,
    parent_name VARCHAR(255) NOT NULL,
    parent_phone_number VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

