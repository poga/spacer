#!/bin/bash
trap "exit" INT TERM ERR
trap "kill 0" EXIT

SPACER_ENV=production openresty -p `pwd`/ -c `pwd`/config/nginx.production.conf &
LOG_LEVEL=error spacer start . --env production &

wait
