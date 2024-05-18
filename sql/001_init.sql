DROP TABLE IF EXISTS users;

CREATE TABLE users (
    username TEXT PRIMARY KEY,
    salt TEXT NOT NULL,
    hash TEXT NOT NULL
);

-- generated with `hello-go pw abc123`
INSERT INTO users VALUES (
    'alice',
    '82d7e22f2bac14dc314c8282b73fe0d6d427cd14b30dd4ea48dfaf96a1b2d551',
    '4de35aa8986deb0e93e10d8a9601972067d181631a3cc19e292fc695a1baa3d8'
);
-- generated with `hello-go pw 123abc`
INSERT INTO users VALUES (
    'bob',
    '48b67c6ea50720806cbce4fac4c96534b07b382e2603bc3df225f3cbc5633aa6',
    '80479c04a06f3921439a5477afe262ee40bb6f6a18fe6e0fd53bf83ac44be95e'
);
-- generated with `hello-go pw foo123`
INSERT INTO users VALUES (
    'carol',
    '598b7c2cd8c03c377f16327fe7f1b546ec302140e173a14721227081408b30db',
    '26cccce9811d242e5dee43d04319909f1109005fcfc423f9a4d3586fe844ca90'
);
-- generated with `hello-go pw 123foo`
INSERT INTO users VALUES (
    'dan',
    'cefd44900cbbe5ebffdf95455c12699d6fdace90d2a3dbcf2989fa66b07677ff',
    '0ca411ac81a0f8c452978b4e33b242f634610c2900d75058c1bff324928e5210'
);
