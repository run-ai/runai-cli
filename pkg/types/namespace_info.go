package types

const ALL_PROJECTS = "all"

type NamespaceInfo struct {
	Namespace             string
	ProjectName           string
	BackwardCompatability bool
}
