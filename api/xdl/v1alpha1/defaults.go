/*
Copyright 2019 The Alibaba Authors.

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

package v1alpha1

import (
	"strings"

	v1 "github.com/alibaba/kubedl/pkg/job_controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Int32 is a helper routine that allocates a new int32 value
// to store v and returns a pointer to it.
func Int32(v int32) *int32 {
	return &v
}

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func setDefaultXDLJobSpec(spec *XDLJobSpec) {
	// Set default clean pod policy to Running.
	if spec.RunPolicy.CleanPodPolicy == nil {
		running := v1.CleanPodPolicyRunning
		spec.RunPolicy.CleanPodPolicy = &running
	}
	// Set MinFinishWorkerPercentage to default value neither MinFinishWorkerNum nor MinFinishWorkerPercentage
	// are set.
	if spec.MinFinishWorkerNum == nil && spec.MinFinishWorkerPercentage == nil {
		spec.MinFinishWorkerPercentage = Int32(DefaultMinFinishWorkRate)
	}
	// Set MaxFailoverTimes to default value if user not manually set.
	if spec.BackoffLimit == nil {
		spec.BackoffLimit = Int32(DefaultBackoffLimit)
	}
}

// setDefaultPort sets the default ports for xdl container.
func setDefaultPort(spec *corev1.PodSpec) {
	index := 0
	for i, container := range spec.Containers {
		if container.Name == DefaultContainerName {
			index = i
			break
		}
	}

	hasXDLJobPort := false
	for _, port := range spec.Containers[index].Ports {
		if port.Name == DefaultContainerPortName {
			hasXDLJobPort = true
			break
		}
	}
	if !hasXDLJobPort {
		spec.Containers[index].Ports = append(spec.Containers[index].Ports, corev1.ContainerPort{
			Name:          DefaultContainerPortName,
			ContainerPort: DefaultPort,
		})
	}
}

func setDefaultReplicas(spec *v1.ReplicaSpec) {
	if spec.Replicas == nil {
		spec.Replicas = Int32(1)
	}
	if spec.RestartPolicy == "" {
		spec.RestartPolicy = DefaultRestartPolicy
	}
}

func setTypeNamesToCamelCase(xdlJob *XDLJob) {
	setTypeNameToCamelCase(xdlJob, XDLReplicaTypeWorker)
	setTypeNameToCamelCase(xdlJob, XDLReplicaTypePS)
	setTypeNameToCamelCase(xdlJob, XDLReplicaTypeScheduler)
	setTypeNameToCamelCase(xdlJob, XDLReplicaTypeExtendRole)
}

// setTypeNameToCamelCase sets the name of the replica type from any case to correct case.
// E.g. from ps to PS; from WORKER to Worker.
func setTypeNameToCamelCase(xdlJob *XDLJob, typ v1.ReplicaType) {
	for t := range xdlJob.Spec.XDLReplicaSpecs {
		if strings.EqualFold(string(t), string(typ)) && t != typ {
			spec := xdlJob.Spec.XDLReplicaSpecs[t]
			delete(xdlJob.Spec.XDLReplicaSpecs, t)
			xdlJob.Spec.XDLReplicaSpecs[typ] = spec
			return
		}
	}
}

// SetDefaults_XDLJob sets any unspecified values to defaults.
func SetDefaults_XDLJob(xdlJob *XDLJob) {
	setDefaultXDLJobSpec(&xdlJob.Spec)
	setTypeNamesToCamelCase(xdlJob)

	for _, spec := range xdlJob.Spec.XDLReplicaSpecs {
		// Set default replica and restart policy.
		setDefaultReplicas(spec)
		// Set default container port for xdl containers
		setDefaultPort(&spec.Template.Spec)
	}
}
