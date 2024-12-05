
CREATE TABLE credentials (
    user_id INT GENERATED ALWAYS AS IDENTITY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role SMALLINT NOT NULL,
    phone_number VARCHAR(15) UNIQUE,
    create_time TIMESTAMP NOT NULL DEFAULT NOW(),
    update_time TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id)
);

