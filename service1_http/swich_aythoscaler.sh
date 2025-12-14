#!/usr/bin/env bash
set -euo pipefail

NS="service1-lab"
HPA_NAME="service1-hpa"
KEDA_SO_NAME="service1-keda-cpu"
HPA_YAML="hpa/hpa-service1.yaml"
KEDA_YAML="keda/keda-banila.yaml"

usage() {
  cat <<EOF
Usage: $0 [hpa|keda|status]

  hpa    : ネイティブHPAモードに切り替え（KEDA停止）
  keda   : KEDAモードに切り替え（ネイティブHPA停止）
  status : 現在のHPA / ScaledObjectの状態を表示
EOF
  exit 1
}

MODE="${1:-status}"

case "$MODE" in
  hpa)
    echo ">>> Switching to pure HPA mode..."
    # KEDAのScaledObjectを止める
    kubectl delete scaledobject "${KEDA_SO_NAME}" -n "${NS}" --ignore-not-found

    # HPAを適用
    kubectl apply -f "${HPA_YAML}"

    echo
    echo "[HPA status]"
    kubectl get hpa -n "${NS}"
    ;;

  keda)
    echo ">>> Switching to KEDA mode..."
    # ネイティブHPAを止める
    kubectl delete hpa "${HPA_NAME}" -n "${NS}" --ignore-not-found

    # KEDAのScaledObjectを適用
    kubectl apply -f "${KEDA_YAML}"

    echo
    echo "[ScaledObject status]"
    kubectl get scaledobject -n "${NS}"

    echo
    echo "[HPA (KEDA管理) status]"
    kubectl get hpa -n "${NS}"
    ;;

  status)
    echo "[HPA in ${NS}]"
    kubectl get hpa -n "${NS}" || true
    echo
    echo "[ScaledObject in ${NS}]"
    kubectl get scaledobject -n "${NS}" || true
    ;;

  *)
    usage
    ;;
esac
