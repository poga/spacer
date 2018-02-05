trap "exit" INT TERM ERR
trap "kill 0" EXIT

openresty -p `pwd`/ -c `pwd`/config/nginx.production.conf &
./bin/spacer start . production &

wait
