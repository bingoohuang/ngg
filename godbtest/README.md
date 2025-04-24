# godbtest

本工程本意用来验证使用 Go 语言，连接不同数据库，执行 SQL 的效果。

例如：

1. PostgreSQL 的 RETURNING 效果 `INSERT INTO COMPANY (NAME,AGE,ADDRESS,SALARY) VALUES('Paul', 32, 'California', 20000.00) RETURNING ID`
2. 编译安装支持 oracle  `TAGS=ora make -f ~/GitHub/bingoohuang/ngg/ver/Makefile`

```sh
> %connect pgx postgres://postgres:123456@192.168.126.16:14954;
2023/01/18 16:25:01 Connect to postgres://postgres:123456@192.168.126.16:14954 succeed
> CREATE TABLE person( id serial, lastname character varying (50), firstname character varying (50), CONSTRAINT person_pk PRIMARY KEY (id) );
2023/01/18 16:25:05 RunSQL: CREATE TABLE person( id serial, lastname character varying (50), firstname character varying (50), CONSTRAINT person_pk PRIMARY KEY (id) )
+---+--------------+--------------+-------------+
|   | lastInsertId | rowsAffected | cost        |
+---+--------------+--------------+-------------+
| 1 |            (N/A) |            0 | 15.159628ms |
+---+--------------+--------------+-------------+
> INSERT INTO person (lastname,firstname) VALUES ('Smith', 'John') RETURNING ID;
2023/01/18 16:25:09 RunSQL: INSERT INTO person (lastname,firstname) VALUES ('Smith', 'John') RETURNING ID
2023/01/18 16:25:09 Cost 15.467µs
+---+----+
|   | id |
+---+----+
| 1 |  1 |
+---+----+
2023/01/18 16:25:09 result saved to /var/folders/c8/ft7qp47d6lj5579gmyflxbr80000gn/T/103882282.txt
> select * from person;
2023/01/18 16:25:12 RunSQL: select * from person
2023/01/18 16:25:12 Cost 26.679µs
+---+----+----------+-----------+
|   | id | lastname | firstname |
+---+----+----------+-----------+
| 1 |  1 | Smith    | John      |
+---+----+----------+-----------+
2023/01/18 16:25:12 result saved to /var/folders/c8/ft7qp47d6lj5579gmyflxbr80000gn/T/4109013545.txt
```

## build

1. for oracle, `TAGS=ora make -f ~/github/ngg/ver/Makefile fmt install linux-upx`

