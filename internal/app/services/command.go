package services

import "os/exec"

type Command interface {
	Keygen(name string, configType string) error
}

type command struct {
	path string
}

func NewCommand(path string) *command {
	return &command{
		path: path,
	}
}

func (c *command) execute(cmd string, params ...string) error {
	if err := exec.Command(c.path+cmd, params...).Run(); err != nil {
		return err
	}

	return nil
}

func (c *command) Keygen(name string, configType string) error {
	err := c.execute("keygen.sh", name, configType)
	if err != nil {
		return err
	}

	return nil
}
