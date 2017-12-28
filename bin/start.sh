trap "exit" INT TERM ERR
trap "kill 0" EXIT

openresty -p `pwd`/ -c nginx.conf &
sleep 1
tail -0 -f logs/access.log &
sleep 1
tail -0 -f logs/error.log &

wait
