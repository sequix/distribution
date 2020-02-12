package main

import (
	_ "net/http/pprof"

	"github.com/sequix/distribution/registry"
	_ "github.com/sequix/distribution/registry/auth/htpasswd"
	_ "github.com/sequix/distribution/registry/auth/silly"
	_ "github.com/sequix/distribution/registry/auth/token"
	_ "github.com/sequix/distribution/registry/proxy"
	_ "github.com/sequix/distribution/registry/storage/driver/azure"
	_ "github.com/sequix/distribution/registry/storage/driver/filesystem"
	_ "github.com/sequix/distribution/registry/storage/driver/gcs"
	_ "github.com/sequix/distribution/registry/storage/driver/inmemory"
	_ "github.com/sequix/distribution/registry/storage/driver/middleware/alicdn"
	_ "github.com/sequix/distribution/registry/storage/driver/middleware/cloudfront"
	_ "github.com/sequix/distribution/registry/storage/driver/middleware/redirect"
	_ "github.com/sequix/distribution/registry/storage/driver/oss"
	_ "github.com/sequix/distribution/registry/storage/driver/s3-aws"
	_ "github.com/sequix/distribution/registry/storage/driver/swift"
)

func main() {
	registry.RootCmd.Execute()
}
