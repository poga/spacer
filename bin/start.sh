trap "exit" INT TERM ERR
trap "kill 0" EXIT

openresty -p `pwd`/ -c `pwd`/config/nginx.development.conf &
./bin/spacer start . development &
touch log/access.log
touch log/error.log

tail -0f log/access.log | awk '
  /\[400\]/{print "\033[33m" $0 "\033[39m"; next}
  /\[500\]/{print "\033[33m" $0 "\033[39m"; next}
  /\[404\]/{print "\033[33m" $0 "\033[39m"; next}
  1 {print "\033[32m" $0 "\033[39m"; next}
' &
tail -0f log/error.log | awk '{print "\033[31m" $0 "\033[39m"}' &

wait
