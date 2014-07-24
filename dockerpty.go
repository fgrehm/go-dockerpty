package dockerpty

import (
	"errors"
	"github.com/fgrehm/go-dockerpty/term"
	"github.com/fsouza/go-dockerclient"
	"io"
	"os"
)

func Start(client *docker.Client, container *docker.Container) (err error) {
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

	listenerChannel := make(chan *docker.APIEvents)
	if err = client.AddEventListener(listenerChannel); err != nil {
		return
	}

	go func() {
		for {
			event := <-listenerChannel
			if event.ID == container.ID && event.Status == "die" {
				term.RestoreTerminal(terminalFd, oldState)
				exitedChannel <- struct{}{}
			}
		}
	}()

	oldState, err = term.SetRawTerminal(terminalFd)
	if err != nil {
		return
	}

	// Attach to the container on a separate goroutine
	go func() {
		err = client.AttachToContainer(docker.AttachToContainerOptions{
			Container:    container.ID,
			InputStream:  os.Stdin,
			OutputStream: os.Stdout,
			ErrorStream:  os.Stderr,
			Stdin:        true,
			Stdout:       true,
			Stderr:       true,
			Stream:       true,
			RawTerminal:  true,
		})
	}()

	// Start container
	err = client.StartContainer(container.ID, &docker.HostConfig{})
	if err != nil {
		return err
	}

	<-exitedChannel
	return err
}
