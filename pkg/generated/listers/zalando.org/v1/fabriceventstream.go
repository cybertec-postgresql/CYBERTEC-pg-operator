/*
Copyright 2023 Compose, Zalando SE

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/cybertec-postgresql/CYBERTEC-pg-operator/tree/v0.7.0-rc3/pkg/apis/zalando.org/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// FabricEventStreamLister helps list FabricEventStreams.
// All objects returned here must be treated as read-only.
type FabricEventStreamLister interface {
	// List lists all FabricEventStreams in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.FabricEventStream, err error)
	// FabricEventStreams returns an object that can list and get FabricEventStreams.
	FabricEventStreams(namespace string) FabricEventStreamNamespaceLister
	FabricEventStreamListerExpansion
}

// fabricEventStreamLister implements the FabricEventStreamLister interface.
type fabricEventStreamLister struct {
	indexer cache.Indexer
}

// NewFabricEventStreamLister returns a new FabricEventStreamLister.
func NewFabricEventStreamLister(indexer cache.Indexer) FabricEventStreamLister {
	return &fabricEventStreamLister{indexer: indexer}
}

// List lists all FabricEventStreams in the indexer.
func (s *fabricEventStreamLister) List(selector labels.Selector) (ret []*v1.FabricEventStream, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.FabricEventStream))
	})
	return ret, err
}

// FabricEventStreams returns an object that can list and get FabricEventStreams.
func (s *fabricEventStreamLister) FabricEventStreams(namespace string) FabricEventStreamNamespaceLister {
	return fabricEventStreamNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// FabricEventStreamNamespaceLister helps list and get FabricEventStreams.
// All objects returned here must be treated as read-only.
type FabricEventStreamNamespaceLister interface {
	// List lists all FabricEventStreams in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.FabricEventStream, err error)
	// Get retrieves the FabricEventStream from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.FabricEventStream, error)
	FabricEventStreamNamespaceListerExpansion
}

// fabricEventStreamNamespaceLister implements the FabricEventStreamNamespaceLister
// interface.
type fabricEventStreamNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all FabricEventStreams in the indexer for a given namespace.
func (s fabricEventStreamNamespaceLister) List(selector labels.Selector) (ret []*v1.FabricEventStream, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.FabricEventStream))
	})
	return ret, err
}

// Get retrieves the FabricEventStream from the indexer for a given namespace and name.
func (s fabricEventStreamNamespaceLister) Get(name string) (*v1.FabricEventStream, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("fabriceventstream"), name)
	}
	return obj.(*v1.FabricEventStream), nil
}
