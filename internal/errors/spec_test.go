package errors_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	errPkg "github.com/untappedtech/conduit/internal/errors"
)

func TestErrorSpec_Immutability(t *testing.T) {
	base := errPkg.ErrBadRequest

	// Baseline check: catalog error must not have details
	if len(base.Details) != 0 {
		t.Fatalf("expected base ErrBadRequest details to be empty, got %v", base.Details)
	}

	rawErr := errors.New("invalid parameter")
	instance1 := base.With(rawErr, "Field: username")

	// Base must remain untouched
	if len(base.Details) != 0 {
		t.Fatalf("base ErrBadRequest was mutated, details: %v", base.Details)
	}

	// Instance 1 must have details
	if len(instance1.Details) != 2 {
		t.Fatalf("expected instance1 to have 2 details, got %d: %v", len(instance1.Details), instance1.Details)
	}

	// Chaining / creating another instance from instance1
	instance2 := instance1.With(nil, "Extra detail")
	if len(instance1.Details) != 2 {
		t.Fatalf("expected instance1 details to remain 2, got %d", len(instance1.Details))
	}
	if len(instance2.Details) != 3 {
		t.Fatalf("expected instance2 details to be 3, got %d", len(instance2.Details))
	}
}

func TestErrorSpec_ConcurrentSafety(t *testing.T) {
	const goroutines = 100
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			requestErr := fmt.Errorf("error-from-req-%d", id)
			contextTag := fmt.Sprintf("req-id-%d", id)

			instance := errPkg.ErrNotFound.With(requestErr, contextTag)

			// Validate that this instance contains ONLY its own details
			if len(instance.Details) != 2 {
				t.Errorf("goroutine %d expected 2 details, got %d: %v", id, len(instance.Details), instance.Details)
				return
			}

			expectedErr := fmt.Sprintf("Error: error-from-req-%d", id)
			if instance.Details[0] != expectedErr {
				t.Errorf("goroutine %d: detail[0] mismatch. expected %q, got %q", id, expectedErr, instance.Details[0])
			}
			if instance.Details[1] != contextTag {
				t.Errorf("goroutine %d: detail[1] mismatch. expected %q, got %q", id, contextTag, instance.Details[1])
			}
		}(i)
	}

	wg.Wait()

	// Ensure ErrNotFound is still completely clean
	if len(errPkg.ErrNotFound.Details) != 0 {
		t.Fatalf("catalog ErrNotFound has leaked details: %v", errPkg.ErrNotFound.Details)
	}
}

func TestErrorSpec_ErrorInterface(t *testing.T) {
	base := errPkg.ErrNotFound
	if base.Error() != "Not Found: The requested resource does not exist." {
		t.Fatalf("unexpected Error() output: %s", base.Error())
	}

	withDetails := base.With(errors.New("db timeout"), "table: users")
	expected := "Not Found: The requested resource does not exist. (Error: db timeout; table: users)"
	if withDetails.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, withDetails.Error())
	}
}
