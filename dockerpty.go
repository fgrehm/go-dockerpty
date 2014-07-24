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
		terminalFd    uintptr
		oldState      *term.State
		out           io.Writer     = os.Stdout
		exitedChannel chan struct{} = make(chan struct{})
	)

	if file, ok := out.(*os.File); ok {
		terminalFd = file.Fd()
	} else {
		return errors.New("Not a terminal!")
	}

	// This goroutine will listen to Docker events and will signal that is has
	// stopped at the exitedChannel
	go listenForContainerExit(client, container.ID, exitedChannel)

	// Set up the pseudo terminal
	oldState, err = term.SetRawTerminal(terminalFd)
	if err != nil {
		return
	}

	// Attach to the container on a separate goroutine
	go attachToContainer(client, container.ID)

	// And finally start it
	err = client.StartContainer(container.ID, hostConfig)
	if err != nil {
		return err
	}

	// Wait until the container has exited
	<-exitedChannel

	// Clean up after the container has exited
	term.RestoreTerminal(terminalFd, oldState)

	return err
}

func listenForContainerExit(client *docker.Client, containerID string, exitedChannel chan struct{}) error {
	listenerChannel := make(chan *docker.APIEvents)
	client.AddEventListener(listenerChannel)

	for {
		event := <-listenerChannel
		if event.ID == containerID && event.Status == "die" {
			exitedChannel <- struct{}{}
		}
	}
}

func attachToContainer(client *docker.Client, containerID string) {
	client.AttachToContainer(docker.AttachToContainerOptions{
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
}
