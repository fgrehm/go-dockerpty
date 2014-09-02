package dockerpty

import (
	"errors"
	"github.com/fgrehm/go-dockerpty/term"
	"github.com/fsouza/go-dockerclient"
	"io"
	"os"
)

func Start(client *docker.Client, container *docker.Container, hostConfig *docker.HostConfig) (err error) {
	var (
		terminalFd uintptr
		oldState   *term.State
		out        io.Writer = os.Stdout
	)

	if file, ok := out.(*os.File); ok {
		terminalFd = file.Fd()
	} else {
		return errors.New("Not a terminal!")
	}

	// Set up the pseudo terminal
	oldState, err = term.SetRawTerminal(terminalFd)
	if err != nil {
		return
	}

	// Clean up after the container has exited
	defer term.RestoreTerminal(terminalFd, oldState)

	// Attach to the container on a separate thread
	attachChan := make(chan error)
	go attachToContainer(client, container.ID, attachChan)

	// Start it
	err = client.StartContainer(container.ID, hostConfig)
	if err != nil {
		return
	}

	return <-attachChan
}

func attachToContainer(client *docker.Client, containerID string, errorChan chan error) {
	err := client.AttachToContainer(docker.AttachToContainerOptions{
		Container:    containerID,
		InputStream:  os.Stdin,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		RawTerminal:  true,
	})
	errorChan <- err
}
