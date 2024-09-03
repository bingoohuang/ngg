%connect sqlite://idhash.db

create table idhash(id text PRIMARY KEY, sm3 text, xxhash text, sha256 text, sha512 text, md5 text);
create INDEX idx_sm3 ON idhash(sm3);
create INDEX idx_sha256 on idhash(sha256);
create INDEX idx_sha512 on idhash(sha512);
create INDEX idx_md5 on idhash(md5);
create INDEX idx_xxhash on idhash(xxhash);
PRAGMA index_list(idhash);
select * from sqlite_master;
insert into idhash(id, sm3, xxhash, sha256, sha512, md5) values('@身份证', '@sm3(身份证)', '@xxhash(身份证)', '@sha256(身份证)', '@sha512(身份证)', '@md5(身份证)');
select * from idhash;
insert into idhash(id, sm3, xxhash, sha256, sha512, md5) values('@身份证', '@sm3(身份证)', '@xxhash(身份证)', '@sha256(身份证)', '@sha512(身份证)', '@md5(身份证)')\P
%set -h
%perf -h
select count(*) from idhash;
SELECT sm3, COUNT(sm3) AS count
FROM idhash
GROUP BY sm3
HAVING COUNT(sm3) > 1;
select xxhash, count(xxhash) from idhash group by xxhash having count(xxhash) > 1;
SELECT md5, COUNT(md5) AS count FROM idhash GROUP BY md5 HAVING COUNT(md5) > 1;
SELECT sha512, COUNT(sha512) AS count FROM idhash GROUP BY sha512 HAVING COUNT(sha512) > 1;
SELECT sha256, COUNT(sha256) AS count FROM idhash GROUP BY sha256 HAVING COUNT(sha256) > 1;
