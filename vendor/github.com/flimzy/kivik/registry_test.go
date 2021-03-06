package kivik

import (
	"sync"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik/driver"
)

// to protect the registry from concurrent tests
var registryMU sync.Mutex

func TestRegister(t *testing.T) {
	registryMU.Lock()
	defer registryMU.Unlock()
	t.Run("nil driver", func(t *testing.T) {
		defer func() {
			drivers = make(map[string]driver.Driver)
		}()
		p := func() (p interface{}) {
			defer func() {
				p = recover()
			}()
			Register("foo", nil)
			return ""
		}()
		if p.(string) != "kivik: Register driver is nil" {
			t.Errorf("Unexpected panic: %v", p)
		}
	})

	t.Run("duplicate driver", func(t *testing.T) {
		defer func() {
			drivers = make(map[string]driver.Driver)
		}()
		p := func() (p interface{}) {
			defer func() {
				p = recover()
			}()
			Register("foo", &mockDriver{})
			Register("foo", &mockDriver{})
			return ""
		}()
		if p.(string) != "kivk: Register called twice for driver foo" {
			t.Errorf("Unexpected panic: %v", p)
		}
	})

	t.Run("success", func(t *testing.T) {
		defer func() {
			drivers = make(map[string]driver.Driver)
		}()
		p := func() (p interface{}) {
			defer func() {
				p = recover()
			}()
			Register("foo", &mockDriver{})
			return ""
		}()
		if p != nil {
			t.Errorf("Unexpected panic: %v", p)
		}
		expected := map[string]driver.Driver{
			"foo": &mockDriver{},
		}
		if d := diff.Interface(expected, drivers); d != nil {
			t.Error(d)
		}
	})
}
