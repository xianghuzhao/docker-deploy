package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type dockerHost interface {
	env(contextDir string) (map[string]string, error)
	clean() error
}

type (
	config struct {
		contextDir  string
		keepContext bool
		host        string

		sshKey string

		tlsVerify bool
		tlsCACert string
		tlsCert   string
		tlsKey    string

		errorExit bool
		script    []string
	}

	plugin struct {
		config config

		scheme     string
		dockerHost dockerHost
	}
)

func (p *plugin) getScheme() error {
	index := strings.Index(p.config.host, "://")
	if index == -1 {
		p.scheme = "tcp"
	} else {
		scheme := p.config.host[:index]
		switch scheme {
		case "tcp", "ssh":
			p.scheme = scheme
		default:
			return fmt.Errorf("Scheme not supported: %s", scheme)
		}
	}
	return nil
}

func (p *plugin) initDockerHost() error {
	switch p.scheme {
	case "tcp":
		p.dockerHost = &dockerHostTCP{
			host:      p.config.host,
			tlsVerify: p.config.tlsVerify,
			tlsCACert: p.config.tlsCACert,
			tlsCert:   p.config.tlsCert,
			tlsKey:    p.config.tlsKey,
		}
	case "ssh":
		p.dockerHost = &dockerHostSSH{
			host:   p.config.host,
			sshKey: p.config.sshKey,
		}
	default:
		return fmt.Errorf("Scheme could not be initialized: %s", p.scheme)
	}

	return nil
}

func (p *plugin) getContextDir() (string, error) {
	if p.config.contextDir != "" {
		contextDir := p.config.contextDir
		err := os.MkdirAll(contextDir, 0700)
		if err != nil {
			logrus.Errorf("Create context directory failed: %s", err)
			return "", err
		}
		return contextDir, nil
	}

	contextDir, err := ioutil.TempDir("", "docker-remote-")
	if err != nil {
		logrus.Errorf("Create temp context directory failed: %s", err)
		return "", err
	}

	return contextDir, nil
}

func (p *plugin) executeScript(envs map[string]string) error {
	args := []string{"-c", "-x"}

	if p.config.errorExit {
		args = append(args, "-e")
	}

	fullScript := strings.Join(p.config.script, "\n")
	args = append(args, fullScript)

	fullEnv := os.Environ()
	for k, v := range envs {
		fullEnv = append(fullEnv, k+"="+v)
	}

	cmd := exec.Command("/bin/sh", args...)
	cmd.Env = fullEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Script execution error: %s", err)
		return err
	}
	return nil
}

func (p *plugin) clean(contextDir string) {
	if p.config.keepContext {
		return
	}

	err := p.dockerHost.clean()
	if err != nil {
		logrus.Errorf(`Clean docker host error: %s`, err)
	}

	if p.config.contextDir == "" && contextDir != "" {
		err := os.RemoveAll(contextDir)
		if err != nil {
			logrus.Errorf(`Remove temp context directory "%s" error: %s`, contextDir, err)
		}
	}
}

// exec executes the docker commands with remote host
func (p *plugin) exec() error {
	var err error

	err = p.getScheme()
	if err != nil {
		return err
	}
	logrus.Infof("Scheme for docker host: %s", p.scheme)

	err = p.initDockerHost()
	if err != nil {
		return err
	}

	contextDir, err := p.getContextDir()
	if err != nil {
		return err
	}
	logrus.Infof("Context directory for docker remote: %s", contextDir)

	defer p.clean(contextDir)

	envs, err := p.dockerHost.env(contextDir)
	if err != nil {
		return err
	}
	logrus.Infof("Env for docker remote: %s", envs)

	err = p.executeScript(envs)
	if err != nil {
		return err
	}

	return nil
}
