package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type dockerHost interface {
	env(contextDir string) (map[string]string, error)
	clean(dir string) error
}

type (
	config struct {
		contextDir string
		host       string
		sshUser    string
		sshKey     string
		tlsVerify  bool
		tlsCACert  string
		tlsCert    string
		tlsKey     string
		script     []string
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
		err := os.MkdirAll(contextDir, os.ModePerm)
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
	return nil
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
