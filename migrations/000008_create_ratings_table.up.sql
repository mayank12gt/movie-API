CREATE TABLE IF NOT EXISTS ratings(
user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE, 
movie_id bigint NOT NULL REFERENCES movies ON DELETE CASCADE,
rating FLOAT,
created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
version integer NOT NULL DEFAULT 1,
PRIMARY KEY (user_id,movie_id)
);