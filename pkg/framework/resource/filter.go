/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"
	"strings"

	"k8s.io/kubectl/pkg/framework/merge"
	"k8s.io/kubernetes/pkg/kubectl/apply"
)

// Filter is an interface whose methods are used to arbitrarily filter resources and subresources in a Resources object. Filtering criteria can be anything at all, so long as the "accepted" resources return true, and the filtered out resources return false.
type Filter interface {
	Resource(*Resource) bool
	SubResource(*SubResource) bool
}

// NewEmptyFilter returns a new emptyFilter.
func NewEmptyFilter() Filter {
	return &emptyFilter{}
}

// NewAndFilter returns a new andFilter. An andFilter contains the Filters field, a slice of filters which then work as a logical AND.
func NewAndFilter() Filter {
	return &andFilter{}
}

// NewSkipSubresourceFilter returns a new skipSubresourceFilter.
func NewSkipSubresourceFilter() Filter {
	return &skipSubresourceFilter{}
}

// NewOrFilter returns a new orFilter. An orFilter contains the Filters field, a slice of filters which then work as a logical OR.
func NewOrFilter() Filter {
	return &orFilter{}
}

// NewFieldFilter returns a new fieldFilter. It filters based on the fields contained in the slice of string it takes as an argument.
func NewFieldFilter(path []string) Filter {
	return &fieldFilter{emptyFilter{}, path}
}

// NewPrefixStrategy returns a new prefixStrategy.
func NewPrefixStrategy() prefixStrategy {
	return prefixStrategy{}
}

type emptyFilter struct {
}

func (*emptyFilter) Resource(*Resource) bool {
	return true
}

func (*emptyFilter) SubResource(*SubResource) bool {
	return true
}

type skipSubresourceFilter struct {
	emptyFilter
}

func (*skipSubresourceFilter) SubResource(sr *SubResource) bool {
	return !strings.HasSuffix(sr.Resource.Name, "/status")
}

type andFilter struct {
	Filters []Filter
}

func (a *andFilter) Resource(r *Resource) bool {
	for _, f := range a.Filters {
		if !f.Resource(r) {
			return false
		}
	}
	return true
}

func (a *andFilter) SubResource(sr *SubResource) bool {
	for _, f := range a.Filters {
		if !f.SubResource(sr) {
			return false
		}
	}
	return true
}

type orFilter struct {
	Filters []Filter
}

func (a *orFilter) Resource(r *Resource) bool {
	for _, f := range a.Filters {
		if f.Resource(r) {
			return true
		}
	}
	return false
}

func (a *orFilter) SubResource(sr *SubResource) bool {
	for _, f := range a.Filters {
		if f.SubResource(sr) {
			return true
		}
	}
	return false
}

type fieldFilter struct {
	emptyFilter
	path []string
}

func (f *fieldFilter) Resource(r *Resource) bool {
	return r.HasField(f.path)
}

type prefixStrategy struct {
	merge.EmptyStrategy
	prefix string
}

func (fs *prefixStrategy) MergePrimitive(element apply.PrimitiveElement) (apply.Result, error) {
	return apply.Result{MergedResult: fmt.Sprintf("%s%v", fs.prefix, element.GetRemote())}, nil
}
