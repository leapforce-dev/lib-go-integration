package integration

type Environment string

const (
	environmentNone Environment = ""
	EnvironmentTest Environment = "test"
	EnvironmentLive Environment = "live"
)
