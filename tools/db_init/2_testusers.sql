-- test account (name: admin / password: password / admin)
insert into users(name, passhash, admin) values ('admin', '$2a$10$AqftzLHYcaGH2GxUXiGO/OzHnIMJO.PGMrLFqm7mPbpqZlQrIRrq.', true);

-- test account (name: judge / password: password / admin)
insert into users(name, passhash, admin) values ('judge', '$2a$10$AqftzLHYcaGH2GxUXiGO/OzHnIMJO.PGMrLFqm7mPbpqZlQrIRrq.', true);

-- test account (name: upload / password: password / admin)
insert into users(name, passhash, admin) values ('upload', '$2a$10$AqftzLHYcaGH2GxUXiGO/OzHnIMJO.PGMrLFqm7mPbpqZlQrIRrq.', true);

-- test account (name: tester / password: password)
insert into users(name, passhash, admin) values ('tester', '$2a$10$AqftzLHYcaGH2GxUXiGO/OzHnIMJO.PGMrLFqm7mPbpqZlQrIRrq.', false);
