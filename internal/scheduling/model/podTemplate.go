package model

import (
	podGroupv1 "github.com/SMALL-head/podGroup/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func PodTemplate2PodSpec(template podGroupv1.PodTemplate, podgroupMetadata metav1.ObjectMeta, affinityNode string, ownerRefGVK schema.GroupVersionKind) v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Metadata.Name,
			Namespace: podgroupMetadata.Namespace,
			Labels:    template.Metadata.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&podgroupMetadata, ownerRefGVK),
			},
		},
		Spec: v1.PodSpec{
			Containers: template.Spec.Containers,
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "kubernetes.io/hostname",
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{affinityNode},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return pod
}
