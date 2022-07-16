#/bin/bash


 cd /home/bjyipeng
pingcount=100

lists=`cat ./iplist.txt`
function do_ping()
{
    DATE=`date "+%m%d%H"`
    host=$1
    mkdir -p "$DATE"

    start_time=`date +%s`
    echo -e "start ping $ip `date "+%Y-%m-%d %H:%M:%S"`" | tee -a ${DATE}/$1.ping
    
    local retPing=`ping -c $pingcount $1 `
    echo "$retPing" >> ${DATE}/$1.ping
    end_time=`date "+%Y-%m-%d %H:%M:%S"`
    echo "$retPing" | awk -F"[,]" -v hostip="$1" -v OFS='; ' '{for(i=1;i<=NF;i++){if($i~/'packet\ loss'/){print strftime("%Y/%m/%d %H:%M:%S",'$start_time')strftime("-%H:%M:%S",systime()),'$pingcount'" packets",$(i), hostip}}}' >> total.log
    
    echo -e "end ping $ip $end_time\n" >> ${DATE}/$1.ping
}
function thread_list()
{
    while ((1)); do
        threads=`ps -ef | grep -w "ping -c $pingcount" |grep -v grep | grep bjyipeng | wc -l`
        if (( $threads < 9 )); then
            # echo "$ip, threads=$threads "
            do_ping "$ip" &
            break
        fi
    done
}
function start_run()
{
    while ((1)); do
        
        determine_pid
        if (( `date "+%y%m%d"` > 221309 )); then
            exit 1
        fi

        for ip in $lists; do
            if [ -z $ip ]; then
                echo "`date "+%Y-%m-%d %H:%M:%S"` ip is null"
            fi
            thread_list
        done
    done
}
function determine_pid()
{
    if [ -f $0.pid ]; then
        oldpid=`head -n1 $0.pid`
        if [ -f /proc/$oldpid/status ]; then
#            echo "$0 exist, $oldpid"
            exit 0
        fi
    fi

propid=`ps -ef | grep -w $0 | grep -v grep`
newpid=`echo $propid| awk '{print$2}'`
echo "$newpid" > $0.pid
cat $0.pid

}


start_run | tee -a $0.log
