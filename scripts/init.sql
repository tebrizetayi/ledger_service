CREATE TABLE IF NOT EXISTS  users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS  transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);





