package pkg

import (
	"path"
	"regexp"
)

var imgRE = regexp.MustCompile(`!\[[^\]]*\]\(\s*<?(/media/[^)\s>]+)>?(?:\s+"[^"]*")?\s*\)`)

// MediaRefs returns map of media s3_key found in md using regex.
func MediaRefs(md string) map[string]struct{} {
	refs := map[string]struct{}{}
	for _, m := range imgRE.FindAllStringSubmatch(md, -1) {
		refs[path.Base(m[1])] = struct{}{}
	}
	return refs
}

// Difference returns the elements of a that are not in b, i.e. a \ b.
//
// NOTE: The result is in unspecified order. The inputs are not modified.
func Difference[T comparable](a, b map[T]struct{}) []T {
	var out []T
	for k := range a {
		if _, ok := b[k]; !ok {
			out = append(out, k)
		}
	}
	return out
}
