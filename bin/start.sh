trap "exit" INT TERM ERR
trap "kill 0" EXIT

openresty -p `pwd`/ -c nginx.conf &
spacer start . &

wait