CREATE TABLE profiles (
   id int(11) not null primary key AUTO_INCREMENT,
   username varchar(255) not null,
   type varchar(255) not null,
   path varchar(255) not null,
   publickey varchar(255) not null,
   privatekey varchar(255) not null,
   config text,
   qrcode text,
   is_active bool default true
);