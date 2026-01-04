CREATE TABLE addresses (
    address_id UUID DEFAULT gen_random_uuid(),
    customer_id UUID,
    address_name VARCHAR(255),                 
    sub_district VARCHAR(255),                 
    district VARCHAR(255),
    province VARCHAR(255),
    postal_code VARCHAR(20),
    PRIMARY KEY(address_id),
    CONSTRAINT fk_profile
        FOREIGN KEY(customer_id)
        REFERENCES customers(customer_id)
        ON DELETE CASCADE
);


