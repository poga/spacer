trap "exit" INT TERM ERR
trap "kill 0" EXIT

openresty -p `pwd`/ -c nginx.conf &
sleep 1
tail -0 -f logs/access.log &
tail -0 -f logs/error.log | awk '
  /[error]/ {print "\033[31m" $0 "\033[39m"}
' &

wait
