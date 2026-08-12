#!/bin/bash
# watchdog.sh - 模块化、纯参数化虚拟机监控与自动恢复脚本
set -o pipefail

# ================= 全局配置 =================
# 格式: "SOURCE_IP,TARGET_IP,SSH_HOST,VM_NAME"
CONFIGS=(
    "183.232.144.131,120.232.56.109,120.232.56.99,vm3"
    # 在这里添加更多组，例如:
    "183.232.144.131,120.232.56.100,120.232.56.97,vm4"
)

PING_COUNT=3
PING_INTERVAL=5
MAX_FAILURES=3
PID_FILE="/home/bjyipeng/pcdn/watchdog.pid"

SSH_OPTS=(
    -o ConnectTimeout=5 -o ServerAliveInterval=5 -o ServerAliveCountMax=2
    -o StrictHostKeyChecking=no -o BatchMode=yes -o LogLevel=ERROR
)

# ================= 1. 日志模块 =================
# 【核心修复】日志必须输出到 stderr (>&2)，防止被 $() 捕获污染状态变量
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >&2
}

# ================= 2. SSH 执行模块 =================
rvirsh() {
    local ssh_ip=$1 ssh_host=$2 vm_name=$3; shift 3
    local out rc
    out=$(ssh -b "${ssh_ip}" "${SSH_OPTS[@]}" root@"${ssh_host}" "$@" 2>&1)
    rc=$?
    if [[ ${rc} -ne 0 ]]; then
        log "[$vm_name] SSH ERR (${rc}): ${out}"
        return ${rc}
    fi
    echo "${out}"
}

# ================= 3. 状态检查模块 =================
wait_vm_state() {
    local ssh_ip=$1 ssh_host=$2 vm_name=$3 timeout=$4 mode=$5
    local elapsed=0 out

    while [[ ${elapsed} -lt ${timeout} ]]; do
        out=$(rvirsh "${ssh_ip}" "${ssh_host}" "${vm_name}" virsh list --name 2>/dev/null) || {
            log "[$vm_name] WARN: SSH unreachable, retrying..."
            sleep 2; ((elapsed+=2)); continue
        }
        case "${mode}" in
            check_gone)
                echo "${out}" | grep -qx "${vm_name}" || { log "[$vm_name] state: gone (${elapsed}s)"; return 0; } ;;
            check_running)
                rvirsh "${ssh_ip}" "${ssh_host}" "${vm_name}" virsh domstate "${vm_name}" 2>/dev/null | grep -qx "running" && { log "[$vm_name] state: running (${elapsed}s)"; return 0; } ;;
        esac
        sleep 2; ((elapsed+=2))
    done
    log "[$vm_name] ERROR: ${mode} timeout after ${timeout}s"
    return 1
}

# ================= 4. 恢复模块 =================
recover_vm() {
    local ssh_ip=$1 ssh_host=$2 vm_name=$3
    log "[$vm_name] === RECOVERY START ==="

    log "[$vm_name] virsh destroy ${vm_name}"
    rvirsh "${ssh_ip}" "${ssh_host}" "${vm_name}" virsh destroy "${vm_name}" 2>/dev/null || \
        log "[$vm_name] WARN: destroy failed (may already be down)"
    sleep 3

    if ! wait_vm_state "${ssh_ip}" "${ssh_host}" "${vm_name}" 30 check_gone; then
        log "[$vm_name] ERROR: destroy timeout"; return 1
    fi

    log "[$vm_name] virsh start ${vm_name}"
    rvirsh "${ssh_ip}" "${ssh_host}" "${vm_name}" virsh start "${vm_name}" || { 
        log "[$vm_name] ERROR: start failed!"; return 1; 
    }
    sleep 3

    if ! wait_vm_state "${ssh_ip}" "${ssh_host}" "${vm_name}" 120 check_running; then
        log "[$vm_name] ERROR: start timeout"; return 1
    fi

    log "[$vm_name] === RECOVERY SUCCESS ==="
    sleep 15
}

# ================= 5. 核心调度模块 =================
check_vm() {
    local src_ip=$1 tgt_ip=$2 ssh_host=$3 vm_name=$4 fail_count=$5
    local lost new_fail_count ping_output

    # 1. 【绝对优先】先打印日志

    ping_output=$(ping -I "${src_ip}" -c "${PING_COUNT}" -W 2 "${tgt_ip}" 2>&1)
    
    # 2. 提取丢包率数字
    lost=$(echo "${ping_output}" | grep -oE '[0-9]+% packet loss' | grep -oE '[0-9]+')
    lost=${lost:-100}

    # 6. 评估结果 (注意：只有这里的 echo 会被 $() 捕获作为返回值)
    if [[ "${lost}" -eq 100 ]]; then
        new_fail_count=$((fail_count + 1))
        log "[$vm_name] PING FAILED (${lost}% loss) [${new_fail_count}/${MAX_FAILURES}]"
        
        if [[ ${new_fail_count} -ge ${MAX_FAILURES} ]]; then
            log "[$vm_name] === ${MAX_FAILURES}x 100% loss, recover ==="
            recover_vm "${src_ip}" "${ssh_host}" "${vm_name}" || log "[$vm_name] RECOVERY FAILED"
            echo "0"  # 返回重置后的计数
        else
            echo "${new_fail_count}"  # 返回新的失败计数
        fi
    else
        log "[$vm_name] ping -I ${src_ip} ${tgt_ip} ${lost}% loss"
    fi
}

# ================= 6. 主程序入口 =================
# PID 锁机制
if [[ -f "${PID_FILE}" ]]; then
    old_pid=$(cat "${PID_FILE}")
    if kill -0 "${old_pid}" 2>/dev/null; then
        echo "ERROR: watchdog already running (pid=${old_pid})" >&2
        exit 1
    fi
fi
echo $$ > "${PID_FILE}"
trap 'rm -f "${PID_FILE}"; log "[SYSTEM] watchdog stopped"' EXIT
trap 'log "[SYSTEM] SIGINT/SIGTERM, shutting down..."; exit 0' INT TERM

# 初始化状态数组
FAILURES=()
for ((i=0; i<${#CONFIGS[@]}; i++)); do FAILURES[$i]=0; done

log "[SYSTEM] started (pid=$$) | monitoring ${#CONFIGS[@]} VMs"

# 主循环
while true; do
    for ((i=0; i<${#CONFIGS[@]}; i++)); do
        IFS=',' read -r SRC_IP TGT_IP SSH_HOST VM_NAME <<< "${CONFIGS[$i]}"
        
        # 调用函数并接收返回的失败计数
        FAILURES[$i]=$(check_vm "${SRC_IP}" "${TGT_IP}" "${SSH_HOST}" "${VM_NAME}" "${FAILURES[$i]}")
    done
    sleep "${PING_INTERVAL}"
done
