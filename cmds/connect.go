// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"bytes"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	mutex "sync"
	"time"

	"github.com/Bowery/bowery/api"
	"github.com/Bowery/bowery/broome"
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/delancey"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/bowery/sync"
	"github.com/Bowery/bowery/version"
	"github.com/Bowery/gopackages/keen"
	"github.com/Bowery/gopackages/log"
	"github.com/Bowery/gopackages/schemas"
	"github.com/garyburd/redigo/redis"
)

var (
	ENV       = os.Getenv("ENV")
	connected = true
)

func init() {
	Cmds["connect"] = &Cmd{
		Run:   connectRun,
		Usage: "connect",
		Short: "Bootup the app in the current directory.",
	}
}

func connectRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	hasUploaded := make([]string, 0)
	signals := make(chan os.Signal, 1)
	done := make(chan int, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	defer signal.Stop(signals)

	ver, err := api.GetVersion()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	if VersionOutOfDate(version.Version, ver) {
		log.Println("yellow", errors.ErrOutOfDate)
	}

	dev, err := getDeveloper()
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	log.Println("magenta", "Hey there,", strings.Split(dev.Developer.Name, " ")[0]+
		". Connecting you to Bowery now...")

	state, err := updateOrCreateApp(dev)
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	defer api.Disconnect(dev.Token)

	keen.AddEvent("bowery connect", map[string]interface{}{
		"user": dev,
		"app":  state,
	})

	syncer, servicesUploading, err := initiateSync(state)
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	defer syncer.Close()

	// Check if any services are still uploading after a while.
	go func() {
		<-time.After(20 * time.Second)
		hasNotUploaded := make([]string, 0)
		for _, service := range servicesUploading {
			found := false
			for _, s := range hasUploaded {
				if service.Name == s {
					found = true
					break
				}
			}

			if !found {
				hasNotUploaded = append(hasNotUploaded, service.Name)
			}
		}
		services := strings.Join(hasNotUploaded, ", ")

		if len(hasNotUploaded) > 0 {
			log.Println("yellow", "Service(s)", services, "still uploading...")
			log.Println("yellow", "Large applications and inital uploads can take some time to complete.")
			log.Println("yellow", "Check `bowery logs` to see the current status of the build.")
		}
	}()

	logChan := make(chan redis.Conn, 1)
	logFile, err := tailLogs(state, logChan)
	if err != nil {
		rollbar.Report(err)
		return 1
	}
	defer logFile.Close()

	go func() {
		<-signals
		done <- 0
	}()

	go apiStatus(dev.Token)

	// Watch for various events.
	go func() {
		for {
			select {
			case service := <-syncer.Upload:
				config := state.Config[service.Name]
				hasUploaded = append(hasUploaded, service.Name)

				if config.Path != "" {
					log.Println("cyan", "Service", service.Name, "upload complete. Syncing file changes.")
				} else {
					log.Println("cyan", "Service", service.Name, "commands started. To sync file changes add a path.")
				}

				// If all the services has uploaded and none have paths exit.
				if len(hasUploaded) == len(servicesUploading) {
					hasPaths := false

					for _, service := range servicesUploading {
						config := state.Config[service.Name]

						if config.Path != "" {
							hasPaths = true
							break
						}
					}

					if !hasPaths {
						log.Println("yellow", "Your services have been successfully created.")
						log.Println("yellow", "To initiate file syncing, specify a path for at least one service.")
						done <- 0
					}
				}
			case conn := <-logChan:
				defer conn.Close()
			case ev := <-syncer.Event:
				if connected {
					log.Println("", ev)
					keen.AddEvent("file synced", map[string]interface{}{"user": dev, "event": ev.String()})
				}
			case err = <-syncer.Error:
				rollbar.Report(err)
				done <- 1
			}
		}
	}()

	return <-done
}

func updateOrCreateApp(dev *db.Developer) (*db.State, error) {
	log.Debug("Updating/Creating application.")

	// Get the apps services.
	services, err := db.GetServices()
	if err != nil {
		return nil, err
	}

	// Add a service if none exist.
	if len(services.Data) <= 0 {
		log.Println("yellow", "At least one service is required to connect, in the future run `bowery add` to add services.\n")

		err = addServices(services)
		if err != nil {
			return nil, err
		}
	}

	// Get the apps current state, and update token/config.
	state, err := db.GetState()
	if err != nil && err != errors.ErrNotConnected {
		return nil, err
	}
	state.Token = dev.Token
	state.Config = services.Data

	// Connect and retrieve new state.
	err = api.Connect(state)
	if err != nil {
		return nil, err
	}

	// Write new services.
	services.Data = state.Config
	err = services.Save()
	if err != nil {
		return nil, err
	}

	log.Debug("Updated application", state)
	err = state.Save()
	if err != nil {
		return nil, err
	}

	return state, nil
}

