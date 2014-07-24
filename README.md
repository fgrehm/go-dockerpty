# go-dockerpty

Provides the functionality needed to operate the pseudo-tty (PTY) allocated to a
docker container, using the [Go client](https://github.com/fsouza/go-dockerclient).

Inspired by https://github.com/d11wtq/dockerpty

## Usage

This package provides a single function: `dockerpty.Start`. The following example
will run Busybox in a docker container and place the user at the shell prompt via
Go. It is the same as running `docker run -ti --rm busybox /bin/sh`.

_This obviously only works when run in a terminal._

```go
package main

import (
	"fmt"
	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
	"os"
)

func main() {
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)

	// Create container
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        "busybox",
			Cmd:          []string{"/bin/sh"},
			OpenStdin:    true,
			StdinOnce:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		},
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Cleanup when done
	defer func() {
		client.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID,
			Force: true,
		})
	}()

	// Fire up the console
	if err = dockerpty.Start(client, container, &docker.HostConfig{}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```

When `dockerpty.Start` is called, control is yielded to the container's PTY until
the container exits, or the container's PTY is closed.

This is a safe operation and all resources should be restored back to their original
states.

[![baby-gopher](https://raw2.github.com/drnic/babygopher-site/gh-pages/images/babygopher-badge.png)](http://www.babygopher.org)
