CREATE TABLE users (
    id int(11) not null primary key AUTO_INCREMENT,
    email varchar(320) not null unique,
    password_hash varchar(255) not null,
    role varchar(63) default null
);