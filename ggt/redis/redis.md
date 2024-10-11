# redis

1. 通过模板 set 值: `ggt redis -s 192.168.56.110:6379 -p 1qazzaq1 set -k __gateway_redis__ -f api_version_update_time -v '{{unixTime}}' --tmpl`
