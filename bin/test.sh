trap "exit" INT TERM ERR
trap "kill 0" EXIT

openresty -p `pwd`/ -c `pwd`/config/nginx.test.conf &
LOG_LEVEL=error ./bin/spacer test .
