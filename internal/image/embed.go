package image

import _ "embed"

//go:embed embed/Dockerfile
var dockerfile []byte

//go:embed embed/entrypoint.sh
var entrypointSh []byte
