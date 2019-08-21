package controller

import (
	"github.com/bysnupy/maintpage-operator/pkg/controller/maintpage"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, maintpage.Add)
}
