trap "exit" INT TERM ERR
trap "kill 0" EXIT

openresty -p `pwd`/ -c config/nginx.production.conf &
./bin/spacer start . production &

wait
