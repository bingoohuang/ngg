# Postgresql & PgAdmin powered by compose

fork from [khezen/compose-postgres](https://github.com/khezen/compose-postgres)

## Requirements:

* docker >= 17.12.0+
* docker-compose

## Quick Start

* Clone or download this repository
* Go inside of directory,  `cd compose-postgres`
* Run this command `docker-compose up -d`

## Environments

This Compose file contains the following environment variables:

* `POSTGRES_USER` the default value is **postgres**
* `POSTGRES_PASSWORD` the default value is **changeme**
* `PGADMIN_PORT` the default value is **5050**
* `PGADMIN_DEFAULT_EMAIL` the default value is **pgadmin4@pgadmin.org**
* `PGADMIN_DEFAULT_PASSWORD` the default value is **admin**

## Access to postgres:

* `localhost:5432`
* **Username:** postgres (as a default)
* **Password:** changeme (as a default)

## Access to PgAdmin:

* **URL:** `http://localhost:5050`
* **Username:** pgadmin4@pgadmin.org (as a default)
* **Password:** admin (as a default)

## Add a new server in PgAdmin:

* **Host name/address** `postgres`
* **Port** `5432`
* **Username** as `POSTGRES_USER`, by default: `postgres`
* **Password** as `POSTGRES_PASSWORD`, by default `changeme`

## Logging

There are no easy way to configure pgadmin log verbosity and it can be overwhelming at times. It is possible to disable
pgadmin logging on the container level.

Add the following to `pgadmin` service in the `docker-compose.yml`:

```
logging:
  driver: "none"
```

[reference](https://github.com/khezen/compose-postgres/pull/23/files)

## 一键启动项目

`docker-compose up -d`

## 查看容器

```shell
$ docker-compose ps
NAME                 IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
pgadmin_container    dpage/pgadmin4      "/entrypoint.sh"         pgadmin             10 minutes ago      Up 10 minutes       443/tcp, 0.0.0.0:5050->80/tcp
postgres_container   postgres            "docker-entrypoint.s…"   postgres            10 minutes ago      Up 10 minutes       0.0.0.0:5432->5432/tcp
```

## 连接postgresql数据库

```sh
$ docker exec -it postgres_container bash
root@7646ced085b8:/# psql -U postgres -W
Password:
psql (15.3 (Debian 15.3-1.pgdg120+1))
Type "help" for help.

# 查询当前时间
postgres=# select now();
              now
-------------------------------
 2023-08-01 02:12:53.330125+00
(1 row)

# 查询亚洲/上海地区时间
postgres=# select  now() at time zone 'Asia/Shanghai';
          timezone
----------------------------
 2023-08-01 10:12:59.585708
(1 row)

# 设置postgres数据库的时区
postgres=# ALTER DATABASE "postgres" SET timezone TO 'Asia/Shanghai';
ALTER DATABASE
```

## 创建数据库

```shell
# 创建数据库xybdiy
postgres=# CREATE DATABASE xybdiy;
CREATE DATABASE
# 进入数据库xybdiy
postgres=# \connect xybdiy
Password:
You are now connected to database "xybdiy" as user "postgres"
# 查看已存在的数据库
xybdiy=# \list
                                                List of databases
   Name    |  Owner   | Encoding |  Collate   |   Ctype    | ICU Locale | Locale Provider |   Access privileges
-----------+----------+----------+------------+------------+------------+-----------------+-----------------------
 postgres  | postgres | UTF8     | en_US.utf8 | en_US.utf8 |            | libc            |
 template0 | postgres | UTF8     | en_US.utf8 | en_US.utf8 |            | libc            | =c/postgres          +
           |          |          |            |            |            |                 | postgres=CTc/postgres
 template1 | postgres | UTF8     | en_US.utf8 | en_US.utf8 |            | libc            | =c/postgres          +
           |          |          |            |            |            |                 | postgres=CTc/postgres
 xybdiy    | postgres | UTF8     | en_US.utf8 | en_US.utf8 |            | libc            |
```

## 创建表格

```sh
xybdiy=# CREATE TABLE COMPANY(
 ID INT PRIMARY KEY     NOT NULL,
 NAME           TEXT    NOT NULL,
 AGE            INT     NOT NULL,
 ADDRESS        CHAR(50),
 SALARY         REAL
);
CREATE TABLE
# 查看表格
xybdiy=# \display
                 List of relations
 Schema |     Name     | Type  |  Owner   |  Table
--------+--------------+-------+----------+---------
 public | company_pkey | index | postgres | company
(1 row)

# 查看表格
xybdiy=# \d company
                  Table "public.company"
 Column  |     Type      | Collation | Nullable | Default
---------+---------------+-----------+----------+---------
 id      | integer       |           | not null |
 name    | text          |           | not null |
 age     | integer       |           | not null |
 address | character(50) |           |          |
 salary  | real          |           |          |
Indexes:
    "company_pkey" PRIMARY KEY, btree (id)

xybdiy=# \d
          List of relations
 Schema |  Name   | Type  |  Owner
--------+---------+-------+----------
 public | company | table | postgres
(1 row)
```

## godbtest

build: `TAGS=ora,sqlite3,pgx make -f ../gg/Makefile linux-upx`

```sh
$ godbtest                                             
> %connect pgx 'postgres://postgres:changeme@127.0.0.1:5432/xybdiy?sslmode=disable';
> select * from COMPANY;
+----+------+-----+---------+--------+
| id | name | age | address | salary |
+----+------+-----+---------+--------+
+----+------+-----+---------+--------+
> insert into COMPANY(id, name, age, address, salary) values(1, '@姓名', 30, '@地址', 1000);
2023/08/01 10:26:31 Args: [0: 1][1: 柯衩滥][2: 30][3: 云南省德宏傣族景颇族自治州镊曂路3617号榾蒦小区4单元1235室][4: 1000]
+--------------+--------------+------------+----------------------------------------------+
| lastInsertId | rowsAffected | cost       | error                                        |
+--------------+--------------+------------+----------------------------------------------+
|            0 |            1 | 6.220823ms | LastInsertId is not supported by this driver |
+--------------+--------------+------------+----------------------------------------------+
> select * from COMPANY;
+----+--------+-----+----------------------------------------------------------------------------+--------+
| id | name   | age | address                                                                    | salary |
+----+--------+-----+----------------------------------------------------------------------------+--------+
|  1 | 柯衩滥 |  30 | 云南省德宏傣族景颇族自治州镊曂路3617号榾蒦小区4单元1235室                  |   1000 |
+----+--------+-----+----------------------------------------------------------------------------+--------+
>  
```
