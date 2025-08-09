CREATE TABLE IF NOT EXISTS riders (
    id           TEXT NOT NULL,
    username     TEXT UNIQUE NOT NULL,
    phone_number TEXT UNIQUE NOT NULL,
    PRIMARY KEY(id)
);

-- TODO: might remove deliveries.id and use order_id as PK
CREATE TABLE IF NOT EXISTS deliveries(
    id             TEXT NOT NULL,
    order_id       TEXT UNIQUE NOT NULL,
    rider_id       TEXT,
    pickup_code    TEXT NOT NULL,
    pickup_lat     REAL NOT NULL,
    pickup_lng     REAL NOT NULL,
    drop_off_lat   REAL NOT NULL,
    drop_off_lng   REAL NOT NULL,
    status         INTEGER NOT NULL
                   DEFAULT 0
                   CHECK(status IN (0,1,2)),
    create_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accept_time    DATETIME,
    deliver_time   DATETIME,

    PRIMARY KEY(id)
    FOREIGN KEY (rider_id)
        REFERENCES riders(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

