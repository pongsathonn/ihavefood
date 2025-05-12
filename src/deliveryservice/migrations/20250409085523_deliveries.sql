-- Add migration script here

CREATE TABLE IF NOT EXISTS riders (
    id           INTEGER NOT NULL,
    name         TEXT UNIQUE NOT NULL,
    phone_number TEXT UNIQUE NOT NULL,
    PRIMARY KEY(id)
);

-- order_id is generated from OrderService (Mongo ObjectID 12bytes)
CREATE TABLE IF NOT EXISTS deliveries(
    id             INTEGER NOT NULL,
    order_id       TEXT UNIQUE NOT NULL,
    rider_id       INTEGER,
    pickup_code    TEXT NOT NULL,
    pickup_lat     REAL NOT NULL,
    pickup_lng     REAL NOT NULL,
    drop_off_lat   REAL NOT NULL,
    drop_off_lng   REAL NOT NULL,
    status         INTEGER NOT NULL
                   DEFAULT 0
                   CHECK(status IN (0,1,2)),
    create_time    DATETIME NOT NULL,
    accept_time    DATETIME,
    deliver_time   DATETIME,

    PRIMARY KEY(id)
    FOREIGN KEY (rider_id)
        REFERENCES riders(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

