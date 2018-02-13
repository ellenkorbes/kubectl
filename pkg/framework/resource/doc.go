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

/*
Package resource implements tools for the discovery, indexing, and filtering of resources in an API server.

The starting point should be the creation of a new Parser object, which can then use the Resources() method to discover resources in the API server.

For example:

	p := resource.NewParser()
	r, err := p.Resources()
	if err != nil {
		panic(err)
	}
	for name, versions := range r {
		fmt.Println("\n→", name)
		for _, version := range versions {
			fmt.Printf("→→ %+v\n", version)
		}
	}

should output

	...
	→ services
	→→ &{Resource:{Name:services SingularName: Namespaced:true Group:v1 Version: Kind:Service Verbs:[create delete get list patch proxy update watch] ShortNames:[svc] Categories:[all]} ApiGroupVersion:v1 Schema:0xc420eca410 SubResources:[]}

	→ deployments
	→→ &{Resource:{Name:deployments SingularName: Namespaced:true Group:v1beta1 Version: Kind:Deployment Verbs:[create delete deletecollection get list patch update watch] ShortNames:[deploy] Categories:[all]} ApiGroupVersion:extensions/v1beta1 Schema:0xc420ed7400 SubResources:[]}
	→→ &{Resource:{Name:deployments SingularName: Namespaced:true Group:v1beta1 Version: Kind:Deployment Verbs:[create delete deletecollection get list patch update watch] ShortNames:[deploy] Categories:[all]} ApiGroupVersion:apps/v1beta1 Schema:0xc4202e9400 SubResources:[]}
	...

Filtering can also be applied by implementing the Filter interface.

The following (admittedly ludicrous) example implements a filter that excludes all resources that do not start with the letter "n":

	type letterN struct {
	}

	func (*letterN) Resource(r *resource.Resource) bool {
		return string(r.Resource.Name[0]) == "n"
	}

	func (*letterN) SubResource(*resource.SubResource) bool {
		return true
	}
	func main() {
		p := resource.NewParser()
		r, err := p.Resources()
		if err != nil {
			panic(err)
		}
		r = r.Filter(&letterN{})
		for name, versions := range r {
			fmt.Println("\n→", name)
			for _, version := range versions {
				fmt.Printf("→→ %+v\n", version)
			}
		}
	}

*/
package resource
