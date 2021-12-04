package main

type dockerHostSSH struct {
	host   string
	sshKey string
}

const sshConfigTemplate = `\
ControlMaster     auto
ControlPath       ~/.ssh/control-%C
ControlPersist    yes

Host %s
`

func (host *dockerHostSSH) env(contextDir string) (map[string]string, error) {
	return nil, nil
}

func (host *dockerHostSSH) clean(dir string) error {
	return nil
}
