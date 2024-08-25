# gops

```shell
[root@192_168_126_5 ~]# ps -eo pid,lstart,cmd | grep 25391
25391 Thu Dec 21 14:03:54 2023 /usr/java/jdk1.8.0_191/bin/java -Xms512m -Xmx512m -XX:+UseConcMarkSweepGC -XX:CMSInitiatingOccupancyFraction=75 -XX:+UseCMSInitiatingOccupancyOnly -Des.networkaddress.cache.ttl=60 -Des.networkaddress.cache.negative.ttl=10 -XX:+AlwaysPreTouch -Xss1m -Djava.awt.headless=true -Dfile.encoding=UTF-8 -Djna.nosys=true -XX:-OmitStackTraceInFastThrow -Dio.netty.noUnsafe=true -Dio.netty.noKeySetOptimization=true -Dio.netty.recycler.maxCapacityPerThread=0 -Dlog4j.shutdownHookEnabled=false -Dlog4j2.disable.jmx=true -Djava.io.tmpdir=${ES_TMPDIR} -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=data -XX:ErrorFile=logs/hs_err_pid%p.log -Des.path.home=/usr/local/elasticsearch-5.5.3 -cp /usr/local/elasticsearch-5.5.3/lib/* org.elasticsearch.bootstrap.Elasticsearch -p /usr/local/elasticsearch-5.5.3/elasticsearch.pid --quiet
26659 Mon Jan  8 17:49:22 2024 grep --color=auto 25391

[root@192_168_126_5 0]# gops -pid 25391
2024/01/08 18:08:55 Ppid: 1, Executable: java
2024/01/08 18:08:55 Cmdline: [/usr/java/jdk1.8.0_191/bin/java -Xms512m -Xmx512m -XX:+UseConcMarkSweepGC -XX:CMSInitiatingOccupancyFraction=75 -XX:+UseCMSInitiatingOccupancyOnly -Des.networkaddress.cache.ttl=60 -Des.networkaddress.cache.negative.ttl=10 -XX:+AlwaysPreTouch -Xss1m -Djava.awt.headless=true -Dfile.encoding=UTF-8 -Djna.nosys=true -XX:-OmitStackTraceInFastThrow -Dio.netty.noUnsafe=true -Dio.netty.noKeySetOptimization=true -Dio.netty.recycler.maxCapacityPerThread=0 -Dlog4j.shutdownHookEnabled=false -Dlog4j2.disable.jmx=true -Djava.io.tmpdir=${ES_TMPDIR} -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=data -XX:ErrorFile=logs/hs_err_pid%p.log -Des.path.home=/usr/local/elasticsearch-5.5.3 -cp /usr/local/elasticsearch-5.5.3/lib/* org.elasticsearch.bootstrap.Elasticsearch -p /usr/local/elasticsearch-5.5.3/elasticsearch.pid --quiet]
2024/01/08 18:08:55 CmdlineSlice: [[/usr/java/jdk1.8.0_191/bin/java -Xms512m -Xmx512m -XX:+UseConcMarkSweepGC -XX:CMSInitiatingOccupancyFraction=75 -XX:+UseCMSInitiatingOccupancyOnly -Des.networkaddress.cache.ttl=60 -Des.networkaddress.cache.negative.ttl=10 -XX:+AlwaysPreTouch -Xss1m -Djava.awt.headless=true -Dfile.encoding=UTF-8 -Djna.nosys=true -XX:-OmitStackTraceInFastThrow -Dio.netty.noUnsafe=true -Dio.netty.noKeySetOptimization=true -Dio.netty.recycler.maxCapacityPerThread=0 -Dlog4j.shutdownHookEnabled=false -Dlog4j2.disable.jmx=true -Djava.io.tmpdir=${ES_TMPDIR} -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=data -XX:ErrorFile=logs/hs_err_pid%p.log -Des.path.home=/usr/local/elasticsearch-5.5.3 -cp /usr/local/elasticsearch-5.5.3/lib/* org.elasticsearch.bootstrap.Elasticsearch -p /usr/local/elasticsearch-5.5.3/elasticsearch.pid --quiet]]
2024/01/08 18:08:55 Username: [elasticsearch]
2024/01/08 18:08:55 Cwd: [/usr/local/elasticsearch-5.5.3]
2024/01/08 18:08:55 Exe: [/usr/local/jdk1.8.0_191/bin/java]
2024/01/08 18:08:55 CPUPercent: [0.192545]
2024/01/08 18:08:55 CreateTime: [2023-12-21 14:03:54]
2024/01/08 18:08:55 Background: [true]
2024/01/08 18:08:55 Name: [java]
2024/01/08 18:08:55 String: [{"pid":25391}]
2024/01/08 18:08:55 Status: [[sleep]]
2024/01/08 18:08:55 Environ: [HOSTNAME=192_168_126_5]
2024/01/08 18:08:55 Environ: [SHELL=/sbin/nologin]
2024/01/08 18:08:55 Environ: [DATA_DIR=/usr/local/elasticsearch-5.5.3/data]
2024/01/08 18:08:55 Environ: [USER=elasticsearch]
2024/01/08 18:08:55 Environ: [ES_HOME=/usr/local/elasticsearch-5.5.3]
2024/01/08 18:08:55 Environ: [PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin]
2024/01/08 18:08:55 Environ: [CONF_DIR=/usr/local/elasticsearch-5.5.3/config]
2024/01/08 18:08:55 Environ: [PWD=/usr/local/elasticsearch-5.5.3]
2024/01/08 18:08:55 Environ: [LOG_DIR=/usr/local/elasticsearch-5.5.3/logs]
2024/01/08 18:08:55 Environ: [JAVA_HOME=/usr/java/jdk1.8.0_191]
2024/01/08 18:08:55 Environ: [LANG=en_US.UTF-8]
2024/01/08 18:08:55 Environ: [ES_JVM_OPTIONS=/usr/local/elasticsearch-5.5.3/config/jvm.options]
2024/01/08 18:08:55 Environ: [SHLVL=0]
2024/01/08 18:08:55 Environ: [HOME=/nonexistent]
2024/01/08 18:08:55 Environ: [LOGNAME=elasticsearch]
2024/01/08 18:08:55 Environ: [ES_PATH_CONF=/usr/local/elasticsearch-5.5.3/config]
2024/01/08 18:08:55 Environ: [PID_DIR=/usr/local/elasticsearch-5.5.3]
2024/01/08 18:08:55 IsRunning: [true]
2024/01/08 18:08:55 MemoryInfo: [{"rss":262995968,"vms":6653075456,"hwm":0,"data":0,"stack":0,"locked":0,"swap":0}]
2024/01/08 18:08:55 NumFDs: [321]
2024/01/08 18:08:55 NumThreads: [66]
2024/01/08 18:08:55 Children: [[]]
```

## chatGPT

```sh
User:
ps -o lstart  中的  lstart 是 UTC 时间，还是本地时间？

ChatGPT:
ps -o lstart 命令显示的 lstart 字段通常表示进程启动的时间，但该时间是以本地时间（Local Time）而非UTC时间（Coordinated Universal Time）来表示的。在这个上下文中，lstart 显示的是进程启动的本地时间。
```
