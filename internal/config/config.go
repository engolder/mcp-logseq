package config

type Config struct {
	HTTP struct {
		Enable     bool
		ServerPort int
	}
	Stdio struct {
		Enable bool
	}
}
