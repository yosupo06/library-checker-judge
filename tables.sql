create table users (
  id int primary key,
  name varchar(255) unique not null,
  passhash varchar(255)
);

create table problems (
  name varchar(255) primary key,
  timelimit int,
  testhash varchar(255),
  testzip bytea
);

create table submittions (
  id serial primary key,
  submittime timestamp,
  userid int, -- null ok
  problem varchar(32) not null,
  lang varchar(32) not null,
  source text not null,
  status varchar(32), -- AC, WA, TLE, WJ, ...
  maxtime int,
  maxmemory int,
  details json
);

create table tasks ( -- Between front and judge
  id serial primary key,
  submittion int not null
);