```log
$ godbtest
> %help
2022/11/16 16:07:47 connect dm dm://SYSDBA:123456@127.0.0.1:5236?schema=demo;
2022/11/16 16:07:47 connect pgx postgres://SYSTEM:123456@127.0.0.1:54321/demo?sslmode=disable;
2022/11/16 16:07:47 connect mysql root:123456@(127.0.0.1:3306)/mysql?charset=utf8mb4&parseTime=true&loc=Local;
2022/11/16 16:07:47 connect sqlite :memory:;
2022/11/16 16:07:47 your SQL;
2022/11/16 16:07:47 begin/commit/rollback;
> %connect sqlite gowormhole.db;
> select * from sqlite_schema ;
2022/11/16 16:08:04 Connect to gowormhole.db succeed
2022/11/16 16:08:04 RunSQL: select * from sqlite_schema
2022/11/16 16:08:04 Cost 41.922µs
+---+-------+------------------------------------+-----------------+----------+---------------------------------+
| # | type  | name                               | tbl_name        | rootpage | sql                             |
+---+-------+------------------------------------+-----------------+----------+---------------------------------+
| 1 | table | gowormhole_recv                    | gowormhole_recv |        2 | CREATE TABLE gowormhole_recv(   |
|   |       |                                    |                 |          |         hash text not null,     |
|   |       |                                    |                 |          |         size integer not null,  |
|   |       |                                    |                 |          |         pos integer not null,   |
|   |       |                                    |                 |          |         expired datetime,       |
|   |       |                                    |                 |          |         updated datetime,       |
|   |       |                                    |                 |          |         name text not null,     |
|   |       |                                    |                 |          |         full text not null,     |
|   |       |                                    |                 |          |         hostname text,          |
|   |       |                                    |                 |          |         ips text,               |
|   |       |                                    |                 |          |         whoami text,            |
|   |       |                                    |                 |          |         cost text,              |
|   |       |                                    |                 |          |         primary key(hash)       |
|   |       |                                    |                 |          |     )                           |
| 2 | index | sqlite_autoindex_gowormhole_recv_1 | gowormhole_recv |        3 | <nil>                           |
+---+-------+------------------------------------+-----------------+----------+---------------------------------+
> select * from gowormhole_recv;
2022/11/16 16:08:08 RunSQL: select * from gowormhole_recv
2022/11/16 16:08:08 Cost 67.289µs
+---+----------------------+----------+----------+--------------------------------------+--------------------------------------+------------------------+------------------------+-----------------+--------------------------------------------+--------+---------------+
| # | hash                 |     size |      pos | expired                              | updated                              | name                   | full                   | hostname        | ips                                        | whoami | cost          |
+---+----------------------+----------+----------+--------------------------------------+--------------------------------------+------------------------+------------------------+-----------------+--------------------------------------------+--------+---------------+
| 1 | 798344184424802469   |  5806695 |  5806695 | 2022-11-16 19:08:09.198488 +0800 CST | 2022-11-15 19:08:15.350352 +0800 CST | 040.png                | 040.png                | VM-24-15-centos | 10.0.24.15                                 |        | 6.115024136s  |
| 2 | 14520279925969270108 |   491520 |   491520 | 2022-11-16 19:08:09.199397 +0800 CST | 2022-11-16 15:16:14.973846 +0800 CST | 061.png                | 061.png                | VM-24-15-centos | 10.0.24.15                                 |        | 19.349854ms   |
| 3 | 16835367656095118685 |  9838686 |  9838686 | 2022-11-16 19:08:09.200228 +0800 CST | 2022-11-15 19:08:27.7349 +0800 CST   | 090.png                | 090.png                | VM-24-15-centos | 10.0.24.15                                 |        | 11.630343101s |
| 4 | 18399982152889154493 | 80474875 | 79986688 | 2022-11-17 09:25:45.665634 +0800 CST | 2022-11-16 09:26:10.789849 +0800 CST | tdm64-gcc-10.3.0-2.exe | tdm64-gcc-10.3.0-2.exe | LAPTOP-06FTC5UC | 192.168.232.1,192.168.231.1,192.168.88.101 |        | 25.04572597s  |
+---+----------------------+----------+----------+--------------------------------------+--------------------------------------+------------------------+------------------------+-----------------+--------------------------------------------+--------+---------------+
> %connect sqlite :memory:;
2023/01/11 21:38:09 Connect to :memory: succeed
> select "1" str,2 num, 3.4 float, null;
2023/01/11 21:38:14 RunSQL: select "1" str,2 num, 3.4 float, null
2023/01/11 21:38:14 Cost 33.682µs
+---+-----+-----+-------+-------+
|   | str | num | float | null  |
+---+-----+-----+-------+-------+
| 1 | 1   | 2   | 3.4   | <nil> |
+---+-----+-----+-------+-------+
> select "1" str,2 num, 3.4 float, null\G
2023/01/11 21:38:21 RunSQL: select "1" str,2 num, 3.4 float, null
+---+--------+----------------+
|   | Column | value of row 1 |
+---+--------+----------------+
| 1 | str    | 1              |
| 2 | num    | 2              |
| 3 | float  | 3.4            |
| 4 | null   | <nil>          |
+---+--------+----------------+
2023/01/11 21:38:21 Cost 115.687µs
> select "1" str,2 num, 3.4 float, null\J
2023/01/11 21:38:24 RunSQL: select "1" str,2 num, 3.4 float, null
2023/01/11 21:38:24 Row 001: {"float":"3.4","null":null,"num":"2","str":"1"}
2023/01/11 21:38:24 Cost 85.139µs
> select "1" str,2 num, 3.4 float, null\I
2023/01/11 21:38:29 RunSQL: select "1" str,2 num, 3.4 float, null
insert into dual(str, num, float, null) values('1', '2', '3.4', null);
> exit
```

## relative resource

1. [danvergara/dblab](https://github.com/danvergara/dblab) The database client every command line junkie deserves. 每个命令行迷都应得的数据库客户端。
