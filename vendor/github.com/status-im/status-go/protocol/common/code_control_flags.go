package common

type CodeControlFlags struct {
	// AutoRequestHistoricMessages indicates whether we should automatically request
	// historic messages on getting online, connecting to store node, etc.
	AutoRequestHistoricMessages bool

	// CuratedCommunitiesUpdateLoopEnabled indicates whether we should disable the curated communities update loop.
	// Usually should be disabled in tests.
	CuratedCommunitiesUpdateLoopEnabled bool
}
