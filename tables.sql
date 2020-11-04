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
  testhash varchar(255)
);

create table submissions (
  id serial primary key,
  submit_time timestamp with time zone,
  user_name varchar(255) references users(name),
  problem_name varchar(255) not null references problems(name),
  lang varchar(32) not null,
  source text not null,
  status varchar(32), -- AC, WA, TLE, WJ, ...
  prev_status varchar(32),
  hacked boolean,
  testhash varchar(255),
  compile_error text,
  max_time int,
  max_memory bigint,
  judge_ping timestamp with time zone,
  judge_name varchar(255) not null
);

create table tasks ( -- Between front and judge
  id serial primary key,
  priority int not null,
  available timestamp with time zone,
  submission int not null references submissions(id)
);

create table submission_testcase_results (
  submission int references submissions(id),       -- primary main
  testcase varchar(63), -- primary sub
  status varchar(32),
  time int,
  memory bigint,
  primary key(submission, testcase)
);
