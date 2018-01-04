trap "exit" INT TERM ERR
trap "kill 0" EXIT

openresty -p `pwd`/ -c config/nginx.conf &
spacer start . &
touch logs/access.log
touch logs/error.log

tail -0f logs/access.log | awk '
  /\[400\]/{print "\033[33m" $0 "\033[39m"; next}
  /\[500\]/{print "\033[33m" $0 "\033[39m"; next}
  /\[404\]/{print "\033[33m" $0 "\033[39m"; next}
  1 {print "\033[32m" $0 "\033[39m"; next}
' &
tail -0f logs/error.log | awk '{print "\033[31m" $0 "\033[39m"}' &

wait
