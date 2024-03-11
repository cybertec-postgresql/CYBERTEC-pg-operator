/*
Copyright 2024 Compose, Zalando SE

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
	v1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// PostgresqlLister helps list Postgresqls.
// All objects returned here must be treated as read-only.
type PostgresqlLister interface {
	// List lists all Postgresqls in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.Postgresql, err error)
	// Postgresqls returns an object that can list and get Postgresqls.
	Postgresqls(namespace string) PostgresqlNamespaceLister
	PostgresqlListerExpansion
}

// postgresqlLister implements the PostgresqlLister interface.
type postgresqlLister struct {
	indexer cache.Indexer
}

// NewPostgresqlLister returns a new PostgresqlLister.
func NewPostgresqlLister(indexer cache.Indexer) PostgresqlLister {
	return &postgresqlLister{indexer: indexer}
}

// List lists all Postgresqls in the indexer.
func (s *postgresqlLister) List(selector labels.Selector) (ret []*v1.Postgresql, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Postgresql))
	})
	return ret, err
}

// Postgresqls returns an object that can list and get Postgresqls.
func (s *postgresqlLister) Postgresqls(namespace string) PostgresqlNamespaceLister {
	return postgresqlNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PostgresqlNamespaceLister helps list and get Postgresqls.
// All objects returned here must be treated as read-only.
type PostgresqlNamespaceLister interface {
	// List lists all Postgresqls in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.Postgresql, err error)
	// Get retrieves the Postgresql from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.Postgresql, error)
	PostgresqlNamespaceListerExpansion
}

// postgresqlNamespaceLister implements the PostgresqlNamespaceLister
// interface.
type postgresqlNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Postgresqls in the indexer for a given namespace.
func (s postgresqlNamespaceLister) List(selector labels.Selector) (ret []*v1.Postgresql, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Postgresql))
	})
	return ret, err
}

// Get retrieves the Postgresql from the indexer for a given namespace and name.
func (s postgresqlNamespaceLister) Get(name string) (*v1.Postgresql, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("postgresql"), name)
	}
	return obj.(*v1.Postgresql), nil
}
