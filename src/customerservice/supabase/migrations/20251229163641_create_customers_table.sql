CREATE TABLE customers (
    customer_id UUID,
    username VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    email VARCHAR(255) UNIQUE NOT NULL,
    facebook VARCHAR(255),
    instagram VARCHAR(255),
    line VARCHAR(255),
    create_time TIMESTAMP NOT NULL DEFAULT NOW(),
    update_time TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (customer_id)
);
