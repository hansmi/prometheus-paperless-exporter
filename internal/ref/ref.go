package ref

// Return a pointer to a value.
//
// https://github.com/golang/go/issues/45624
func Ref[T any](x T) *T {
	return &x
}
