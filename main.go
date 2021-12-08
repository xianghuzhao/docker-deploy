package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	version = "unknown"
)

func main() {
	app := &cli.App{
		Name:    "docker remote plugin",
		Usage:   "docker remote plugin",
		Action:  run,
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Value:   "INFO",
				Usage:   "log level, DEBUG/INFO/WARN/ERROR/FATAL",
				EnvVars: []string{"PLUGIN_LOG_LEVEL"},
			},
			&cli.StringFlag{
				Name:    "env-file",
				Aliases: []string{"e"},
				Value:   ".env",
				Usage:   "source env file",
				EnvVars: []string{"PLUGIN_ENV_FILE"},
			},
			&cli.StringFlag{
				Name:    "context-dir",
				Usage:   "context file directory",
				EnvVars: []string{"PLUGIN_CONTEXT_DIR"},
			},
			&cli.BoolFlag{
				Name:    "keep-context",
				Usage:   "reserve context file after execution",
				EnvVars: []string{"PLUGIN_KEEP_CONTEXT"},
			},
			&cli.StringFlag{
				Name:    "host",
				Usage:   "docker host",
				Aliases: []string{"H"},
				EnvVars: []string{"PLUGIN_HOST"},
			},
			&cli.StringFlag{
				Name:    "ssh-user",
				Usage:   "SSH username",
				EnvVars: []string{"PLUGIN_SSH_USER"},
			},
			&cli.StringFlag{
				Name:    "ssh-key",
				Usage:   "SSH key",
				EnvVars: []string{"PLUGIN_SSH_KEY"},
			},
			&cli.BoolFlag{
				Name:    "tls-verify",
				Usage:   "TLS verify",
				EnvVars: []string{"PLUGIN_TLS_VERIFY"},
			},
			&cli.StringFlag{
				Name:    "tls-ca-cert",
				Usage:   "TLS CA cert",
				EnvVars: []string{"PLUGIN_TLS_CA_CERT"},
			},
			&cli.StringFlag{
				Name:    "tls-cert",
				Usage:   "TLS cert",
				EnvVars: []string{"PLUGIN_TLS_CERT"},
			},
			&cli.StringFlag{
				Name:    "tls-key",
				Usage:   "TLS key",
				EnvVars: []string{"PLUGIN_TLS_KEY"},
			},
			&cli.BoolFlag{
				Name:    "error-exit",
				Usage:   "exit script on command error",
				EnvVars: []string{"PLUGIN_ERROR_EXIT"},
			},
			&cli.StringSliceFlag{
				Name:    "script",
				Usage:   "execute commands",
				EnvVars: []string{"PLUGIN_SCRIPT"},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func setupLog(level string) error {
	levelLogrus, err := logrus.ParseLevel(level)
	if err != nil {
		levelLogrus = logrus.InfoLevel
	}

	logrus.SetLevel(levelLogrus)
	return nil
}

func run(c *cli.Context) error {
	logLevel := c.String("log-level")
	if logLevel != "" {
		setupLog(logLevel)
		logrus.Debugf(`Setup log level: "%s"`, logLevel)
	}

	envFile := c.String("env-file")
	if envFile != "" {
		logrus.Debugf(`Load env-file: "%s"`, envFile)
		logrus.Debugf(`Load env: "%s"`, godotenv.Load(envFile))
	}

	p := &plugin{
		config: config{
			contextDir:  c.String("context-dir"),
			keepContext: c.Bool("keep-context"),
			host:        c.String("host"),
			sshUser:     c.String("ssh-user"),
			sshKey:      c.String("ssh-key"),
			tlsVerify:   c.Bool("tls-verify"),
			tlsCACert:   c.String("tls-ca-cert"),
			tlsCert:     c.String("tls-cert"),
			tlsKey:      c.String("tls-key"),
			errorExit:   c.Bool("error-exit"),
			script:      c.StringSlice("script"),
		},
	}

	return p.exec()
}
