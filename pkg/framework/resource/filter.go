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

// Filter is used to filter resources and subresources within API resource lists.
type Filter interface {
	resource(*Resource) bool
	subResource(*SubResource) bool
}

// NewEmptyFilter returns a new emptyFilter.
func NewEmptyFilter() Filter {
	return &emptyFilter{}
}

// NewAndFilter returns a new andFilter.
func NewAndFilter() Filter {
	return &andFilter{}
}

// NewSkipSubresourceFilter returns a new skipSubresourceFilter.
func NewSkipSubresourceFilter() Filter {
	return &skipSubresourceFilter{}
}

// NewOrFilter returns a new orFilter.
func NewOrFilter() Filter {
	return &orFilter{}
}

// NewFieldFilter returns a new fieldFilter.
func NewFieldFilter(path []string) *fieldFilter {
	return &fieldFilter{emptyFilter{}, path}
}

// NewPrefixStrategy returns a new prefixStrategy.
func NewPrefixStrategy() prefixStrategy {
	return prefixStrategy{}
}

type emptyFilter struct {
}

func (*emptyFilter) resource(*Resource) bool {
	return true
}

func (*emptyFilter) subResource(*SubResource) bool {
	return true
}

type skipSubresourceFilter struct {
	emptyFilter
}

func (*skipSubresourceFilter) subResource(sr *SubResource) bool {
	return !strings.HasSuffix(sr.Resource.Name, "/status")
}

type andFilter struct {
	Filters []Filter
}

func (a *andFilter) resource(r *Resource) bool {
	for _, f := range a.Filters {
		if !f.resource(r) {
			return false
		}
	}
	return true
}

func (a *andFilter) subResource(sr *SubResource) bool {
	for _, f := range a.Filters {
		if !f.subResource(sr) {
			return false
		}
	}
	return true
}

type orFilter struct {
	Filters []Filter
}

func (a *orFilter) resource(r *Resource) bool {
	for _, f := range a.Filters {
		if f.resource(r) {
			return true
		}
	}
	return false
}

func (a *orFilter) subResource(sr *SubResource) bool {
	for _, f := range a.Filters {
		if f.subResource(sr) {
			return true
		}
	}
	return false
}

type fieldFilter struct {
	emptyFilter
	path []string
}

func (f *fieldFilter) resource(r *Resource) bool {
	return r.HasField(f.path)
}

type prefixStrategy struct {
	merge.EmptyStrategy
	prefix string
}

func (fs *prefixStrategy) MergePrimitive(element apply.PrimitiveElement) (apply.Result, error) {
	return apply.Result{MergedResult: fmt.Sprintf("%s%v", fs.prefix, element.GetRemote())}, nil
}
