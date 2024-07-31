/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"os/exec"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	labelv1 "github.com/heimdull/node-cpu-label-operator/api/v1"
)

type NodeCPULabelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop
func (r *NodeCPULabelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	nodes := &corev1.NodeList{}
	if err := r.List(ctx, nodes); err != nil {
		return ctrl.Result{}, err
	}

	for _, node := range nodes.Items {
		cpuType, err := getNodeCPUType(node.Name)
		if err != nil {
			return ctrl.Result{}, err
		}

		label := classifyCPU(cpuType)
		if node.Labels["cpu-speed"] != label {
			node.Labels["cpu-speed"] = label
			if err := r.Update(ctx, &node); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func getNodeCPUType(nodeName string) (string, error) {
	out, err := exec.Command("ssh", nodeName, "lscpu | grep 'Model name:' | awk -F: '{print $2}'").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func classifyCPU(cpuType string) string {
	if strings.Contains(cpuType, "Intel") {
		if strings.Contains(cpuType, "Xeon") {
			return "fast"
		} else if strings.Contains(cpuType, "Core") {
			return "medium"
		} else {
			return "slow"
		}
	}
	return "unknown"
}

func (r *NodeCPULabelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&labelv1.NodeCPULabel{}).
		Complete(r)
}
