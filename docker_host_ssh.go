package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type dockerHostSSH struct {
	host   string
	sshKey string

	hostInfo struct {
		workDir    string
		configPath string
		sshKeyPath string

		sshDir        string
		sshConfigFile string
		startLine     string
		endLine       string

		hostSection string
		hostName    string
		username    string
		port        string
	}
}

const sshConfigCommon = `  StrictHostKeyChecking    no
  UserKnownHostsFile       /dev/null
  ControlMaster            auto
  ControlPath              ~/.ssh/control-%C
  ControlPersist           30s
`

const promptCommentStart = "### DOCKER_REMOTE_START %s -- These lines are added by docker-remote automatically, which can be safely removed"
const promptCommentEnd = "### DOCKER_REMOTE_END %s"

func (h *dockerHostSSH) parseHost() error {
	u, err := url.Parse(h.host)
	if err != nil {
		logrus.Errorf(`Host parse error for "%s": %s`, h.host, err)
		return err
	}

	h.hostInfo.hostName, h.hostInfo.port, err = net.SplitHostPort(u.Host)
	if err != nil {
		logrus.Debugf(`Host port parse error for "%s": %s`, u.Host, err)
		h.hostInfo.hostName = u.Host
	}

	if h.hostInfo.hostName == "" {
		logrus.Errorf(`Host name is empty for host "%s"`, h.host)
		return fmt.Errorf("Host name is empty")
	}

	h.hostInfo.username = u.User.Username()

	return nil
}

func (h *dockerHostSSH) createCertDir(contextDir string) error {
	h.hostInfo.workDir = path.Join(contextDir, "docker-ssh")
	err := os.MkdirAll(h.hostInfo.workDir, 0700)
	if err != nil {
		logrus.Errorf("Create cert path directory failed: %s", err)
		return err
	}
	return nil
}

func (h *dockerHostSSH) createSSHKey() error {
	if h.sshKey == "" {
		return nil
	}

	// Make sure to write the ssh key with a trailing line feed
	if h.sshKey[len(h.sshKey)-1] != '\n' {
		h.sshKey += "\n"
	}

	h.hostInfo.sshKeyPath = path.Join(h.hostInfo.workDir, "ssh-key")

	err := os.WriteFile(h.hostInfo.sshKeyPath, []byte(h.sshKey), 0400)
	if err != nil {
		logrus.Errorf("Create SSH key file failed: %s", err)
		return err
	}

	return nil
}

func (h *dockerHostSSH) createSSHConfig() error {
	h.hostInfo.configPath = path.Join(h.hostInfo.workDir, "ssh_host_config")

	h.hostInfo.hostSection = "docker-remote-" + fmt.Sprint(time.Now().UnixNano())

	configFragments := []string{"Host " + h.hostInfo.hostSection}
	if h.hostInfo.hostName != "" {
		configFragments = append(configFragments, "  HostName                 "+h.hostInfo.hostName)
	}
	if h.hostInfo.port != "" {
		configFragments = append(configFragments, "  Port                     "+h.hostInfo.port)
	}
	if h.hostInfo.username != "" {
		configFragments = append(configFragments, "  User                     "+h.hostInfo.username)
	}
	if h.hostInfo.sshKeyPath != "" {
		configFragments = append(configFragments, "  IdentityFile             "+h.hostInfo.sshKeyPath)
	}
	configFragments = append(configFragments, sshConfigCommon)

	err := os.WriteFile(h.hostInfo.configPath, []byte(strings.Join(configFragments, "\n")), 0400)
	if err != nil {
		logrus.Errorf("Create ssh config file failed: %s", err)
		return err
	}

	return nil
}

func (h *dockerHostSSH) appendInclude() error {
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Errorf("Get user home dir error: %s", err)
		return err
	}

	h.hostInfo.sshDir = path.Join(home, ".ssh")
	err = os.MkdirAll(h.hostInfo.sshDir, 0700)
	if err != nil {
		logrus.Errorf("Create user ssh directory failed: %s", err)
		return err
	}

	h.hostInfo.sshConfigFile = path.Join(h.hostInfo.sshDir, "config")
	f, err := os.OpenFile(h.hostInfo.sshConfigFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		logrus.Errorf("Open user ssh config error: %s", err)
		return err
	}

	defer f.Close()

	h.hostInfo.startLine = fmt.Sprintf(promptCommentStart, h.hostInfo.hostSection)
	h.hostInfo.endLine = fmt.Sprintf(promptCommentEnd, h.hostInfo.hostSection)

	line := fmt.Sprintf("\n%s\nMatch all\nInclude  %s\n%s\n", h.hostInfo.startLine, h.hostInfo.configPath, h.hostInfo.endLine)
	if _, err = f.WriteString(line); err != nil {
		logrus.Errorf("Patch user ssh config error: %s", err)
		return err
	}

	return nil
}

func (h *dockerHostSSH) removeInclude() error {
	sshConfigContent, err := ioutil.ReadFile(h.hostInfo.sshConfigFile)
	if err != nil {
		logrus.Errorf("Read ssh config file failed: %s", err)
		return err
	}

	pattern := fmt.Sprintf("(?s)\n%s\n.*?\n%s\n", h.hostInfo.startLine, h.hostInfo.endLine)
	logrus.Debugf("Replace pattern: %s", pattern)

	re, err := regexp.Compile(pattern)
	if err != nil {
		logrus.Errorf("Compile regex pattern error: %s", err)
		return err
	}

	sshConfigContent = re.ReplaceAll(sshConfigContent, []byte{})

	err = os.WriteFile(h.hostInfo.sshConfigFile, sshConfigContent, 0600)
	if err != nil {
		logrus.Errorf("Write ssh config new content error: %s", err)
		return err
	}

	return nil
}

func (h *dockerHostSSH) removeWorkDir() error {
	if h.hostInfo.workDir == "" {
		return nil
	}

	err := os.RemoveAll(h.hostInfo.workDir)
	if err != nil {
		logrus.Errorf("Remove cert path error: %s", err)
	}

	h.hostInfo.workDir = ""

	return nil
}

func (h *dockerHostSSH) env(contextDir string) (map[string]string, error) {
	var err error
	envs := make(map[string]string)

	err = h.parseHost()
	if err != nil {
		return nil, err
	}

	err = h.createCertDir(contextDir)
	if err != nil {
		return nil, err
	}

	err = h.createSSHKey()
	if err != nil {
		return nil, err
	}

	err = h.createSSHConfig()
	if err != nil {
		return nil, err
	}

	err = h.appendInclude()
	if err != nil {
		return nil, err
	}

	envs["DOCKER_HOST"] = "ssh://" + h.hostInfo.hostSection

	return envs, nil
}

func (h *dockerHostSSH) clean() error {
	h.removeInclude()

	h.removeWorkDir()

	return nil
}
