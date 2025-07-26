-- Создаём таблицу сегментов
CREATE TABLE IF NOT EXISTS segments (
                                        slug VARCHAR(50) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW()
    );

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     created_at TIMESTAMP DEFAULT NOW()
    );

-- Связь пользователей с сегментами (many-to-many)
CREATE TABLE IF NOT EXISTS user_segments (
                                             user_id INT REFERENCES users(id) ON DELETE CASCADE,
    segment_slug VARCHAR(50) REFERENCES segments(slug) ON DELETE CASCADE,
    PRIMARY KEY (user_id, segment_slug)
    );