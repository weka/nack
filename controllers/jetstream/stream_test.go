package jetstream

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	apis "github.com/nats-io/nack/pkg/jetstream/apis/jetstream/v1"
	clientsetfake "github.com/nats-io/nack/pkg/jetstream/generated/clientset/versioned/fake"
	informers "github.com/nats-io/nack/pkg/jetstream/generated/informers/externalversions"

	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

func TestMain(m *testing.M) {
	// Disable error logs.
	utilruntime.ErrorHandlers = []func(error){
		func(err error) {},
	}

	os.Exit(m.Run())
}

func TestValidateStreamUpdate(t *testing.T) {
	t.Parallel()

	t.Run("no spec changes, no update", func(t *testing.T) {
		s := &apis.Stream{
			Spec: apis.StreamSpec{
				Name:   "foo",
				MaxAge: "1h",
			},
		}

		if err := validateStreamUpdate(s, s); !errors.Is(err, errNothingToUpdate) {
			t.Fatalf("got=%v; want=%v", err, errNothingToUpdate)
		}
	})

	t.Run("spec changed, update ok", func(t *testing.T) {
		prev := &apis.Stream{
			Spec: apis.StreamSpec{
				Name:   "foo",
				MaxAge: "1h",
			},
		}
		next := &apis.Stream{
			Spec: apis.StreamSpec{
				Name:   "foo",
				MaxAge: "10h",
			},
		}

		if err := validateStreamUpdate(prev, next); err != nil {
			t.Fatalf("got=%v; want=nil", err)
		}
	})

	t.Run("stream name changed, update bad", func(t *testing.T) {
		prev := &apis.Stream{
			Spec: apis.StreamSpec{
				Name:   "foo",
				MaxAge: "1h",
			},
		}
		next := &apis.Stream{
			Spec: apis.StreamSpec{
				Name:   "bar",
				MaxAge: "1h",
			},
		}

		if err := validateStreamUpdate(prev, next); err == nil {
			t.Fatal("got=nil; want=err")
		}
	})
}

func TestEnqueueStreamWork(t *testing.T) {
	t.Parallel()

	limiter := workqueue.DefaultControllerRateLimiter()
	q := workqueue.NewNamedRateLimitingQueue(limiter, "StreamsTest")
	defer q.ShutDown()

	s := &apis.Stream{
		ObjectMeta: k8smeta.ObjectMeta{
			Namespace: "default",
			Name:      "my-stream",
		},
	}

	if err := enqueueStreamWork(q, s); err != nil {
		t.Fatal(err)
	}

	if got, want := q.Len(), 1; got != want {
		t.Error("unexpected queue length")
		t.Fatalf("got=%d; want=%d", got, want)
	}

	wantItem := fmt.Sprintf("%s/%s", s.Namespace, s.Name)
	gotItem, _ := q.Get()
	if gotItem != wantItem {
		t.Error("unexpected queue item")
		t.Fatalf("got=%s; want=%s", gotItem, wantItem)
	}
}

