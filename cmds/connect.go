// Copyright 2013-2014 Bowery, Inc.
package cmds

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Bowery/bowery/api"
	"github.com/Bowery/bowery/db"
	"github.com/Bowery/bowery/errors"
	"github.com/Bowery/bowery/log"
	"github.com/Bowery/bowery/requests"
	"github.com/Bowery/bowery/rollbar"
	"github.com/Bowery/bowery/sync"
	"github.com/Bowery/bowery/version"
	"github.com/Bowery/gopackages/keen"
	"github.com/Bowery/gopackages/schemas"

	"github.com/garyburd/redigo/redis"
)

var (
	ENV       = os.Getenv("ENV")
	connected = true
)

func init() {
	Cmds["connect"] = &Cmd{connectRun, "connect", "Bootup the app in the current directory.", ""}
}

func connectRun(keen *keen.Client, rollbar *rollbar.Client, args ...string) int {
	// Set the ulimit on Darwin.
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("ulimit", "-Hn", "65535")
		cmd.Run()
	}

	// Create and register signals.
	signals := make(chan os.Signal, 1)
	done := make(chan int, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	defer signal.Stop(signals)

	ver, err := requests.GetVersion()
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	if VersionOutOfDate(version.Version, ver) {
		log.Fprintln(os.Stderr, "red", errors.ErrOutOfDate)
		return 1
	}

	dev, err := db.GetDeveloper()
	if err != nil && err != errors.ErrNoDeveloper {
		rollbar.Report(err)
		return 1
	}

	// If there's no developer get a token.
	if dev.Token == "" {
		log.Println("yellow", "Oops! You must be logged in.")

		err = getToken(dev)
		if err != nil {
			rollbar.Report(err)
			return 1
		}
	}

	err = getDeveloper(dev)
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	log.Println("magenta", "Hey there,", strings.Split(dev.Developer.Name, " ")[0]+
		". Connecting you to Bowery now...")

	state, err := updateOrCreateApp(dev)
	if err != nil {
		if err == errors.ErrNoServicePaths {
			log.Println("yellow", "Your services have been successfully created.")
			log.Println("yellow", "To initiate file syncing, specify a path.")
			return 0
		}
		rollbar.Report(err)
		return 1
	}

	syncer, err := initiateSync(state)
	if syncer != nil {
		defer syncer.Close()
	}

	if err != nil {
		if err == errors.ErrNoServicePaths {
			log.Println("yellow", "Your services have been successfully created.")
			log.Println("yellow", "To initiate file syncing, specify a path.")
			return 0
		}

		rollbar.Report(err)
		return 1
	}

	logChan := make(chan redis.Conn, 1)
	logFile, err := tailLogs(state, logChan)
	if logFile != nil {
		defer logFile.Close()
	}
	if err != nil {
		rollbar.Report(err)
		return 1
	}

	go func() {
		<-signals
		requests.Disconnect(dev.Token)
		done <- 0
	}()

	// If a service takes more than 10 seconds to upload, alert the developer
	// that some services take a while to upload.
	hasUploaded := make(map[string]bool)
	for _, s := range state.App.Services {
		if state.Config[s.Name].Path != "" {
			hasUploaded[s.Name] = false
		}
	}

	go func() {
		<-time.After(20 * time.Second)
		hasNotUploaded := []string{}
		for service, isUploaded := range hasUploaded {
			if !isUploaded {
				hasNotUploaded = append(hasNotUploaded, service)
			}
		}

		services := strings.Join(hasNotUploaded, ", ")

		if len(hasNotUploaded) > 0 {
			if len(hasNotUploaded) == 1 {
				log.Println("yellow", "Service", services, "is still uploading...")
			} else {
				log.Println("yellow", "Services", services, "are still uploading...")
			}
			log.Println("yellow", "Large applications and inital uploads can take some time to complete.")
			log.Println("yellow", "Check `bowery logs` to see the current status of the build.")
		}
	}()

	go apiStatus(dev.Token)

	keen.AddEvent("bowery connect", map[string]interface{}{
		"user": dev,
		"app":  state,
	})

	// Watch for various events.
	go func() {
		for {
			select {
			case service := <-syncer.Upload:
				hasUploaded[service.Name] = true
				log.Println("cyan", "Service", service.Name, "upload complete. Syncing file changes.")
			case conn := <-logChan:
				defer conn.Close()
			case ev := <-syncer.Event:
				if connected {
					keen.AddEvent("file synced", map[string]interface{}{"user": dev, "event": ev.String()})
					log.Println("", ev)
				}
			case err = <-syncer.Error:
				rollbar.Report(err)
				done <- 0
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

	// Get the apps current state, and update token/config.
	state, err := db.GetState()
	if err != nil && err != errors.ErrNotConnected {
		return nil, err
	}
	state.Token = dev.Token
	state.Config = services.Data

	// Connect, and retrieve new state.
	err = requests.Connect(state)
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

	if len(services.Data) <= 0 {
		return nil, errors.ErrNoServices
	}

	return state, nil
}

func initiateSync(state *db.State) (*sync.Syncer, error) {
	log.Debug("Intiating file sync.")
	syncs := make(map[string][]*schemas.Service)
	hasPaths := false

	log.Println("", "Services are availble in the forms:")
	log.Println("", "  <name>.<id>.boweryapps.com")
	log.Println("", "  <port>.<name>.<id>.boweryapps.com")

	printService := func(service *schemas.Service) {
		log.Println("magenta", "Service", service.Name, "is available at:")

		url := "80: " + service.PublicAddr
		appIdentifier := state.App.ID
		if state.App.Name != "" {
			appIdentifier = state.App.Name
		}
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

	// Get the service addrs that need syncing.
	for _, service := range state.App.Services {
		config := state.Config[service.Name]
		// Even if no path is given, ensure container is up
		i := 0

		var err error

		// Attempt to connect.
		for i < 1000 {
			<-time.After(8 * time.Millisecond)
			err = requests.SatelliteCheckHealth(service.SatelliteAddr)

			if err == nil {
				break
			}

			i++
		}

		if err != nil {
			return nil, errors.ErrContainerConnect
		}

		// No reason to sync if no path is given.
		if config.Path == "" {
			log.Println("yellow", "Skipping file changes for", service.Name+", no path given.")
			printService(service)
			continue
		}

		// Ensure the path is a directory.
		info, err := os.Lstat(config.Path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, errors.Newf(errors.ErrPathNotFoundTmpl, service.Name, config.Path)
			}

			return nil, err
		}
		if !info.IsDir() {
			log.Println("yellow", "Skipping file changes for", service.Name+", path is not a directory.")
			printService(service)
			continue
		}

		// Add the current service to it's sync path.
		serv, ok := syncs[config.Path]
		if !ok || serv == nil {
			syncs[config.Path] = make([]*schemas.Service, 0)
			serv = syncs[config.Path]
		}
		syncs[config.Path] = append(serv, service)
		hasPaths = true
	}

	if !hasPaths {
		return nil, errors.ErrNoServicePaths
	}
	syncer := sync.NewSyncer()

	// Start syncing all directories, if an error occurs return error with
	// syncer so it can close the existing watchers.
	for path, services := range syncs {
		for _, service := range services {
			log.Debug("Watching", path, "and syncing to", service.SatelliteAddr)

			err := syncer.Watch(path, service)
			if err != nil {
				return syncer, err
			}

			log.Println("cyan", "Uploading file changes to", service.Name+", check \"bowery logs\" for details.")
			printService(service)
		}
	}

	return syncer, nil
}

func tailLogs(state *db.State, logChan chan redis.Conn) (*os.File, error) {
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
		var conn redis.Conn
		var err error
		i := 0

		// Attempt to connect.
		for i < 1000 {
			if err != nil {
				<-time.After(time.Millisecond * 50)
			}

			conn, err = redis.Dial("tcp", api.RedisPath)
			if err == nil {
				logChan <- conn
				break
			}

			i++
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

	return file, nil
}

func apiStatus(token string) {
	log.Debug("Starting api status pings.")

	for {
		<-time.After(5 * time.Second)
		err := requests.DevPing(token)
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
	File *os.File
}

// Write writes the given buffer and syncs to the fs.
func (sw *syncWriter) Write(b []byte) (int, error) {
	n, err := sw.File.Write(b)
	if err != nil {
		return n, err
	}

	return n, sw.File.Sync()
}
