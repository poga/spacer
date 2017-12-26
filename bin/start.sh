trap "exit" INT TERM ERR
trap "kill 0" EXIT

openresty -p `pwd`/ -c nginx.conf &
sleep 1
spacer start . &
spacer pretty-log . &

wait