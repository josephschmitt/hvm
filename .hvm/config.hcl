// General config options go out here
debug = "info"

use = {
  "node": "16.3.0",
  "pnpm": "5.0.0",
  "bazel": "3.2.0"
}

package "pnpm" {
  source = "https://urbancompass.jfrog.io/urbancompass/api/npm/npm/pnpm/-/pnpm-${version}.tgz"
}
