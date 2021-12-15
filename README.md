# docker-remote

Drone plugin for running docker commands on remote host.
SSH and TCP (HTTPS) are supported.

## Security consideration for ssh scheme

It is ssh key could 

The `authorized_keys` could include more configurations for a specified key.
Options could be added at the beginning of the public key line.
For more details, check the [official docs](https://www.ssh.com/academy/ssh/authorized_keys/openssh).

These options disable interactive login using this key:

```
no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty ssh-rsa XXXXX user@host
```

The allowd IP source could also be restricted with `from` option:

```
from="xx.xx.xx.xx",no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty ssh-rsa XXXXX user@host
```

Now user can only access the host with this key to run command directly.
In case we do not want arbitrary commands to be executed,
only the docker commands should be allowed,
write a script which restricts the `docker system dial-stdio` command to run:

```shell
#!/bin/sh

if [ "$SSH_ORIGINAL_COMMAND" != 'docker system dial-stdio' ]; then
  echo "Command not allowed: $SSH_ORIGINAL_COMMAND"
  exit 1
fi

# Run the command
eval "$SSH_ORIGINAL_COMMAND"
```

Make sure the file is executable：

```
$ chmod +x ~/.ssh/filter-docker.sh
```

Add this command to the `authorized_keys` line:

```
command="~/.ssh/filter-docker.sh",from="xx.xx.xx.xx",no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty ssh-rsa XXXXX user@host
```
