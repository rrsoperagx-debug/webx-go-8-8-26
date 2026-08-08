# Runbook High Latency p99>10ms Go
1. kubectl top pod -l app=webx-go
2. kubectl exec -- df -h /tmp
3. curl /metrics | grep in_flight
4. Mitigation: rollout restart, rollback to v2.2.1, scale cpu
