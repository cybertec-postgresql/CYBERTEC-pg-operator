/*
Copyright 2025 Compose, Zalando SE

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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	time "time"

	zalandoorgv1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/zalando.org/v1"
	versioned "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/generated/informers/externalversions/internalinterfaces"
	v1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/generated/listers/zalando.org/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// FabricEventStreamInformer provides access to a shared informer and lister for
// FabricEventStreams.
type FabricEventStreamInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.FabricEventStreamLister
}

type fabricEventStreamInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewFabricEventStreamInformer constructs a new informer for FabricEventStream type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFabricEventStreamInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredFabricEventStreamInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredFabricEventStreamInformer constructs a new informer for FabricEventStream type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredFabricEventStreamInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ZalandoV1().FabricEventStreams(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ZalandoV1().FabricEventStreams(namespace).Watch(context.TODO(), options)
			},
		},
		&zalandoorgv1.FabricEventStream{},
		resyncPeriod,
		indexers,
	)
}

func (f *fabricEventStreamInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredFabricEventStreamInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *fabricEventStreamInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&zalandoorgv1.FabricEventStream{}, f.defaultInformer)
}

func (f *fabricEventStreamInformer) Lister() v1.FabricEventStreamLister {
	return v1.NewFabricEventStreamLister(f.Informer().GetIndexer())
}
