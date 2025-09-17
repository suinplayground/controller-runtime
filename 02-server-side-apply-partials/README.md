# Experiment: Server-Side Apply with Partial Manifests

This experiment investigates the behavior of Server-Side Apply (SSA) when applying partial manifests for the same resource with the same field manager.

## Goal

To understand how Kubernetes handles fields that are present in one apply request but omitted in a subsequent one from the same field manager.

## Steps

The test `TestDemo` in `demo_test.go` performs the following actions:

1.  **First Apply**: A `Cat` resource named `my-cat` is created by applying a partial manifest. This manifest only specifies the `spec.breed` field. The operation uses `cat-owner` as the field manager.

    ```go
    // Apply a cat with partial spec
    cat := applycatv1.Cat("my-cat", "default").
        WithSpec(applycatv1.CatSpec().
            WithBreed("Maine Coon"))

    err := cl.Apply(ctx, cat, client.FieldOwner("cat-owner"), client.ForceOwnership)
    ```

2.  **Second Apply**: Another partial manifest is applied to the same `my-cat` resource. This manifest only specifies the `spec.color` field, omitting `spec.breed`. The field manager is still `cat-owner`.

    ```go
    // Apply a cat with partial spec
    cat := applycatv1.Cat("my-cat", "default").
        WithSpec(applycatv1.CatSpec().
            WithColor("Black"))

    err := cl.Apply(ctx, cat, client.FieldOwner("cat-owner"), client.ForceOwnership)
    ```

## Observation

-   After the first apply, the `Cat` resource's spec contains `breed: "Maine Coon"`. The `managedFields` entry shows that `cat-owner` owns `spec.breed`.

    ```json
    "spec": {
      "breed": "Maine Coon"
    },
    "managedFields": [
      {
        "manager": "cat-owner",
        ...
        "fieldsV1": {
          "f:spec": {
            "f:breed": {}
          }
        }
      }
    ]
    ```

-   After the second apply, the `Cat` resource's spec is updated to contain `color: "Black"`, but the `breed` field is **removed**. The `managedFields` entry for `cat-owner` is updated to show ownership of only `spec.color`.

    ```json
    "spec": {
      "color": "Black"
    },
    "managedFields": [
      {
        "manager": "cat-owner",
        ...
        "fieldsV1": {
          "f:spec": {
            "f:color": {}
          }
        }
      }
    ]
    ```

## Conclusion

The behavior observed in this experiment is a core feature of Server-Side Apply's field management. When an applier (a "field manager") sends a manifest, it is claiming ownership of the fields included in that manifest.

The official Kubernetes documentation on [Server-Side Apply](https://kubernetes.io/docs/reference/using-api/server-side-apply/#deleting-a-field) explains what happens when a previously managed field is omitted from a subsequent apply request:

> If you remove a field from a manifest and apply that manifest, Server-Side Apply checks if there are any other field managers that also own the field. If the field is not owned by any other field managers, it is either deleted from the live object or reset to its default value, if it has one.

This directly explains the outcome of the experiment:
1.  After the first apply, the `cat-owner` manager was the sole owner of the `spec.breed` field.
2.  The second apply request from `cat-owner` omitted `spec.breed`.
3.  Because no other field manager owned `spec.breed`, Server-Side Apply deleted the field from the live object.

This mechanism ensures that when an applier stops managing a field, it is cleanly removed from the resource instead of becoming cruft. To avoid unintentionally dropping fields, an applier must include all the fields it intends to manage in every apply request.

---

## Addendum: Experiment with a Different Field Manager

A subsequent experiment was conducted to observe the behavior when a different field manager applies a partial manifest to the same resource.

### Steps

Building on the previous state, a third `Apply` operation was performed:

3.  **Third Apply (Different Manager)**: A new field manager, `age-controller`, applies a partial manifest containing only the `spec.age` field. The existing field `spec.color` is managed by `cat-owner`.

### Observation

-   After the third apply, the `Cat` resource's spec contains both `color: "Black"` and `age: 3`.
-   The `managedFields` list now has two entries, one for each field manager, correctly tracking which manager owns which field.

    ```json
    "spec": {
      "color": "Black",
      "age": 3
    },
    "managedFields": [
      {
        "manager": "age-controller",
        "operation": "Apply",
        ...
        "fieldsV1": {
          "f:spec": {
            "f:age": {}
          }
        }
      },
      {
        "manager": "cat-owner",
        "operation": "Apply",
        ...
        "fieldsV1": {
          "f:spec": {
            "f:color": {}
          }
        }
      }
    ]
    ```

### Conclusion

This confirms the core design of Server-Side Apply. Since the `age-controller`'s apply request did not omit any fields that *it* previously owned (it owned none), it simply claimed ownership of `spec.age`. This operation did not affect the fields owned by `cat-owner`.

This demonstrates how Server-Side Apply allows multiple independent controllers or actors to manage different fields of the same resource safely and declaratively.
