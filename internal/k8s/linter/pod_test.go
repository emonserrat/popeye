package linter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestPoCheckStatus(t *testing.T) {
	uu := []struct {
		phase    v1.PodPhase
		issues   int
		severity Level
	}{
		{phase: v1.PodPending, issues: 1, severity: ErrorLevel},
		{phase: v1.PodRunning, issues: 0},
		{phase: v1.PodSucceeded, issues: 0},
		{phase: v1.PodFailed, issues: 1, severity: ErrorLevel},
		{phase: v1.PodUnknown, issues: 1, severity: ErrorLevel},
	}

	for _, u := range uu {
		po := v1.Pod{
			Status: v1.PodStatus{
				Phase: u.phase,
			},
		}
		l := NewPod()
		l.checkStatus(po.Status)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestPoCheckContainerStatus(t *testing.T) {
	uu := []struct {
		state    v1.ContainerState
		ready    bool
		issues   int
		severity Level
	}{
		{ready: true, state: v1.ContainerState{Running: new(v1.ContainerStateRunning)}, issues: 0},
		{ready: false, state: v1.ContainerState{Running: new(v1.ContainerStateRunning)}, issues: 1, severity: ErrorLevel},
		{ready: false, state: v1.ContainerState{Waiting: new(v1.ContainerStateWaiting)}, issues: 1, severity: WarnLevel},
		{ready: false, state: v1.ContainerState{Terminated: new(v1.ContainerStateTerminated)}, issues: 1, severity: WarnLevel},
	}

	for _, u := range uu {
		po := v1.Pod{
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						State: u.state,
						Ready: u.ready,
					},
				},
			},
		}

		l := NewPod()
		l.checkContainerStatus(po.Status.ContainerStatuses, false)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestPoCheckContainers(t *testing.T) {
	uu := []struct {
		request, limit      bool
		liveness, readiness bool
		issues              int
		severity            Level
	}{
		{issues: 3, severity: InfoLevel},
		{readiness: true, issues: 2, severity: InfoLevel},
		{liveness: true, issues: 2, severity: InfoLevel},
		{liveness: true, readiness: true, issues: 1, severity: InfoLevel},
		{limit: true, readiness: false, issues: 2, severity: InfoLevel},
		{limit: true, readiness: true, issues: 1, severity: InfoLevel},
		{limit: true, liveness: true, issues: 1, severity: InfoLevel},
		{limit: true, liveness: true, readiness: true, issues: 0},
		{request: true, issues: 2, severity: InfoLevel},
		{request: true, readiness: true, issues: 1, severity: InfoLevel},
		{request: true, liveness: true, issues: 1, severity: InfoLevel},
		{request: true, liveness: true, readiness: true, issues: 0},
		{request: true, limit: true, issues: 2, severity: InfoLevel},
		{request: true, limit: true, readiness: true, issues: 1, severity: InfoLevel},
		{request: true, limit: true, liveness: true, issues: 1, severity: InfoLevel},
		{request: true, limit: true, liveness: true, readiness: true, issues: 0},
	}

	for _, u := range uu {
		po := v1.Pod{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Name: "c1"},
				},
			},
		}
		if u.request {
			po.Spec.Containers[0].Resources = v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.Quantity{},
				},
			}
		}
		if u.limit {
			po.Spec.Containers[0].Resources = v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU: resource.Quantity{},
				},
			}
		}
		if u.liveness {
			po.Spec.Containers[0].LivenessProbe = &v1.Probe{}
		}
		if u.readiness {
			po.Spec.Containers[0].ReadinessProbe = &v1.Probe{}
		}

		l := NewPod()
		l.checkContainers(po.Spec.Containers)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestPoCheckProbes(t *testing.T) {
	uu := []struct {
		liveness, readiness bool
		issues              int
		severity            Level
	}{
		{issues: 2, severity: InfoLevel},
		{liveness: true, readiness: true, issues: 0},
		{liveness: true, issues: 1, severity: InfoLevel},
		{readiness: true, issues: 1, severity: InfoLevel},
	}

	for _, u := range uu {
		po := v1.Pod{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Name: "c1"},
				},
			},
		}
		if u.liveness {
			po.Spec.Containers[0].LivenessProbe = &v1.Probe{}
		}
		if u.readiness {
			po.Spec.Containers[0].ReadinessProbe = &v1.Probe{}
		}

		l := NewPod()
		l.checkProbes(po.Spec.Containers)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}