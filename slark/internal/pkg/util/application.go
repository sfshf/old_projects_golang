package util

import "github.com/nextsurfer/ground/pkg/rpc"

var applicationMaps = map[string]string{
	"org.n1xt.word": "word",
}

// ApplicationNameForContext get application name
func ApplicationNameForContext(ctx *rpc.Context) string {
	application, ok := applicationMaps[ctx.UADevice().APPBundleID]
	if !ok {
		return "test"
	}
	return application
}

func GenderDesc(gender int32) string {
	if gender < 0 || gender > 2 {
		return "uninformed"
	}
	return []string{"male", "female", "uninformed"}[gender]
}
