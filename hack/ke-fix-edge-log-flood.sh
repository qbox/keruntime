#!/usr/bin/env bash
#
# ke-fix-edge-log-flood.sh - 缓解边缘节点 edgecore 日志洪水导致 journald 丢日志的问题。
#
# 背景:
#   现网边缘节点上 edgecore 每 3 分钟会产生十几万到二十几万条日志, 远超 journald 的
#   默认限流 (RateLimitInterval=30s / RateLimitBurst=1000), 表现为 /var/log/messages
#   里持续出现:
#       journal: Suppressed 236781 messages from /system.slice/edgecore.service
#   被 journald 丢掉的日志 rsyslog 也拿不到, 因此 /home/qboxserver/keruntime/_package/
#   run/edgecore.log 本身就是残缺的 —— 这会让任何基于日志的排查失真。
#
#   根治手段是 edgecore 侧少打日志 (见 edge/pkg/metamanager/process.go 的
#   isEdgedInterestedResource 与 edge/pkg/edged/edged.go 的 isNotEdgedResource),
#   但那需要重新构建并重启 edgecore。本脚本只做节点侧的止血: 放开 journald 对
#   edgecore 的限流, 保证日志不再被静默丢弃。
#
# 注意:
#   * 节点是 CentOS 7 / systemd 219, 不支持 unit 级别的 LogRateLimitIntervalSec,
#     因此只能调整全局 /etc/systemd/journald.conf。
#   * 本脚本只重启 systemd-journald, 不会重启 edgecore。
#     !! 不要在当前版本上重启 edgecore !!  该版本存在 "edgecore 重启导致容器被
#     terminated" 的已知缺陷 (upstream 1cf7de0cd 已修, 本仓库尚未合入)。
#
# 用法:
#   ke-fix-edge-log-flood.sh check     # 只读检查, 默认值
#   ke-fix-edge-log-flood.sh apply     # 备份并写入配置, 重启 systemd-journald
#   ke-fix-edge-log-flood.sh rollback  # 回滚到最近一次备份
#
set -euo pipefail

JOURNALD_CONF=/etc/systemd/journald.conf
BACKUP_SUFFIX=".ke-bak"
# 现网实测 edgecore 峰值约 23.7 万条 / 3 分钟 (约 4 万条 / 30s), 这里给 5 倍余量。
# 不用 0 (完全关闭限流) 是为了在极端情况下仍保留保护。
DESIRED_INTERVAL="30s"
DESIRED_BURST="200000"

log() { echo "[$(date '+%F %T')] $*"; }

check() {
  log "=== journald 限流当前生效值 ==="
  grep -vE '^\s*#|^\s*$' "$JOURNALD_CONF" || echo "(全部为注释, 使用默认值 30s/1000)"

  log "=== 最近的丢日志记录 ==="
  grep -a 'Suppressed.*edgecore' /var/log/messages 2>/dev/null | tail -5 \
    || echo "(/var/log/messages 中暂无记录)"

  local n
  n=$(grep -ac 'Suppressed.*edgecore' /var/log/messages 2>/dev/null || echo 0)
  log "本轮 messages 中 edgecore 被限流次数: $n"

  log "=== edgecore 日志量 ==="
  ls -l /home/qboxserver/keruntime/_package/run/edgecore.log 2>/dev/null \
    || echo "(未找到 edgecore.log)"

  log "=== rsyslog 分流规则 ==="
  cat /etc/rsyslog.d/edgecore.conf 2>/dev/null || echo "(未配置)"
}

apply() {
  if [ ! -w "$JOURNALD_CONF" ]; then
    log "无权限写入 $JOURNALD_CONF, 请用 root 执行"; exit 1
  fi

  local backup="${JOURNALD_CONF}${BACKUP_SUFFIX}"
  if [ ! -f "$backup" ]; then
    cp -a "$JOURNALD_CONF" "$backup"
    log "已备份原配置到 $backup"
  else
    log "备份已存在, 保留: $backup"
  fi

  # 幂等写入: 先删掉可能存在的旧设置, 再追加到 [Journal] 段。
  sed -i '/^\s*RateLimitInterval\s*=/d; /^\s*RateLimitIntervalSec\s*=/d; /^\s*RateLimitBurst\s*=/d' "$JOURNALD_CONF"
  sed -i "/^\[Journal\]/a RateLimitInterval=${DESIRED_INTERVAL}\nRateLimitBurst=${DESIRED_BURST}" "$JOURNALD_CONF"

  log "写入后的配置:"
  grep -vE '^\s*#|^\s*$' "$JOURNALD_CONF"

  systemctl restart systemd-journald
  log "systemd-journald 已重启 (edgecore 未受影响)"

  log "验证: 等待 60s 后执行 check, 确认不再出现新的 Suppressed 记录"
}

rollback() {
  local backup="${JOURNALD_CONF}${BACKUP_SUFFIX}"
  if [ ! -f "$backup" ]; then
    log "未找到备份 $backup, 无法回滚"; exit 1
  fi
  cp -a "$backup" "$JOURNALD_CONF"
  systemctl restart systemd-journald
  log "已回滚并重启 systemd-journald"
}

case "${1:-check}" in
  check)    check ;;
  apply)    apply ;;
  rollback) rollback ;;
  *) echo "用法: $0 {check|apply|rollback}" >&2; exit 2 ;;
esac
