package model

// TypeInfo summarizes information about a struct type in the app source code.
type TypeInfo struct {
	StructName    string              // e.g. "Application"
	ImportPath    string              // e.g. "github.com/kenorld/egret/examples/chat/app/controllers"
	PackageName   string              // e.g. "controllers"
	MethodSpecs   []*MethodSpec       // Method specifications, the action functions
	EmbeddedTypes []*EmbeddedTypeName // Used internally to identify controllers that indirectly embed *egret.Controller.
}

// Return the type information as a properly formatted import string
func (s *TypeInfo) String() string {
	return s.ImportPath + "." + s.StructName
}
