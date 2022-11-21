package services

type Command interface {
	Keygen(name string, configType string) error
	RestartWG() error
}
