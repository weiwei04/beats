#!/bin/bash
if [ ! -d "/home/qboxserver/logbeat/log" ];then
	mkdir -p /home/qboxserver/logbeat/log
fi

if [ ! -d "/home/qboxserver/logbeat/status" ];then
	mkdir /home/qboxserver/logbeat/status
fi

cp -f /home/qboxserver/logbeat/current/conf/logbeat-control.conf /etc/supervisord/
supervisorctl update
supervisorctl restart logbeat
