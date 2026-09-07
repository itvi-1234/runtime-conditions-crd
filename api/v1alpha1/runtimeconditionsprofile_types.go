/*
Copyright 2026 Runtime Conditions Contributors.

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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Workload identifies the single application a profile describes.
type Workload struct {
	// uri is a stable identifier for the workload, e.g. a repo URL.
	// +required
	// +kubebuilder:validation:MinLength=1
	URI string `json:"uri"`

	// version is the workload version this profile was generated for.
	// +optional
	// +kubebuilder:validation:MinLength=1
	Version string `json:"version,omitempty"`
}

// ConditionInterface is the workload-facing interface requirement for a
// Condition. `type` is the only field the core spec reserves; everything
// else here is extension-defined (operations, engine, etc.), so we don't
// type it - we keep it and let the admission webhook validate it against
// whatever extension the profile declared.
type ConditionInterface struct {
	// +required
	// +kubebuilder:validation:MinLength=1
	Type string `json:"type"`
}

// Condition is one external runtime dependency a workload requires.
//
// name/optional/kind/interface.type are the fields the core spec reserves,
// so they're typed and validated here. Everything extension-defined
// (additional interface fields, all of configuration) is left schemaless -
// see docs/design/extension-validation.md for why that split exists and
// what still needs to happen in the admission webhook.
type Condition struct {
	// name is a unique label for this condition within the profile.
	// +optional
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// +optional
	// +kubebuilder:default=false
	Optional bool `json:"optional,omitempty"`

	// kind is the extension-defined integration classification, e.g.
	// "api", "datastore", "cache". Whether a given value is actually
	// backed by a resolved extension is checked in the webhook, not here.
	// +required
	// +kubebuilder:validation:MinLength=1
	Kind string `json:"kind"`

	// +required
	// +kubebuilder:pruning:PreserveUnknownFields
	Interface ConditionInterface `json:"interface"`

	// configuration describes how resolved values bind into the
	// workload's runtime config (env vars, etc). Fully extension-defined.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Configuration *apiextensionsv1.JSON `json:"configuration,omitempty"`
}

// RuntimeConditionsProfileSpec is the profile's desired state.
//
// Only the core, spec-reserved shape is validated statically here -
// workload identity, the extension list, and each condition's reserved
// fields. What a condition's kind/interface actually means is defined by
// whichever extensions the profile declares, so that part is intentionally
// left schemaless and pushed to a validating admission webhook (see
// docs/design/extension-validation.md).
//
// metadata.name/labels from the spec map onto this CR's own ObjectMeta, so
// they're not duplicated here.
type RuntimeConditionsProfileSpec struct {
	// +required
	Workload Workload `json:"workload"`

	// extensions is the list of extension identifiers (absolute URIs) this
	// profile depends on. Can be empty, can't have duplicates.
	// +required
	// +kubebuilder:validation:XValidation:rule="self.all(e, e.matches('^[a-zA-Z][a-zA-Z0-9+.-]*://.+'))",message="each extension must be an absolute URI with a scheme"
	// +kubebuilder:validation:XValidation:rule="self.all(x, self.exists_one(y, y == x))",message="extensions must not contain duplicate identifiers"
	Extensions []string `json:"extensions"`

	// conditions is the workload's external runtime dependencies. Can be
	// empty. Vocabulary validation (is this kind/interface.type actually
	// defined by an extension, condition name uniqueness, JSON Schema
	// checks) happens in the admission webhook, not in this schema.
	// +required
	Conditions []Condition `json:"conditions"`
}

// RuntimeConditionsProfileStatus defines the observed state of RuntimeConditionsProfile.
type RuntimeConditionsProfileStatus struct {
	// conditions represent the current state of the RuntimeConditionsProfile resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rcprofile;rcprofiles

// RuntimeConditionsProfile is the Schema for the runtimeconditionsprofiles API
type RuntimeConditionsProfile struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of RuntimeConditionsProfile
	// +required
	Spec RuntimeConditionsProfileSpec `json:"spec"`

	// status defines the observed state of RuntimeConditionsProfile
	// +optional
	Status RuntimeConditionsProfileStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// RuntimeConditionsProfileList contains a list of RuntimeConditionsProfile
type RuntimeConditionsProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []RuntimeConditionsProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &RuntimeConditionsProfile{}, &RuntimeConditionsProfileList{})
		return nil
	})
}
