/*
Copyright The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha

import (
	v1alpha "github.com/EdgeNet-project/edgenet/pkg/apis/apps/v1alpha"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// AuthorityRequestLister helps list AuthorityRequests.
// All objects returned here must be treated as read-only.
type AuthorityRequestLister interface {
	// List lists all AuthorityRequests in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha.AuthorityRequest, err error)
	// Get retrieves the AuthorityRequest from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha.AuthorityRequest, error)
	AuthorityRequestListerExpansion
}

// authorityRequestLister implements the AuthorityRequestLister interface.
type authorityRequestLister struct {
	indexer cache.Indexer
}

// NewAuthorityRequestLister returns a new AuthorityRequestLister.
func NewAuthorityRequestLister(indexer cache.Indexer) AuthorityRequestLister {
	return &authorityRequestLister{indexer: indexer}
}

// List lists all AuthorityRequests in the indexer.
func (s *authorityRequestLister) List(selector labels.Selector) (ret []*v1alpha.AuthorityRequest, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha.AuthorityRequest))
	})
	return ret, err
}

// Get retrieves the AuthorityRequest from the index for a given name.
func (s *authorityRequestLister) Get(name string) (*v1alpha.AuthorityRequest, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha.Resource("authorityrequest"), name)
	}
	return obj.(*v1alpha.AuthorityRequest), nil
}