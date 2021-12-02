package utils

import "regexp"

var SemVerRegex = regexp.MustCompile(`^(<=|<|>|>=|\^)?(0|[1-9]\d*)\.(0|[1-9]d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
