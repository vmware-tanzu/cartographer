#!/usr/bin/env bash

set -euo pipefail

start=1
ending=50

if [ "$#" -ne 2 ] && [ "$#" -ne 0 ]; then
    echo "Illegal number of parameters"
    exit 1
fi

if [ "$#" -eq 2 ] && [ "$1" -gt "0" ] && [ "$2" -gt "$1" ] ; then
  start=$1
  ending=$2
fi

echo Creating workload-$start to workload-$ending ...

for a in $(seq $start $ending) ; do
  cat > /tmp/new-workload.yaml <<EOF
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  name: workload-$a
  labels:
    foo: bar
spec: {}
EOF
  kubectl apply -f /tmp/new-workload.yaml
done
