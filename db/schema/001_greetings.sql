CREATE TABLE IF NOT EXISTS greetings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    message TEXT NOT NULL
);

INSERT OR IGNORE INTO greetings (id, message) VALUES (1, 'Hello, World!');
