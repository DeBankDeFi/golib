package shared

type Info struct {
	Version Version
	AppName string
}

type Version struct {
	GitVersion   string
	GitCommit    string
	GitBranch    string
	GitTreeState string
	BuildTime    string
	GoVersion    string
	Compiler     string
	Platform     string
}
