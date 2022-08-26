/*
Copyright 2022.

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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BlogAppSpec defines the desired state of BlogApp
type BlogAppSpec struct {
	ServiceType v1.ServiceType     `json:"serviceType,omitempty"`
	Size        *resource.Quantity `json:"size,omitempty"`
}

type BlogAppPhase string

const (
	Error   = "Error"
	Pending = "Pending"
	Running = "Running"
)

// BlogAppStatus defines the observed state of BlogApp
type BlogAppStatus struct {
	Phase BlogAppPhase `json:"phase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase",description="Current state of the Blog application"

// BlogApp is the Schema for the blogapps API
type BlogApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BlogAppSpec   `json:"spec,omitempty"`
	Status BlogAppStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BlogAppList contains a list of BlogApp
type BlogAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BlogApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BlogApp{}, &BlogAppList{})
}