func TestProcessStream(t *testing.T) {
	t.Parallel()

	t.Run("delete stream", func(t *testing.T) {
		jc := clientsetfake.NewSimpleClientset()
		informerFactory := informers.NewSharedInformerFactory(jc, 0)
		informer := informerFactory.Jetstream().V1().Streams()

		ts := k8smeta.Unix(1600216923, 0)
		name := "my-stream"

		err := informer.Informer().GetStore().Add(
			&apis.Stream{
				ObjectMeta: k8smeta.ObjectMeta{
					Namespace:         "default",
					Name:              name,
					DeletionTimestamp: &ts,
					Finalizers:        []string{streamFinalizerKey},
				},
				Spec: apis.StreamSpec{
					Name: name,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		jc.PrependReactor("update", "streams", func(a k8stesting.Action) (handled bool, o runtime.Object, err error) {
			ua, ok := a.(k8stesting.UpdateAction)
			if !ok {
				return false, nil, nil
			}

			return true, ua.GetObject(), nil
		})

		wantEvents := 1
		frec := record.NewFakeRecorder(wantEvents)
		ctrl := &Controller{
			ctx:          context.Background(),
			streamLister: informer.Lister(),
			ji:           jc.JetstreamV1(),
			rec:          frec,

			sc: &mockStreamClient{
				existsOK: true,
			},
		}

		if err := ctrl.processStream("default", name); err != nil {
			t.Fatal(err)
		}

		if got := len(frec.Events); got != wantEvents {
			t.Error("unexpected number of events")
			t.Fatalf("got=%d; want=%d", got, wantEvents)
		}

		gotEvent := <-frec.Events
		if !strings.Contains(gotEvent, "Deleting") {
			t.Error("unexpected event")
			t.Fatalf("got=%s; want=%s", gotEvent, "Deleting...")
		}
	})

	t.Run("update stream", func(t *testing.T) {
		jc := clientsetfake.NewSimpleClientset()
		informerFactory := informers.NewSharedInformerFactory(jc, 0)
		informer := informerFactory.Jetstream().V1().Streams()

		name := "my-stream"

		err := informer.Informer().GetStore().Add(
			&apis.Stream{
				ObjectMeta: k8smeta.ObjectMeta{
					Namespace:  "default",
					Name:       name,
					Generation: 2,
				},
				Spec: apis.StreamSpec{
					Name: name,
				},
				Status: apis.StreamStatus{
					ObservedGeneration: 1,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		jc.PrependReactor("update", "streams", func(a k8stesting.Action) (handled bool, o runtime.Object, err error) {
			ua, ok := a.(k8stesting.UpdateAction)
			if !ok {
				return false, nil, nil
			}

			return true, ua.GetObject(), nil
		})

		wantEvents := 2
		frec := record.NewFakeRecorder(wantEvents)
		ctrl := &Controller{
			ctx:          context.Background(),
			streamLister: informer.Lister(),
			ji:           jc.JetstreamV1(),
			rec:          frec,

			sc: &mockStreamClient{
				existsOK: true,
			},
		}

		if err := ctrl.processStream("default", name); err != nil {
			t.Fatal(err)
		}

		if got := len(frec.Events); got != wantEvents {
			t.Error("unexpected number of events")
			t.Fatalf("got=%d; want=%d", got, wantEvents)
		}

		for i := 0; i < len(frec.Events); i++ {
			gotEvent := <-frec.Events
			if !strings.Contains(gotEvent, "Updat") {
				t.Error("unexpected event")
				t.Fatalf("got=%s; want=%s", gotEvent, "Updating/Updated...")
			}
		}
	})

	t.Run("create stream", func(t *testing.T) {
		jc := clientsetfake.NewSimpleClientset()
		informerFactory := informers.NewSharedInformerFactory(jc, 0)
		informer := informerFactory.Jetstream().V1().Streams()

		name := "my-stream"

		err := informer.Informer().GetStore().Add(
			&apis.Stream{
				ObjectMeta: k8smeta.ObjectMeta{
					Namespace:  "default",
					Name:       name,
					Generation: 1,
				},
				Spec: apis.StreamSpec{
					Name: name,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		jc.PrependReactor("update", "streams", func(a k8stesting.Action) (handled bool, o runtime.Object, err error) {
			ua, ok := a.(k8stesting.UpdateAction)
			if !ok {
				return false, nil, nil
			}

			return true, ua.GetObject(), nil
		})

		wantEvents := 2
		frec := record.NewFakeRecorder(wantEvents)
		ctrl := &Controller{
			ctx:          context.Background(),
			streamLister: informer.Lister(),
			ji:           jc.JetstreamV1(),
			rec:          frec,

			sc: &mockStreamClient{
				existsOK: false,
			},
		}

		if err := ctrl.processStream("default", name); err != nil {
			t.Fatal(err)
		}

		if got := len(frec.Events); got != wantEvents {
			t.Error("unexpected number of events")
			t.Fatalf("got=%d; want=%d", got, wantEvents)
		}

		for i := 0; i < len(frec.Events); i++ {
			gotEvent := <-frec.Events
			if !strings.Contains(gotEvent, "Creat") {
				t.Error("unexpected event")
				t.Fatalf("got=%s; want=%s", gotEvent, "Creating/Created...")
			}
		}
	})
}

func TestRunStreamQueue(t *testing.T) {
	t.Parallel()

	t.Run("bad item key", func(t *testing.T) {
		t.Parallel()

		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewNamedRateLimitingQueue(limiter, "StreamsTest")
		defer q.ShutDown()

		ctrl := &Controller{
			streamQueue: q,
		}

		key := "this/is/a/bad/key"
		q.Add(key)

		ctrl.processNextQueueItem()

		if got, want := q.Len(), 0; got != want {
			t.Error("unexpected number of items in queue")
			t.Fatalf("got=%d; want=%d", got, want)
		}

		if got, want := q.NumRequeues(key), 0; got != want {
			t.Error("unexpected number of requeues")
			t.Fatalf("got=%d; want=%d", got, want)
		}
	})

	t.Run("process error", func(t *testing.T) {
		t.Parallel()

		limiter := workqueue.DefaultControllerRateLimiter()
		q := workqueue.NewNamedRateLimitingQueue(limiter, "StreamsTest")
		defer q.ShutDown()

		jc := clientsetfake.NewSimpleClientset()
		informerFactory := informers.NewSharedInformerFactory(jc, 0)
		informer := informerFactory.Jetstream().V1().Streams()

		ns, name := "default", "mystream"

		err := informer.Informer().GetStore().Add(
			&apis.Stream{
				ObjectMeta: k8smeta.ObjectMeta{
					Namespace:  ns,
					Name:       name,
					Generation: 1,
				},
				Spec: apis.StreamSpec{
					Name: name,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		ctrl := &Controller{
			ctx:          context.Background(),
			streamQueue:  q,
			streamLister: informer.Lister(),
			ji:           jc.JetstreamV1(),
			sc: &mockStreamClient{
				connectErr: fmt.Errorf("bad connect"),
			},
		}

		key := fmt.Sprintf("%s/%s", ns, name)
		q.Add(key)

		maxGets := maxQueueRetries+1
		numRequeues := -1
		for i := 0; i < maxGets; i++ {
			if i == maxGets-1 {
				numRequeues = q.NumRequeues(key)
			}

			ctrl.processNextQueueItem()
		}

		if got, want := q.Len(), 0; got != want {
			t.Error("unexpected number of items in queue")
			t.Fatalf("got=%d; want=%d", got, want)
		}

		if got, want := numRequeues, 10; got != want {
			t.Error("unexpected number of requeues")
			t.Fatalf("got=%d; want=%d", got, want)
		}
	})
}