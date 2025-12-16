#!/usr/bin/env bash
set -euo pipefail

# Switch autoscaling configs between HPA and KEDA without double-scaling.
#
# Usage:
#   ./switch-autoscaler.sh hpa-default   [--ns load-test] [--root .] [--dry-run]
#   ./switch-autoscaler.sh hpa-custom    [--ns load-test] [--root .] [--dry-run]
#   ./switch-autoscaler.sh keda-custom   [--ns load-test] [--root .] [--dry-run]
#   ./switch-autoscaler.sh keda-cron     [--ns load-test] [--root .] [--dry-run]
#   ./switch-autoscaler.sh off           [--ns load-test] [--dry-run]
#   ./switch-autoscaler.sh status        [--ns load-test]
#
# Directory layout:
#   ./hpa/default_hpa.yaml
#   ./hpa/custom_hpa.yaml
#   ./keda/custom_keda.yaml
#   ./keda/cron_keda.yaml

MODE="${1:-}"
shift || true

NS="load-test"
ROOT="."
DRY_RUN=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --ns) NS="$2"; shift 2;;
    --root) ROOT="$2"; shift 2;;
    --dry-run) DRY_RUN=1; shift;;
    -h|--help)
      sed -n '1,120p' "$0"
      exit 0
      ;;
    *)
      echo "Unknown arg: $1" >&2
      exit 2
      ;;
  esac
done

run() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "+ $*"
  else
    eval "$@"
  fi
}

require_file() {
  local f="$1"
  if [[ ! -f "$f" ]]; then
    echo "Missing file: $f" >&2
    exit 1
  fi
}

delete_hpa() {
  run "kubectl delete hpa -n \"$NS\" --all --ignore-not-found"
}

delete_keda() {
  run "kubectl delete scaledobject -n \"$NS\" --all --ignore-not-found"
  run "kubectl delete triggerauthentication -n \"$NS\" --all --ignore-not-found"
  run "kubectl delete scaledjob -n \"$NS\" --all --ignore-not-found"
}

apply_yaml() {
  local f="$1"
  require_file "$f"
  run "kubectl apply -n \"$NS\" -f \"$f\""
}

apply_hpa_default() {
  local f="$ROOT/hpa/default_hpa.yaml"
  delete_keda
  apply_yaml "$f"
}

apply_hpa_custom() {
  local f="$ROOT/hpa/custom_hpa.yaml"
  delete_keda
  apply_yaml "$f"
}

apply_keda_custom() {
  local f="$ROOT/keda/custom_keda.yaml"
  delete_hpa
  apply_yaml "$f"
}

apply_keda_cron() {
  local f="$ROOT/keda/cron_keda.yaml"
  delete_hpa
  apply_yaml "$f"
}

status() {
  echo "== Namespace: $NS =="
  run "kubectl get deploy -n \"$NS\" load-target -o wide 2>/dev/null || true"
  echo
  echo "== HPA =="
  run "kubectl get hpa -n \"$NS\" 2>/dev/null || true"
  echo
  echo "== KEDA ScaledObject =="
  run "kubectl get scaledobject -n \"$NS\" 2>/dev/null || true"
}

case "$MODE" in
  hpa-default)
    apply_hpa_default
    status
    ;;
  hpa-custom)
    apply_hpa_custom
    status
    ;;
  keda-custom)
    apply_keda_custom
    status
    ;;
  keda-cron)
    apply_keda_cron
    status
    ;;
  off)
    delete_hpa
    delete_keda
    status
    ;;
  status)
    status
    ;;
  *)
    echo "Usage: $0 {hpa-default|hpa-custom|keda-custom|keda-cron|off|status} [--ns load-test] [--root .] [--dry-run]" >&2
    exit 2
    ;;
esac
