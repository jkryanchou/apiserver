#!/bin/sh
#
#       /etc/rc.d/init.d/apiserver
#
#       apiserver daemon
#
# chkconfig:   2345 95 05
# description: a apiserver script

### BEGIN INIT INFO
# Provides:       apiserver
# Required-Start: $network
# Required-Stop:
# Should-Start:
# Should-Stop:
# Default-Start: 2 3 4 5
# Default-Stop:  0 1 6
# Short-Description: start and stop apiserver
# Description: a apiserver script
### END INIT INFO

set -e

PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:${PATH}
PIDFILE=/var/run/apiserver.pid
DIRECTORY=/home/phuslu/apiserver
SUDO=$(test $(id -u) = 0 || echo sudo)

if [ -n "${SUDO}" ]; then
    echo "ERROR: Please run as root"
    exit 1
fi

start() {
    test $(ulimit -n) -lt 65535 && ulimit -n 65535
    nohup ${DIRECTORY}/apiserver -pidfile ${PIDFILE} >>/dev/null 2>&1 &
    sleep 1
    local pid=$(cat ${PIDFILE})
    echo -n "Starting apiserver(${pid}): "
    if (ps ax 2>/dev/null || ps) | grep "${pid} " >/dev/null 2>&1; then
        echo "OK"
    else
        echo "Failed"
    fi
}

stop() {
    local pid=$(cat ${PIDFILE})
    echo -n "Stopping apiserver(${pid}): "
    if kill ${pid}; then
        echo "OK"
    else
        echo "Failed"
    fi
}

restart() {
    stop
    start
}

reload() {
    kill -HUP $(cat ${PIDFILE})
}

autostart() {
    if ! command -v crontab >/dev/null ; then
        echo "ERROR: please install cron"
    fi
    (crontab -l | grep -v 'apiserver.sh'; echo "*/1 * * * * pgrep apiserver >/dev/null || $(pwd)/apiserver.sh start") | crontab -
}

usage() {
    echo "Usage: [sudo] $(basename "$0") {start|stop|reload|restart|autostart}" >&2
    exit 1
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    reload)
        reload
        ;;
    autostart)
        autostart
        ;;
    *)
        usage
        ;;
esac

exit $?

