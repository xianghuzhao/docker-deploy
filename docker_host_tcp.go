package main

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

type dockerHostTCP struct {
	host      string
	tlsVerify bool
	tlsCACert string
	tlsCert   string
	tlsKey    string

	certPath string
}

func (h *dockerHostTCP) createFiles() error {
	if h.certPath != "" {
		err := os.MkdirAll(h.certPath, 0700)
		if err != nil {
			logrus.Errorf("Create cert path directory failed: %s", err)
			return err
		}
	}

	if h.tlsCACert != "" {
		err := os.WriteFile(path.Join(h.certPath, "ca.pem"), []byte(h.tlsCACert), 0400)
		if err != nil {
			logrus.Errorf("Create CA cert file failed: %s", err)
			return err
		}
	}

	if h.tlsCert != "" {
		err := os.WriteFile(path.Join(h.certPath, "cert.pem"), []byte(h.tlsCert), 0400)
		if err != nil {
			logrus.Errorf("Create cert file failed: %s", err)
			return err
		}
	}

	if h.tlsKey != "" {
		err := os.WriteFile(path.Join(h.certPath, "key.pem"), []byte(h.tlsKey), 0400)
		if err != nil {
			logrus.Errorf("Create key file failed: %s", err)
			return err
		}
	}

	return nil
}

func (h *dockerHostTCP) env(contextDir string) (map[string]string, error) {
	envs := make(map[string]string)

	envs["DOCKER_HOST"] = h.host

	if h.tlsVerify {
		h.certPath = path.Join(contextDir, "docker-cert")
		h.createFiles()
		envs["DOCKER_CERT_PATH"] = h.certPath
		envs["DOCKER_TLS_VERIFY"] = "1"
	}

	return envs, nil
}

func (h *dockerHostTCP) clean() error {
	if h.certPath == "" {
		return nil
	}

	err := os.RemoveAll(h.certPath)
	if err != nil {
		logrus.Errorf("Remove cert path error: %s", err)
	}

	h.certPath = ""

	return nil
}
