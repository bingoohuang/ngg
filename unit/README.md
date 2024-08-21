# Humane Units

Just a few functions for helping humanize times and sizes.
 

## Sizes

This lets you take numbers like `82854982` and convert them to useful
strings like, `83 MB` or `79 MiB` (whichever you prefer).

Example:

```go
fmt.Printf("That file is %s.", unit.Bytes(82854982)) // That file is 83 MB.
```
