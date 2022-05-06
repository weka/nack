// Copyright 2020 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1beta2 "github.com/nats-io/nack/pkg/jetstream/apis/jetstream/v1beta2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeStreams implements StreamInterface
type FakeStreams struct {
	Fake *FakeJetstreamV1beta2
	ns   string
}

var streamsResource = schema.GroupVersionResource{Group: "jetstream.nats.io", Version: "v1beta2", Resource: "streams"}

var streamsKind = schema.GroupVersionKind{Group: "jetstream.nats.io", Version: "v1beta2", Kind: "Stream"}

// Get takes name of the stream, and returns the corresponding stream object, and an error if there is any.
func (c *FakeStreams) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta2.Stream, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(streamsResource, c.ns, name), &v1beta2.Stream{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.Stream), err
}

// List takes label and field selectors, and returns the list of Streams that match those selectors.
func (c *FakeStreams) List(ctx context.Context, opts v1.ListOptions) (result *v1beta2.StreamList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(streamsResource, streamsKind, c.ns, opts), &v1beta2.StreamList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta2.StreamList{ListMeta: obj.(*v1beta2.StreamList).ListMeta}
	for _, item := range obj.(*v1beta2.StreamList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested streams.
func (c *FakeStreams) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(streamsResource, c.ns, opts))

}

// Create takes the representation of a stream and creates it.  Returns the server's representation of the stream, and an error, if there is any.
func (c *FakeStreams) Create(ctx context.Context, stream *v1beta2.Stream, opts v1.CreateOptions) (result *v1beta2.Stream, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(streamsResource, c.ns, stream), &v1beta2.Stream{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.Stream), err
}

// Update takes the representation of a stream and updates it. Returns the server's representation of the stream, and an error, if there is any.
func (c *FakeStreams) Update(ctx context.Context, stream *v1beta2.Stream, opts v1.UpdateOptions) (result *v1beta2.Stream, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(streamsResource, c.ns, stream), &v1beta2.Stream{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.Stream), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeStreams) UpdateStatus(ctx context.Context, stream *v1beta2.Stream, opts v1.UpdateOptions) (*v1beta2.Stream, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(streamsResource, "status", c.ns, stream), &v1beta2.Stream{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.Stream), err
}

// Delete takes name of the stream and deletes it. Returns an error if one occurs.
func (c *FakeStreams) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(streamsResource, c.ns, name, opts), &v1beta2.Stream{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeStreams) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(streamsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta2.StreamList{})
	return err
}

// Patch applies the patch and returns the patched stream.
func (c *FakeStreams) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta2.Stream, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(streamsResource, c.ns, name, pt, data, subresources...), &v1beta2.Stream{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.Stream), err
}