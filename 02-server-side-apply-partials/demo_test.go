package serversideapply

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	catv1 "github.com/suinplayground/controller-runtime-playground/api/v1"
	applycatv1 "github.com/suinplayground/controller-runtime-playground/client/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestDemo(t *testing.T) {
	cl := NewClient(t)
	ctx := context.Background()

	getCat := func () {
		fmt.Println("Getting cat")
		cat := &catv1.Cat{}
		err := cl.Get(ctx, client.ObjectKey{Name: "my-cat", Namespace: "default"}, cat)
		if err != nil {
			t.Fatalf("failed to get cat: %v", err)
		}

		json, err := json.MarshalIndent(cat, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal cat: %v", err)
		}
		fmt.Printf("cat: %s\n", json)
	}

	func() {
		fmt.Println("Applying cat with partial spec with breed")
		cat := applycatv1.Cat("my-cat", "default").
			WithSpec(applycatv1.CatSpec().
				WithBreed("Maine Coon"))

		err := cl.Apply(ctx, cat, client.FieldOwner("cat-owner"), client.ForceOwnership)
		if err != nil {
			t.Fatalf("failed to patch cat: %v", err)
		}

		t.Cleanup(func() {
			_ = cl.Delete(ctx, &catv1.Cat{ObjectMeta: metav1.ObjectMeta{Name: "my-cat", Namespace: "default"}})
		})
	}()

	getCat()

	func() {
		fmt.Println("Applying cat with partial spec with color")
		cat := applycatv1.Cat("my-cat", "default").
			WithSpec(applycatv1.CatSpec().
				WithColor("Black"))

		err := cl.Apply(ctx, cat, client.FieldOwner("cat-owner"), client.ForceOwnership)
		if err != nil {
			t.Fatalf("failed to patch cat: %v", err)
		}
	}()

	getCat()

	func() {
		fmt.Println("Applying cat with partial spec with age by different field manager")
		cat := applycatv1.Cat("my-cat", "default").
			WithSpec(applycatv1.CatSpec().
				WithAge(3))

		err := cl.Apply(ctx, cat, client.FieldOwner("age-controller"))
		if err != nil {
			t.Fatalf("failed to patch cat: %v", err)
		}
	}()

	getCat()
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
