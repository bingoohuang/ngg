-- godbtest 连接 oralce
%connect oracle://cyz:f6fa9a861b90HB@127.0.0.1:39964/BJCADB;

-- 创建 card 表主键序列
drop SEQUENCE seq_card_id_gen;
CREATE SEQUENCE seq_card_id_gen START WITH 1 INCREMENT BY 1 ORDER;
-- 创建 get_next_card 函数所用的序列
drop SEQUENCE seq_card;
CREATE SEQUENCE seq_card START WITH 1 INCREMENT BY 1 CACHE 1000 ORDER;
-- 创建 card 表
drop table t_card;
CREATE TABLE t_card (id INT PRIMARY KEY, uuid VARCHAR2(36), state int default 0);

-- 设置换行分隔符为/
%set --sep '/'

-- 创建函数 f_card
CREATE OR REPLACE FUNCTION f_card RETURN VARCHAR2 IS
  v_id INT;
  v_uuid VARCHAR2(36);
  PRAGMA AUTONOMOUS_TRANSACTION;
BEGIN
  -- 从序列中获取下一个值
  SELECT seq_card.NEXTVAL INTO v_id FROM DUAL;

  -- 假设state字段为0表示未分配，我们更新id为v_id的记录
  -- 这里假设card表的uuid是唯一的，且id和uuid是一对
  UPDATE t_card SET state = 200 WHERE id = v_id returning uuid INTO v_uuid;
  COMMIT;
  -- 返回uuid值
  RETURN v_uuid;
EXCEPTION
  WHEN NO_DATA_FOUND THEN
    -- 如果没有找到记录，返回NULL或者抛出异常
    RETURN NULL;
END;
/

-- 恢复默认换行分隔符 ;
%set --sep ';'

-- 向 card 表打入10万数据
%set -n 100000
insert into t_card (id, uuid) values(seq_card_id_gen.NEXTVAL, '@uuid')\P

-- 单词验证
select f_card() from dual;

-- 执行压测
%set -n 100000
select f_card() from dual\P
-- TPS: 11380
