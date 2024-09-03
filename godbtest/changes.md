# 变更

## 2024年06月12日

支持随机表情符号 `@emoji` `@emoji(3)` `@emoji(3-5)`

## 2024年05月31日

支持随机值 SQL 生成

```sh
> %set --dry --times 3 --noPrepared
> insert into t2(id,name,addr,email,phone,age,idc) values('@ksuid','@姓名','@地址','@邮箱','@手机','@random_int(15-95)','@身份证');
2024/05/31 10:11:53 SQL: insert into t2(id,name,addr,email,phone,age,idc) values('2hDH1gFg9xv1C5fRFbo0t50iFO0','舒尬莡','四川省攀枝花市攋瑚路3496号宼驼小区11单元1962室','wlfbdcbq@hotio.space','15247485288','64','36114619740719808X')
2024/05/31 10:11:53 SQL: insert into t2(id,name,addr,email,phone,age,idc) values('2hDH1ZGq6y1xRL4gKRh9wGjfMBD','于箸駆','四川省攀枝花市椥谜路4189号鞨賗小区8单元2095室','bizgxocq@eqjow.biz','13656833724','36','21568519930103808X')
2024/05/31 10:11:53 SQL: insert into t2(id,name,addr,email,phone,age,idc) values('2hDH1YjHXtaAqf0kgkZMwZ7OBP1','吴碀眮','广西壮族自治区南宁市鏴髍路7675号橥誌小区9单元1122室','mgnbhhqq@brdrc.info','14735371589','87','114163198505280287')
```

## 2024年01月02日

支持以下 sql

```sh
%connect sqlite://idhash.db

create table idhash(id text PRIMARY KEY, sm3 text, xxhash text, sha256 text, sha512 text, md5 text);
create INDEX idx_sm3 ON idhash(sm3);
create INDEX idx_xxhash on idhash(xxhash);
create INDEX idx_sha256 on idhash(sha256);
create INDEX idx_sha512 on idhash(sha512);
create INDEX idx_md5 on idhash(md5);

insert into idhash(id, sm3, xxhash, sha256, sha512, md5) values('@身份证', '@sm3(身份证)', '@xxhash(身份证)', '@sha256(身份证)', '@sha512(身份证)', '@md5(身份证)')\P

SELECT sm3, COUNT(sm3) AS count FROM idhash GROUP BY sm3 HAVING COUNT(sm3) > 1;
SELECT xxhash, COUNT(xxhash) AS count FROM idhash GROUP BY xxhash HAVING COUNT(xxhash) > 1;
SELECT sha256, COUNT(sha256) AS count FROM idhash GROUP BY sha256 HAVING COUNT(sha256) > 1;
SELECT sha512, COUNT(sha512) AS count FROM idhash GROUP BY sha512 HAVING COUNT(sha512) > 1;
SELECT md5, COUNT(md5) AS count FROM idhash GROUP BY md5 HAVING COUNT(md5) > 1;
```

```sh
> %connect mysql://root:root@127.0.0.1:13306/bingoo;

CREATE TABLE idhash (id INT AUTO_INCREMENT PRIMARY KEY, idcard VARCHAR(255), sm3 VARCHAR(255));
insert into idhash(id, sm3) values('@身份证', '@sm3(身份证)')\P

30315600 / 30271600 [-------------------------] 100.15% 26947 p/s 18m45s
2024/01/02 14:23:31 Average 37.116µs/record, total cost: 18m45.220245831s, total affected: 30315600, errors: 0

[root@k8s-master bingoo]# ls -hl
总用量 5.6G
-rw-r-----. 1 root root 5.6G  1月  2 14:23 idhash.ibd
[root@k8s-master bingoo]# ls -hl

CREATE INDEX idx_sm3 ON idhash(sm3);

> CREATE INDEX idx_sm3 ON idhash(sm3);
+--------------+--------------+-----------------+
| lastInsertId | rowsAffected | cost            |
+--------------+--------------+-----------------+
|            0 |            0 | 6m30.342631813s |
+--------------+--------------+-----------------+

[root@k8s-master bingoo]# ls -hl
总用量 7.4G
-rw-r-----. 1 root root 7.4G  1月  2 14:30 idhash.ibd

```

## 2023年10月23日

支持 mssql

```shell
%connect 'mssql://sa:sa@172.16.14.129?database=学生信息';
```

## 2023年06月28日

支持 blob 等字段类型的写入与查询

```shell
> create table t_face (
    id   bigint auto_increment primary key comment 'ID',
    face MEDIUMBLOB comment '人脸照片, Up to 16 Mb'
) engine = innodb default charset = utf8mb4 comment '人脸表';
> insert into t_face(face) values('@file(/Volumes/e2t/500px/ai640.png,:bytes)');
2023/06/28 18:12:22 Args: [�PNG

IHDR����gAMA��
              �a cHRMz&�����u0�`:�p��Q<bKGD��������IDATx���i�m[v...]
+--------------+--------------+--------------+
| lastInsertId | rowsAffected | cost         |
+--------------+--------------+--------------+
|            7 |            1 | 190.215342ms |
+--------------+--------------+--------------+
> %writeLob --dir /Users/bingoo/aaa --ext .png;
> select * from t_face;
2023/06/28 18:09:28 Args:
+----+----------------------------------------+
| id | face                                   |
+----+----------------------------------------+
|  1 | (see /Users/bingoo/aaa/1589769177.png) |
|  3 | (see /Users/bingoo/aaa/2615222146.png) |
+----+----------------------------------------+

```

## 2023年01月18日

自动输出 CSV/MD/HTML/SQL 临时文件.
