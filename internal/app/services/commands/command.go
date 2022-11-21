package commands

import (
	"os/exec"
)

type Command struct {
	path string
}

func NewCommand(path string) *Command {
	return &Command{
		path: path,
	}
}

func (c *Command) execute(cmd string, params ...string) error {
	if err := exec.Command(c.path+cmd, params...).Run(); err != nil {
		return err
	}

	return nil
}

func (c *Command) Keygen(name string, configType string) error {
	err := c.execute("keygen.sh", name, configType)
	if err != nil {
		return err
	}

	return nil
}

func (c *Command) RestartWG() error {
	//TODO: Check output for containing error
	err := c.execute("restartwg.sh")
	if err != nil {
		return err
	}

	return nil
}
