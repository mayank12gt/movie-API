-- Create a temporary table to stage the data
CREATE TEMP TABLE temp_table (
title text NOT NULL,
year integer NOT NULL,
runtime integer NOT NULL,
genres text[] NOT NULL
);

-- Copy data from CSV to the temporary table
--COPY temp_table (title, year, runtime, genres)
--FROM '/tmp/movie_seed.csv' DELIMITER ',' CSV HEADER;

COPY temp_table (title, year, runtime, genres)
FROM 'C:\Temp\movie_seed.csv' DELIMITER ',' CSV HEADER;





-- Insert data from temporary table into the main table
INSERT INTO movies (title, year, runtime, genres)
SELECT title, year, runtime, genres FROM temp_table LIMIT 100;
