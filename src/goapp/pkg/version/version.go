package version

import (
	"fmt"
	"runtime"
)

var (
	GITTAG     = ""
	BUILD_TIME = ""
)

func String(app string) string {
	return fmt.Sprintf("app:%s, goversion:%s, git:%s, time:%s", app, runtime.Version(), GITTAG, BUILD_TIME)
}
