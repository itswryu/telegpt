#!/bin/bash

# 테스트를 위한 Secret 생성
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: telegpt-secrets
type: Opaque
stringData:
  TELEGRAM_BOT_TOKEN: "test-token"
  OPENAI_API_KEY: "test-key"
  OPENAI_MODEL: "gpt-4.1-nano"
  ALLOWED_CHAT_IDS: "123456789,987654321"
EOF

# ConfigMap과 Deployment 적용
kubectl apply -f deployment.yaml

# Pod 상태 확인
echo "Pod 상태를 확인합니다..."
sleep 5
kubectl get pods

# 로그 확인
echo "Pod 로그를 확인합니다..."
POD_NAME=$(kubectl get pods -l app=telegpt -o jsonpath="{.items[0].metadata.name}")
kubectl logs $POD_NAME
