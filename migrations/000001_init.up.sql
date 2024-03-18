CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    hashed_password VARCHAR(255) NOT NULL
);

CREATE TABLE actors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) DEFAULT NULL,
    gender VARCHAR(10) DEFAULT NULL,
    date DATE DEFAULT NULL
);

CREATE TABLE films (
    id SERIAL PRIMARY KEY,
    name VARCHAR(1000) NOT NULL,
    description VARCHAR(1000),
    date INTEGER CHECK (date >= 1900 AND date <= 2200),
    rating INT CHECK (rating >= 0 AND rating <= 10) DEFAULT NULL
);


CREATE TABLE film_actor (
    film_id INTEGER,
    actor_id INTEGER,
    PRIMARY KEY (film_id, actor_id)
);

INSERT INTO users (username, role, hashed_password) VALUES
('admin', 'admin','$2a$10$GJJT0w7LaC4oOfTygVVi8Ofn8D5kZHTXWdhA8PzscqkK846A2oG.C');