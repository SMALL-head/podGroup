/*
Copyright 2025.

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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ScheduledPhase  = "Scheduled"
	SchedulingPhase = "Scheduling"
	FailedPhase     = "Failed"
	DeletedPhase    = "Deleted"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodGroupSpec defines the desired state of PodGroup
type PodGroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	PodList      []PodTemplate `json:"podList,omitempty"`
	Dependencies []Dependency  `json:"dependencies,omitempty"`
	// +optional
	NodeNum int `json:"nodeNum,omitempty"`
}

// PodTemplate 由于kubernetes禁止使用v1.Pod中的Metadata嵌套，因此这里我���自行定义
type PodTemplate struct {
	Metadata PodMetadata `json:"metadata,omitempty"`
	Spec     v1.PodSpec  `json:"spec,omitempty"`
}

type PodMetadata struct {
	Name   string            `json:"name,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

type Dependency struct {
	P1 string `json:"p1,omitempty"`
	P2 string `json:"p2,omitempty"`
}

// PodGroupStatus defines the observed state of PodGroup.
type PodGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase 表示 PodGroup 的调度状态
	// 可选值: "Scheduling", "Scheduled", "Failed"
	// +kubebuilder:validation:Enum=Scheduling;Scheduled;Failed
	Phase string `json:"phase,omitempty"`
	// +optional
	ScheduleResult []PodNodeBinding `json:"scheduleResult,omitempty"`
}

type PodNodeBinding struct {
	PodUID   string `json:"podUID,omitempty"`
	PodName  string `json:"podName,omitempty"`
	NodeName string `json:"nodeName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:webhook:path=/validate-core-cic-io-v1-podgroup,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.cic.io,resources=podgroups,verbs=create;update,versions=v1,name=vpodgroup.kb.io,admissionReviewVersions=v1

// PodGroup is the Schema for the podgroups API
type PodGroup struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of PodGroup
	// +required
	Spec PodGroupSpec `json:"spec"`

	// status defines the observed state of PodGroup
	// +optional
	Status PodGroupStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// PodGroupList contains a list of PodGroup
type PodGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodGroup{}, &PodGroupList{})
}
