create database if not exists gcache;
use gcache;
create table if not exists gcache(k varchar(300) primary key, v varchar(300));
insert into gcache(k, v) values('k1', 'mysql value');
