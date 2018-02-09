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
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/kube-openapi/pkg/util/proto"
	"k8s.io/kubectl/pkg/framework/internal/inject"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi"
)

// Parser is an object type that can be used to discover resources from an API server and parse them into indexed data structures.
type Parser struct {
	resources  openapi.Resources
	discovery  discovery.DiscoveryInterface
	rest       rest.Interface
	apiGroup   string
	apiVersion string
}

// NewParser populates the fields of and returns a new Parser.
func NewParser() Parser {
	return Parser{
		inject.FactorySingleton.GetResources(),
		inject.FactorySingleton.GetDiscovery(),
		inject.FactorySingleton.GetRest(),
		inject.FactorySingleton.GetApiGroup(),
		inject.FactorySingleton.GetApiVersion(),
	}
}

// Resources discovers and indexes resources from the API server.
// It returns a map of resource name to resources matching that name ordered by preference as reported by the server.
func (p *Parser) Resources() (Resources, error) {
	gvs, err := p.discovery.ServerResources()
	if err != nil {
		return nil, err
	}
	resources, byGVR := p.indexResources(gvs)
	err = p.attachSubResources(gvs, resources, byGVR)
	return resources, err
}

// subResource returns a resource name, subresource name pair, and true if the resource is a subresource.
// It returns a resource name, an empty string, and false if the resource is not a subresource.
func (*Parser) subResource(resource *v1.APIResource) (string, string, bool) {
	parts := strings.Split(resource.Name, "/")
	if len(parts) > 1 {
		return parts[0], parts[1], true
	}
	return parts[0], "", false
}

// resource returns the resource name and true if it is a resource (not a subresource).
// It returns the resource name and false it if it is a subresource.
func (p *Parser) resource(resource *v1.APIResource) (string, bool) {
	r, _, b := p.subResource(resource)
	return r, !b
}

// copyGroupVersion copies the group and version from source to destination if either is missing on the destination.
// If the source group is empty and the destination group is empty, sets the destination group to "core".
func (*Parser) copyGroupVersion(src *v1.APIResource, dest *v1.APIResource) {
	if len(dest.Group) == 0 {
		dest.Group = src.Group
	}
	if len(dest.Group) == 0 {
		dest.Group = "core"
	}
	if len(dest.Version) == 0 {
		dest.Version = src.Version
	}
}

// splitGroupVersion splits the groupVersion string into its group and version components.
func (*Parser) splitGroupVersion(groupVersion string) (string, string) {
	parts := strings.Split(groupVersion, "/")
	var group, version string
	if len(parts) > 1 {
		group = parts[0]
	}
	if len(parts) > 1 {
		version = parts[1]
	} else if len(parts) > 0 {
		version = parts[0]
	} else {
		version = "v1"
	}
	return group, version
}

// defaultGroupVersion sets the group and version to the API group and version if they're missing.
func (p *Parser) defaultGroupVersion(resource *v1.APIResource, group, version string) {
	if len(resource.Group) == 0 {
		resource.Group = group
	}
	if len(resource.Version) == 0 {
		resource.Group = version
	}
}

// isGroupVersionMatch returns false if either group or version doesn't match with the API.
func (p *Parser) isGroupVersionMatch(group, version string) bool {
	if len(p.apiGroup) > 0 && p.apiGroup != group {
		return false
	}
	if len(p.apiVersion) > 0 && p.apiVersion != version {
		return false
	}
	return true
}

// getOpenAPI retrieves a schema object from the API based on a GroupVersionResource triplet.
func (p *Parser) getOpenAPI(group, version, kind string) (proto.Schema, bool) {
	schema := p.resources.LookupResource(schema.GroupVersionKind{group, version, kind})
	if schema == nil {
		return nil, false
	}
	return schema, true
}

// indexResources indexes into maps the resources from the API resource list. It returns a map indexed by resource name, and a map indexed by GroupVersionResource objects.
func (p *Parser) indexResources(gvs []*v1.APIResourceList) (map[string][]*Resource, map[schema.GroupVersionResource]*Resource) {
	resources := map[string][]*Resource{}
	bygvr := map[schema.GroupVersionResource]*Resource{}
	// Find all resources
	for _, gv := range gvs {
		group, version := p.splitGroupVersion(gv.GroupVersion)
		if !p.isGroupVersionMatch(group, version) {
			continue
		}
		for _, r := range gv.APIResources {
			p.defaultGroupVersion(&r, group, version)
			name, isRes := p.resource(&r)
			if !isRes {
				continue
			}
			newSchema, found := p.getOpenAPI(group, version, r.Kind)
			if !found {
				continue
			}
			newResource := &Resource{
				Resource:        r,
				ApiGroupVersion: schema.GroupVersion{group, version},
				Schema:          newSchema,
			}
			resources[name] = append(resources[name], newResource)
			bygvr[schema.GroupVersionResource{group, version, r.Kind}] = newResource
		}
	}
	return resources, bygvr
}

// attachSubResources grabs the subresources in the gvs argument and attaches them to the parent resources in the maps present in the next two arguments.
func (p *Parser) attachSubResources(
	gvs []*v1.APIResourceList,
	resources map[string][]*Resource,
	bygvr map[schema.GroupVersionResource]*Resource) error {
	// Find all subresources and attach to parents
	for _, gv := range gvs {
		group, version := p.splitGroupVersion(gv.GroupVersion)
		if !p.isGroupVersionMatch(group, version) {
			continue
		}
		for _, r := range gv.APIResources {
			p.defaultGroupVersion(&r, group, version)
			resourceName, _, isSubResource := p.subResource(&r)
			if !isSubResource {
				continue
			}
			newSchema, found := p.getOpenAPI(group, version, r.Kind)
			if !found {
				continue
			}
			// Make sure the Parent resources wasn't filtered out
			gvr := schema.GroupVersionResource{group, version, resourceName}
			if _, found := bygvr[gvr]; !found {
				continue
			}
			parent := bygvr[gvr]
			subRes := &SubResource{
				Resource:        r,
				Parent:          parent,
				ApiGroupVersion: schema.GroupVersion{group, version},
				Schema:          newSchema,
			}
			parent.SubResources = append(parent.SubResources, subRes)
		}
	}
	return nil
}
