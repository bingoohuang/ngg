-- docker run --name mysql -e MYSQL_ROOT_PASSWORD=root -p 3306:3306 -d mysql:8.4.0-oracle

%connect mysql://root:root@127.0.0.1:3306

drop database if exists card;
create database card;
use card;

drop table if exists t_card;
drop table if exists t_seq;

CREATE TABLE `t_card` (
  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY ,
  `state` int DEFAULT '0',
  `uuid` varchar(100) NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4;

CREATE TABLE `t_seq` (
  `name` varchar(100) NOT NULL PRIMARY KEY ,
  `num` int DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

drop procedure if exists p_card;

%set --sep '//'

CREATE PROCEDURE p_card()
BEGIN
    -- 声明局部变量
    DECLARE updatedNum INT DEFAULT 0;
    DECLARE returnedSeq INT;
    DECLARE returnedUuid VARCHAR(100);
    DECLARE affectedRows INT DEFAULT 0;

    -- 更新seq表，并获取影响的行数
    UPDATE t_seq
    SET num = num + 1
    WHERE name = '步兵';

    SET affectedRows = ROW_COUNT();

    -- 检查是否有行被更新
    IF affectedRows = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '步兵序号未设置';
    ELSE
        -- 获取更新后的num值
        SELECT num INTO updatedNum
        FROM t_seq
        WHERE name = '步兵'
        LIMIT 1;
    END IF;

    COMMIT;

    -- 使用变量updatedNum更新card表，并获取影响的行数
    UPDATE t_card
    SET state = 1
    WHERE id = updatedNum;

    SET affectedRows = ROW_COUNT();

    -- 检查是否有行被更新
    IF affectedRows = 0 THEN
        SIGNAL SQLSTATE '45001' SET MESSAGE_TEXT = '卡资源不足';
    ELSE
        -- 选择被更新行的seq和uuid字段
        SELECT id, uuid INTO returnedSeq, returnedUuid
        FROM t_card
        WHERE id = updatedNum
        LIMIT 1;
    END IF;

    COMMIT;

    -- 返回被更新的seq和uuid字段
    SELECT returnedSeq as seq, returnedUuid as uuid;
END //

%set --sep ';'

insert into t_seq(name) values ('步兵');

-- 在 card 表中插入100万随机数据
%connect mysql://root:root@127.0.0.1:3306/card
%set -n 100000
insert into t_card(uuid) values ('@uuid')\P

-- 压测10万次执行
%set -n 10000 --asQuery
CALL p_card()\P
-- 314 p/s
