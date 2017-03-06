#!/bin/bash
supervisorctl stop logbeat
rm -f /etc/supervisord/logbeat-control.conf
supervisorctl update
