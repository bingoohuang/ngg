# uuid in postgre

[blog](https://maciejwalkowiak.com/blog/postgres-uuid-primary-key/)

索引占用内存

| table        | index             | table size | index size | 一千万行生成时间 |
|--------------|-------------------|------------|------------|----------|
| bank         | bank_pkey         | 651 MB     | 733 MB     | 2m32s    |
| bank_uuid    | bank_uuid_pkey    | 422 MB     | 386 MB     | 1m54s    |
| bank_uuid_v7 | bank_uuid_v7_pkey | 422 MB     | 383 MB     | 1m36s    |


1. Postgres has a dedicated data type for UUIDs: uuid. UUID is a 128 bit (16 bytes) data type.
2. UUID v4 is a pseudo-random value. UUID v7 produces time-sorted values. It means that each time new UUID v7 is generated, a greater value it has. And that makes it a good fit for B-Tree index.

```sql
-- 建表
create table bank(id text primary key);
create table bank_uuid(id uuid primary key);
create table bank_uuid_v7(id uuid primary key);
create table bank_tsid(id numeric(20, 0) primary key);

-- 查询表和索引大小
select
     relname as "table",
     indexrelname as "index",
     pg_size_pretty(pg_relation_size(relid)) "table size",
     pg_size_pretty(pg_relation_size(indexrelid)) "index size"
from
    pg_stat_all_indexes
where
    relname not like 'pg%';
```

```sh
$ docker run --name pg -e POSTGRES_USER=postgre -e POSTGRES_DB=postgre -e POSTGRES_PASSWORD=postgre -p 1234:5432 -d postgres:16.3
f7d835e2a15db04f8263a16caf67d77b701e25b98457a92666f63d7f3d40f9fb

$ godbtest
> %connect pgx://postgre:postgre@127.0.0.1:1234/postgre?sslmode=disable;
> create table bank(id text primary key);
2024/07/13 21:33:24 Result: lastInsertId: (N/A), rowsAffected: 0, cost: 7.634822ms

> insert into bank(id) values('@uuid');
2024/07/13 21:35:33 SQL: insert into bank(id) values ($1) ::: Args: ["3931d7e2-9d7e-4164-8512-123a778e…"]
2024/07/13 21:35:33 Result: lastInsertId: (N/A), rowsAffected: 1, cost: 3.586089ms
> %set -n 10000000;
> insert into bank(id) values('@uuid')\P
2024/07/13 21:35:53 threads(goroutines): 12, txBatch: 1, batch: 100 with 10000000 request(s)
2024/07/13 21:35:53 preparedQuery: insert into bank(id)  values ($1)
10000000 / 10000000 [----------------------------------------------------------------] 100.00% 65675 p/s 2m32s
2024/07/13 21:38:25 Average 15.246µs/record, total cost: 2m32.46461615s, total affected: 10000000, errors: 0

> create table bank_uuid(id uuid primary key);
2024/07/13 21:34:32 Result: lastInsertId: (N/A), rowsAffected: 0, cost: 4.961273ms

> insert into bank_uuid(id) values('@uuid');
2024/07/13 21:38:41 SQL: insert into bank_uuid(id) values ($1) ::: Args: ["98394467-22ec-461b-bc08-ab50c29d…"]
2024/07/13 21:38:41 Result: lastInsertId: (N/A), rowsAffected: 1, cost: 5.561812ms
> %set -n 10000000;
> insert into bank_uuid(id) values('@uuid')\P
2024/07/13 21:40:42 threads(goroutines): 12, txBatch: 1, batch: 100 with 10000000 request(s)
2024/07/13 21:40:42 preparedQuery: insert into bank_uuid(id)  values ($1)
10000000 / 10000000 [------------------------------------------------------------------] 100.00% 87629 p/s 1m54s
2024/07/13 21:42:36 Average 11.431µs/record, total cost: 1m54.318030121s, total affected: 10000000, errors: 0

> create table bank_uuid_v7(id uuid primary key);
2024/07/13 21:44:08 Result: lastInsertId: (N/A), rowsAffected: 0, cost: 5.251364ms

> insert into bank_uuid_v7(id) values('@uuid(v7)');
2024/07/13 21:44:29 SQL: insert into bank_uuid_v7(id) values ($1) ::: Args: ["0091a087-3a20-4812-9800-5a9f4137…"]
2024/07/13 21:44:29 Result: lastInsertId: (N/A), rowsAffected: 1, cost: 2.15823ms
> insert into bank_uuid_v7(id) values('@uuid(v7)')\P
2024/07/13 21:44:38 threads(goroutines): 12, txBatch: 1, batch: 100 with 10000000 request(s)
2024/07/13 21:44:38 preparedQuery: insert into bank_uuid_v7(id)  values ($1)
10000000 / 10000000 [-------------------------------------------------------------------] 100.00% 104859 p/s 1m36s
2024/07/13 21:46:14 Average 9.556µs/record, total cost: 1m35.566259666s, total affected: 10000000, errors: 0
> select
>>>     relname as "table",
>>>     indexrelname as "index",
>>>     pg_size_pretty(pg_relation_size(relid)) "table size",
>>>     pg_size_pretty(pg_relation_size(indexrelid)) "index size"
>>> from
>>>     pg_stat_all_indexes
>>> where
>>>     relname not like 'pg%';
+--------------+-------------------+------------+------------+
| table        | index             | table size | index size |
+--------------+-------------------+------------+------------+
| bank         | bank_pkey         | 651 MB     | 733 MB     |
| bank_uuid    | bank_uuid_pkey    | 422 MB     | 386 MB     |
| bank_uuid_v7 | bank_uuid_v7_pkey | 422 MB     | 383 MB     |
+--------------+-------------------+------------+------------+

> select * from bank limit 1;
+--------------------------------------+
| id                                   |
+--------------------------------------+
| 3931d7e2-9d7e-4164-8512-123a778e81df |
+--------------------------------------+
> select * from bank_uuid limit 1;
+--------------------------------------+
| id                                   |
+--------------------------------------+
| a1e425d7-942f-4a6a-8ea7-79bf58897712 |
+--------------------------------------+
> select * from bank_uuid_v7 limit 1;
+--------------------------------------+
| id                                   |
+--------------------------------------+
| 0091a087-3a20-4812-9800-5a9f41378653 |
+--------------------------------------+
```
