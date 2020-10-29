package types

const AllProjects = "all"

type NamespaceInfo struct {
	Namespace             string
	ProjectName           string
	BackwardCompatibility bool
}
