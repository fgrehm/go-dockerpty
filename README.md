# go-dockerpty

Provides the functionality needed to operate the pseudo-tty (PTY) allocated to a
docker container, using the [Go client](https://github.com/fsouza/go-dockerclient).

Inspired by https://github.com/d11wtq/dockerpty/

## Usage

The following example will run Ubuntu Trusty 14.04 in a docker container and
place the user at the shell prompt via Go.

This obviously only works when run in a terminal.

```go
package main

import (
	"fmt"
	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
)

func main() {
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)

	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        "ubuntu:14.04",
			Cmd:          []string{"/bin/bash"},
			OpenStdin:    true,
			StdinOnce:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		},
	})

	dockerpty.Start(client, container)
}
```

When the dockerpty is started, control is yielded to the container's PTY until
the container exits, or the container's PTY is closed.

This is a safe operation and all resources are restored back to their original
states.

## Caveats

* In order to properly detect that the container has stopped and return the
  control to the main app, we start an event listener on the background to
  detect that the container has stopped. On an ideal world we shouldn't need
  this but it was the only way I could work around [fsouza/go-dockerclient#117](https://github.com/fsouza/go-dockerclient/issues/117).
* Does not show the initial prompt when running a simple `sh` shell (with
  `bash` it works just fine).
