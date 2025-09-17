package serversideapply

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	catv1 "github.com/suinplayground/controller-runtime-playground/01-server-side-apply/api/v1"
	applycatv1 "github.com/suinplayground/controller-runtime-playground/01-server-side-apply/client/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	applymetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestDemo(t *testing.T) {
	cl := NewClient(t)
	ctx := context.Background()

	func() {
		// Create a cat
		cat := applycatv1.Cat("my-cat", "default").
			WithSpec(applycatv1.CatSpec().
				WithBreed("Maine Coon").
				WithColor("Black").
				WithAge(3))

		err := cl.Apply(ctx, cat, client.FieldOwner("cat-owner"), client.ForceOwnership)
		if err != nil {
			t.Fatalf("failed to patch cat: %v", err)
		}

		t.Cleanup(func() {
			_ = cl.Delete(ctx, &catv1.Cat{ObjectMeta: metav1.ObjectMeta{Name: "my-cat", Namespace: "default"}})
		})
	}()

	func() {
		// Update the patch conditions
		patch := applycatv1.Cat("my-cat", "default").
			WithStatus(applycatv1.CatStatus().
				WithConditions(applymetav1.Condition().
					WithType("Sleepy").
					WithStatus("True").
					WithReason("Sleepy").
					WithMessage("Cat is sleepy").
					WithLastTransitionTime(metav1.Now())))
		err := cl.Status().Patch(ctx,
			unstruct(patch),
			client.Apply,
			client.ForceOwnership,
			client.FieldOwner("sleepiness-controller"),
		)
		if err != nil {
			t.Fatalf("failed to patch cat: %v", err)
		}
	}()

	func() {
		// Update the patch conditions by different controller (manager)
		patch := applycatv1.Cat("my-cat", "default").
			WithStatus(applycatv1.CatStatus().
				WithConditions(applymetav1.Condition().
					WithType("Happy").
					WithStatus("True").
					WithReason("Happy").
					WithMessage("Cat is happy").
					WithLastTransitionTime(metav1.Now())))
		err := cl.Status().Patch(ctx,
			unstruct(patch),
			client.Apply,
			client.ForceOwnership,
			client.FieldOwner("happiness-controller"),
		)
		if err != nil {
			t.Fatalf("failed to patch cat: %v", err)
		}
	}()

	func() {
		// Get the cat
		cat := &catv1.Cat{}
		err := cl.Get(ctx, client.ObjectKey{Name: "my-cat", Namespace: "default"}, cat)
		if err != nil {
			t.Fatalf("failed to get cat: %v", err)
		}

		json, err := json.MarshalIndent(cat, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal cat: %v", err)
		}
		fmt.Printf("cat: %s", json)
	}()
}

func unstruct(apply any) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(apply)
	if err != nil {
		panic(err)
	}
	u.Object = obj
	return u
}
