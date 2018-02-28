#!/bin/bash
trap "exit" INT TERM ERR
trap "kill 0" EXIT

spacer nginx-config . --env production > `pwd`/config/nginx.production.conf

SPACER_ENV=production openresty -p `pwd`/ -c `pwd`/config/nginx.production.conf &
LOG_LEVEL=error spacer start . --env production &

wait
