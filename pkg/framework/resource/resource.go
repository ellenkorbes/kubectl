/*
Copyright 2018 The Kubernetes Authors.

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

package resource

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	openapi "k8s.io/kube-openapi/pkg/util/proto"
)

// Resource is an API Resource.
type Resource struct {
	Resource        v1.APIResource
	ApiGroupVersion schema.GroupVersion
	openapi.Schema
	SubResources []*SubResource
}

// SubResource is an API subresource.
type SubResource struct {
	Resource        v1.APIResource
	Parent          *Resource
	ApiGroupVersion schema.GroupVersion
	openapi.Schema
}

func (r *Resource) HasField(path []string) bool {
	return hasField(r.Schema, path)
}

func (sr *SubResource) HasField(path []string) bool {
	return hasField(sr.Schema, path)
}

func (r *Resource) Field(path []string, obj interface{}, fn ObjectFieldFn) (interface{}, error) {
	return setField(r.Schema, path, obj, fn)
}

func (sr *SubResource) Field(path []string, obj interface{}, fn ObjectFieldFn) (interface{}, error) {
	return setField(sr.Schema, path, obj, fn)
}

// APIGroupVersionKind creates a GroupVersionKind object based on the GroupVersion and the Kind of the Resource at hand.
func (r *Resource) APIGroupVersionKind() schema.GroupVersionKind {
	return r.ApiGroupVersion.WithKind(r.Resource.Kind)
}

// APIGroupVersionKind creates a GroupVersionKind object based on the GroupVersion and the Kind of the SubResource at hand.
func (sr *SubResource) APIGroupVersionKind() schema.GroupVersionKind {
	return sr.ApiGroupVersion.WithKind(sr.Resource.Kind)
}

// ResourceGroupVersionKind returns a GVK object based on the Resource at hand.
func (r *Resource) ResourceGroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{r.Resource.Group, r.Resource.Version, r.Resource.Kind}
}

// ResourceGroupVersionKind returns a GVK object based on the SubResource at hand.
func (sr *SubResource) ResourceGroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{sr.Resource.Group, sr.Resource.Version, sr.Resource.Kind}
}
