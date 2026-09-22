// Package exercises embebe los archivos YAML de ejercicios en el binario
// mediante go:embed. La carga y validación viven en internal/exercise.
package exercises

import "embed"

//go:embed */*.yaml
var FS embed.FS
