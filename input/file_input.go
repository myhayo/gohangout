package input

import "github.com/childe/gohangout/input/file_input"

func init() {
	Register("File", file_input.NewFileInput)
}
