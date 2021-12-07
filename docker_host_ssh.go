package main

type dockerHostSSH struct {
	host   string
	sshKey string
}

const sshConfigTemplate = `\
Host %s
  HostName                 %s
  Port                     %s
  User                     %s
  IdentityFile             %s
  StrictHostKeyChecking    no
  UserKnownHostsFile       /dev/null
  ControlMaster            auto
  ControlPath              ~/.ssh/control-%C
  ControlPersist           yes
`

func (host *dockerHostSSH) env(contextDir string) (map[string]string, error) {
	return nil, nil
}

func (host *dockerHostSSH) clean() error {
	return nil
}
