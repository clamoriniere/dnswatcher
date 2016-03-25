#!/opt/bin/bash

user="username"
pass="password"
host="host dns"
email="email"
smtp="smtp"

trap "rm -f /tmp/dnswatcher.lock" SIGINT SIGTERM
if [ -e /tmp/dnswatcher.lock ]
then
  echo "dnswatcher is running already."
  exit 1
else
  touch /tmp/dnswatcher.lock
  /usr/bin/dnswatcher -logtostderr -email $email -host $host -smtp $smtp -user $user -password $pass
  rm -f /tmp/dnswatcher.lock
  trap - SIGINT SIGTERM
  exit 0
fi
