deepequal
=========

A version of Go's reflect.DeepEqual that includes reasons for failing inequality checks.

This is simply a copy of the DeepEqual function from the Go standard library, enhanced to give a reason for a comparison failure.

Usage:

```
import "github.com/iand/deepequal"

func main() {
    x := map[string]int{ "a":1, "b", 5}    
    y := map[string]int{ "a":1, "b", 8}    

    equal, reason := deepequal.Compare(x, y)
    if !equal {
        println(reason)
    }
}
```

