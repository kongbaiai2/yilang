#!/bin/sh
DATE=`date +%Y-%m-%d_%H-%M-%S`

IP=`ifconfig -a|grep inet|grep -v "127.0.0.1\|inet6" |awk '{print $2}'|tr -d "addr:" | head -1`

a="_"

if [ ! "$#" -eq 1 ]
then 
   echo "Usage:fpingrun_test.sh ipfilename"
fi
 
if [ "$#" -eq 1 ]
then
   cd ${0%/*}
   file=$1   
  ./fping_iptest.pl ./$file ./$file$a$IP$a$DATE 
fi
