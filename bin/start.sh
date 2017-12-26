trap "exit" INT TERM ERR
trap "kill 0" EXIT

openresty -p `pwd`/ -c nginx.conf &
sleep 1
tail -f logs/access.log &
sleep 1
tail -f logs/error.log &

wait