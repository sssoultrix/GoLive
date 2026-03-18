CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS profiles (
    user_id uuid PRIMARY KEY,
    login text NOT NULL UNIQUE,
    display_name text NOT NULL DEFAULT '',
    avatar_url text NOT NULL DEFAULT '',
    bio text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS profiles_login_idx ON profiles(login);