func initiateSync(state *db.State) (*sync.Syncer, []*schemas.Service, error) {
	syncer := sync.NewSyncer()
	services := make([]*schemas.Service, 0)

	log.Debug("Intiating file sync.")
	log.Println("", "Services are availble in the forms:")
	log.Println("", "  <name>.<id>.boweryapps.com")
	log.Println("", "  <port>.<name>.<id>.boweryapps.com")
	log.Println("magenta", "Check \"bowery logs\" for service details.")

	printService := func(service *schemas.Service) {
		log.Println("magenta", "Service", service.Name, "is available at:")
		appIdentifier := state.App.ID
		if state.App.Name != "" {
			appIdentifier = state.App.Name
		}

		url := "80: " + service.PublicAddr
		if ENV != "development" {
			url = service.Name + "." + appIdentifier + ".boweryapps.com"
		}
		log.Println("magenta", " ", url)

		for port, addr := range service.CustomPorts {
			url = port + ": " + addr
			if ENV != "development" {
				url = port + "." + service.Name + "." + appIdentifier + ".boweryapps.com"
			}

			log.Println("magenta", " ", url)
		}
	}

	// Get the list of services that need to upload and/or sync.
	for _, service := range state.App.Services {
		config := state.Config[service.Name]

		// Ensure the path exists and is a directory.
		if config.Path != "" {
			localPath := strings.Split(config.Path, ":")[0]
			info, err := os.Lstat(localPath)
			if err != nil {
				if os.IsNotExist(err) {
					return nil, nil, errors.Newf(errors.ErrPathNotFoundTmpl, service.Name, localPath)
				}

				return nil, nil, err
			}

			if !info.IsDir() {
				return nil, nil, errors.Newf(errors.ErrPathNotDirTmpl, service.Name, localPath)
			}
		}

		services = append(services, service)
	}

	// Check if the services are running and upload/sync them.
	for _, service := range services {
		config := state.Config[service.Name]
		log.Debug("Starting upload for", service.Name, "possibly syncing to", service.SatelliteAddr)

		// Ensure satellite is running for the service.
		var err error
		i := 0
		for i < 1000 {
			err = delancey.CheckHealth(service.SatelliteAddr)
			if err == nil {
				break
			}

			i++
			<-time.After(8 * time.Millisecond)
		}
		if err != nil {
			syncer.Close()
			return nil, nil, errors.NewStackError(errors.ErrContainerConnect)
		}

		syncer.Watch(strings.Split(config.Path, ":")[0], service)
		if config.Path != "" {
			log.Println("cyan", "Uploading file changes and running commands for", service.Name+".")
		} else {
			log.Println("cyan", "No path, only running commands for", service.Name+".")
		}
		printService(service)
	}

	return syncer, services, nil
}

func tailLogs(state *db.State, logChan chan redis.Conn) (*syncWriter, error) {
	log.Debug("Tailing logs.")

	file, err := os.OpenFile(filepath.Join(".bowery", "output.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	output := &syncWriter{File: file}

	// Connect and write logs to file, ignore errors because syncing
	// shouldn't depend on logs.
	go func() {
		var (
			conn redis.Conn
			err  error
		)
		i := 0

		// Attempt to connect.
		for i < 1000 {
			conn, err = redis.Dial("tcp", api.RedisPath)
			if err == nil {
				logChan <- conn
				break
			}

			i++
			<-time.After(time.Millisecond * 50)
		}

		// No successful connection so just forget it.
		if conn == nil {
			log.Debug("Couldn't connect to Redis", api.RedisPath)
			return
		}
		pubsub := redis.PubSubConn{Conn: conn}
		log.Debug("Connected to Redis", api.RedisPath)

		write := func(data []byte) error {
			buf := bytes.NewBuffer(data)

			_, err := io.Copy(output, buf)
			if err != nil {
				err = errors.NewStackError(err)
			}

			return err
		}

		err = pubsub.Subscribe("logs:" + state.App.ID)
		if err != nil {
			return
		}

		for {
			switch res := pubsub.Receive().(type) {
			case redis.Message:
				err = write(res.Data)
				if err != nil {
					return
				}
			case redis.PMessage:
				err = write(res.Data)
				if err != nil {
					return
				}
			case error:
				return
			}
		}
	}()

	return output, nil
}

func apiStatus(token string) {
	log.Debug("Starting api status pings.")

	for {
		<-time.After(5 * time.Second)
		err := broome.DevPing(token)
		if err != nil && connected {
			connected = false
			log.Fprintln(os.Stderr, "red", errors.ErrCantConnect, "Attempting to re-connect...")
		}

		if err == nil && !connected {
			connected = true
			log.Println("magenta", "And we're back in action. Sorry for the inconvenience.")
		}
	}
}

// syncWriter syncs to the fs after writes.
type syncWriter struct {
	File  *os.File
	mutex mutex.Mutex
}

// Write writes the given buffer and syncs to the fs.
func (sw *syncWriter) Write(b []byte) (int, error) {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	n, err := sw.File.Write(b)
	if err != nil {
		return n, err
	}

	return n, sw.File.Sync()
}

// Close closes the writer after any writes have completed.
func (sw *syncWriter) Close() error {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	return sw.File.Close()
}
