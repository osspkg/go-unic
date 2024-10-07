package ref

import (
	"fmt"
	"reflect"
	"sort"
)

func sortReflectValues(keys []reflect.Value) {
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprintf("%+v", keys[i]) < fmt.Sprintf("%+v", keys[j])
	})
}
