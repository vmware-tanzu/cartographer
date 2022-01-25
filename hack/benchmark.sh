#!/usr/bin/env bash
# Copyright 2021 VMware
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -exuo pipefail

# # Cartographer benchmark
#
# Runs multiple passes of the performance test with differing combinations of workloads and runnables, sampling memory and CPU consumption by the controller.
#
# The runs are performed against the cluster configured in the current `kubectl` context. Sampling is done via the metrics API.
#
# Between each pass, the controller is deleted and the benchmark assumes other machinery (e.g. Deployment) will re-create a new one.
#
# The output of each performance test pass is written to a separate log file.
#
# At this time, manual extraction of the values from the logfiles via grep/cut/etc is required. The tail of each logfile
# will contain a line like this:
#
# Memory: 118600Ki min, 118616Ki max
# CPU: 19302840n min, 32752169n max
#
#
# ## Pre-start checklist:
#  - "Similar" cluster setup. Currently we use a GKE cluster with a node pool with 3 e2-standard-4 nodes
#  - Install desired release of cartographer on cluster
#  - Install metrics release on cluster https://github.com/kubernetes-sigs/metrics-server#installation
#  - Identify a host to run the benchmark on. It should be stable, have a fairly robust network connection and be able
#    to run unattended for the better part of the day.
#
# ## Details of each performance test pass
#
# Each pass of the performance test is a single run of the ginkgo test found in in `tests/performance`
#
# This performance test is configured with the following environment variables:
#   CARTO_BENCH_N_SUPPLY_CHAINS   number of supply chain objects to create
#   CARTO_BENCH_N_WORKLOADS       number of workload objects to create (evenly distributes across supplychains)
#   CARTO_BENCH_N_RUNNABLES       number of runnable objects to create
#   CARTO_BENCH_N_RUNNABLE_RUNS   number of times to simulate a change triggering another run on each runnable
#
# Each pass begins by creating the specified number of ClusterSupplyChains, Workloads and Runnables.
#
# For each workload and runnable, the benchmark will poll until they have reconciled successfully, and that they stamped
# out their expected objects.
#
# Further to this, for each object stamped by each runnable (that is, each `run`), the benchmark simulates the stamped
# object completing successfully, and triggers a new run until the number of successful runs is CARTO_BENCH_N_RUNNABLE_RUNS.
#
# Note that because CARTO_BENCH_N_RUNNABLES and CARTO_BENCH_N_RUNNABLE_RUNS multiply, they protract runtime considerably.
#

reset_cartographer_controller() {
  echo Resetting cartographer controller...
  kubectl delete pod -n cartographer-system -l app=cartographer-controller

  while ! kubectl get pod -n cartographer-system -l app=cartographer-controller --field-selector status.phase=Running | grep Running ; do
    echo Waiting for cartographer controller to be ready...
    sleep 1
  done
  while ! kubectl top pod -n cartographer-system ; do
    echo Waiting for cartographer controller to have metrics...
    sleep 1
  done
}

export CARTO_BENCH_N_SUPPLY_CHAINS=5

export CARTO_BENCH_N_RUNNABLE_RUNS=1
for workloads in 2 500 1250 1500; do
  reset_cartographer_controller
  export CARTO_BENCH_N_SUPPLY_CHAINS=5
  export CARTO_BENCH_N_WORKLOADS="$workloads"
  export CARTO_BENCH_N_RUNNABLES="$workloads"
  go run -modfile hack/tools/go.mod github.com/onsi/ginkgo/ginkgo -r benchmark | tee bench_"$workloads"_workloads_w_runnables_1_run.txt
done

export CARTO_BENCH_N_RUNNABLE_RUNS=3
for workloads in 10 50 100 250; do
  reset_cartographer_controller
  export CARTO_BENCH_N_WORKLOADS="$workloads"
  export CARTO_BENCH_N_RUNNABLES="$workloads"
  go run -modfile hack/tools/go.mod github.com/onsi/ginkgo/ginkgo -r benchmark | tee bench_"$workloads"_workloads_w_runnables_3_runs.txt
done

export CARTO_BENCH_N_WORKLOADS=5
for runnables in 5 25 50; do
  for runs in 5 25 50 ; do
      reset_cartographer_controller
      export CARTO_BENCH_N_RUNNABLES="$runnables"
      export CARTO_BENCH_N_RUNNABLE_RUNS="$runs"
      go run -modfile hack/tools/go.mod github.com/onsi/ginkgo/ginkgo -r benchmark | tee bench_"$runnables"_runnables_with_"$runs"_runs.txt
  done
done
