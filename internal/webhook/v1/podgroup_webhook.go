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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	corev1 "github.com/SMALL-head/podGroup/api/v1"
)

// nolint:unused
// log is for logging in this package.
var podgrouplog = logf.Log.WithName("podgroup-resource")

// SetupPodGroupWebhookWithManager registers the webhook for PodGroup in the manager.
func SetupPodGroupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.PodGroup{}).
		WithValidator(&PodGroupCustomValidator{}).
		Complete()
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-core-cic-io-v1-podgroup,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.cic.io,resources=podgroups,verbs=create;update,versions=v1,name=vpodgroup-v1.kb.io,admissionReviewVersions=v1

// PodGroupCustomValidator struct is responsible for validating the PodGroup resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type PodGroupCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &PodGroupCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type PodGroup.
func (v *PodGroupCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	podgroup, ok := obj.(*corev1.PodGroup)
	if !ok {
		return nil, fmt.Errorf("expected a PodGroup object but got %T", obj)
	}
	podgrouplog.Info("Validation for PodGroup upon creation", "name", podgroup.GetName())

	m := make(map[string]bool)

	for _, pod := range podgroup.Spec.PodList {
		if _, exists := m[pod.Metadata.Name]; exists {
			return nil, fmt.Errorf("pod name %s is duplicated in PodGroup %s", pod.Metadata.Name, podgroup.GetName())
		} else {
			m[pod.Metadata.Name] = true
		}
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type PodGroup.
func (v *PodGroupCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	podgroup, ok := newObj.(*corev1.PodGroup)
	if !ok {
		return nil, fmt.Errorf("expected a PodGroup object for the newObj but got %T", newObj)
	}
	podgrouplog.Info("Validation for PodGroup upon update", "name", podgroup.GetName())

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type PodGroup.
func (v *PodGroupCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
