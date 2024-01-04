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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	acidzalandov1 "github.com/cybertec-postgresql/CYBERTEC-pg-operator/tree/v0.7.0-rc5/pkg/apis/cpo.opensource.cybertec.at/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePostgresTeams implements PostgresTeamInterface
type FakePostgresTeams struct {
	Fake *FakeAcidV1
	ns   string
}

var postgresteamsResource = schema.GroupVersionResource{Group: "cpo.opensource.cybertec.at", Version: "v1", Resource: "postgresteams"}

var postgresteamsKind = schema.GroupVersionKind{Group: "cpo.opensource.cybertec.at", Version: "v1", Kind: "PostgresTeam"}

// Get takes name of the postgresTeam, and returns the corresponding postgresTeam object, and an error if there is any.
func (c *FakePostgresTeams) Get(ctx context.Context, name string, options v1.GetOptions) (result *acidzalandov1.PostgresTeam, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(postgresteamsResource, c.ns, name), &acidzalandov1.PostgresTeam{})

	if obj == nil {
		return nil, err
	}
	return obj.(*acidzalandov1.PostgresTeam), err
}

// List takes label and field selectors, and returns the list of PostgresTeams that match those selectors.
func (c *FakePostgresTeams) List(ctx context.Context, opts v1.ListOptions) (result *acidzalandov1.PostgresTeamList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(postgresteamsResource, postgresteamsKind, c.ns, opts), &acidzalandov1.PostgresTeamList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &acidzalandov1.PostgresTeamList{ListMeta: obj.(*acidzalandov1.PostgresTeamList).ListMeta}
	for _, item := range obj.(*acidzalandov1.PostgresTeamList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested postgresTeams.
func (c *FakePostgresTeams) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(postgresteamsResource, c.ns, opts))

}

// Create takes the representation of a postgresTeam and creates it.  Returns the server's representation of the postgresTeam, and an error, if there is any.
func (c *FakePostgresTeams) Create(ctx context.Context, postgresTeam *acidzalandov1.PostgresTeam, opts v1.CreateOptions) (result *acidzalandov1.PostgresTeam, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(postgresteamsResource, c.ns, postgresTeam), &acidzalandov1.PostgresTeam{})

	if obj == nil {
		return nil, err
	}
	return obj.(*acidzalandov1.PostgresTeam), err
}

// Update takes the representation of a postgresTeam and updates it. Returns the server's representation of the postgresTeam, and an error, if there is any.
func (c *FakePostgresTeams) Update(ctx context.Context, postgresTeam *acidzalandov1.PostgresTeam, opts v1.UpdateOptions) (result *acidzalandov1.PostgresTeam, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(postgresteamsResource, c.ns, postgresTeam), &acidzalandov1.PostgresTeam{})

	if obj == nil {
		return nil, err
	}
	return obj.(*acidzalandov1.PostgresTeam), err
}

// Delete takes name of the postgresTeam and deletes it. Returns an error if one occurs.
func (c *FakePostgresTeams) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(postgresteamsResource, c.ns, name, opts), &acidzalandov1.PostgresTeam{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePostgresTeams) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(postgresteamsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &acidzalandov1.PostgresTeamList{})
	return err
}

// Patch applies the patch and returns the patched postgresTeam.
func (c *FakePostgresTeams) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *acidzalandov1.PostgresTeam, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(postgresteamsResource, c.ns, name, pt, data, subresources...), &acidzalandov1.PostgresTeam{})

	if obj == nil {
		return nil, err
	}
	return obj.(*acidzalandov1.PostgresTeam), err
}
