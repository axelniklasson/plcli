package util

type Options struct {
	Slice                string
	NodeCount            int
	SkipHealthCheck      bool
	RemoveFaulty         bool
	AttachToSlice        bool
	Scale                int
	NodesFile            string
	GitBranch            string
	AppPath              string
	PrometheusSDPath     string
	NodeExporter         bool
	ShuffleNodes         bool
	SkipWriteHostsFile   bool
	Sudo                 bool
	BlacklistedHostnames string
	EnvVars              string
}
