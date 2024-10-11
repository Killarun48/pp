CREATE TABLE IF NOT EXISTS users
(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMP
);

INSERT INTO users(name, email) VALUES
('test','test'),
('test2','test2');