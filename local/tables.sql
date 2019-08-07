create table users (  
  name varchar(255) primary key,
  passhash varchar(255),
  admin boolean
);

create table problems (
  name varchar(255) primary key,
  title varchar(255),
  statement text,
  timelimit int,
  testhash varchar(255),
  testzip bytea
);

create table submissions (
  id serial primary key,
  submit_time timestamp,
  user_name varchar(255) references users(name),
  problem varchar(32) not null references problems(name),
  lang varchar(32) not null,
  source text not null,
  status varchar(32), -- AC, WA, TLE, WJ, ...
  max_time int,
  max_memory int,
  details json
);

create table tasks ( -- Between front and judge
  id serial primary key,
  submission int not null
);

create table submission_testcase_results (
  submission int references submissions(id),       -- primary main
  testcase varchar(32), -- primary sub
  status varchar(32),
  time int,
  memory int,
  primary key(submission, testcase)
);
