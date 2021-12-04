package main

type dockerHostTCP struct {
	host      string
	tlsVerify bool
	tlsCACert string
	tlsCert   string
	tlsKey    string
}

func (host *dockerHostTCP) env(contextDir string) (map[string]string, error) {
	return nil, nil
}

func (host *dockerHostTCP) clean(dir string) error {
	return nil
}
