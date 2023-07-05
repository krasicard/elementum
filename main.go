package main

import (
	_ "github.com/anacrolix/envpprof"

	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/anacrolix/sync"
	"github.com/anacrolix/tagflag"
	"github.com/op/go-logging"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/elgatito/elementum/api"
	"github.com/elgatito/elementum/bittorrent"
	"github.com/elgatito/elementum/broadcast"
	"github.com/elgatito/elementum/config"
	"github.com/elgatito/elementum/database"
	"github.com/elgatito/elementum/exit"
	"github.com/elgatito/elementum/library"
	"github.com/elgatito/elementum/lockfile"
	"github.com/elgatito/elementum/scrape"
	"github.com/elgatito/elementum/trakt"
	"github.com/elgatito/elementum/util"
	"github.com/elgatito/elementum/xbmc"
)

var (
	log     = logging.MustGetLogger("main")
	logPath = ""
)

func init() {
	sync.Enable()
}

func ensureSingleInstance(conf *config.Configuration) (lock *lockfile.LockFile, err error) {
	// Avoid killing any process when running as a shared library
	if exit.IsShared {
		return
	}

	file := filepath.Join(conf.Info.Profile, ".lockfile")
	lock, err = lockfile.New(file)
	if err != nil {
		log.Critical("Unable to initialize lockfile:", err)
		return
	}
	var pid int
	var p *os.Process
	pid, err = lock.Lock()
	if pid <= 0 {
		if err = os.Remove(lock.File); err != nil {
			log.Critical("Unable to remove lockfile")
			return
		}
		_, err = lock.Lock()
	} else if err != nil {
		log.Warningf("Unable to acquire lock %q: %v, killing...", lock.File, err)
		p, err = os.FindProcess(pid)
		if err != nil {
			log.Warning("Unable to find other process:", err)
			return
		}
		if err = p.Kill(); err != nil {
			log.Critical("Unable to kill other process:", err)
			return
		}
		if err = os.Remove(lock.File); err != nil {
			log.Critical("Unable to remove lockfile")
			return
		}
		_, err = lock.Lock()
	}
	return
}

func setupLogging() {
	var backend *logging.LogBackend

	if config.Args.LogPath != "" {
		logPath = config.Args.LogPath
	}
	if logPath != "" && config.IsWritablePath(filepath.Base(logPath)) == nil {
		backend = logging.NewLogBackend(&lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    10, // Size in Megabytes
			MaxBackups: 5,
		}, "", 0)
	} else {
		backend = logging.NewLogBackend(os.Stdout, "", 0)
	}

	logging.SetFormatter(logging.MustStringFormatter(
		`%{color}%{level:.4s}  %{module:-12s} ▶ %{shortfunc:-15s}  %{color:reset}%{message}`,
	))
	logging.SetBackend(logging.NewLogBackend(ioutil.Discard, "", 0), backend)
}

func main() {
	now := time.Now()

	// If running in shared library mode, parse Args from variable, provided by library caller.
	if !exit.IsShared || exit.Args == "" {
		tagflag.Parse(&config.Args)
	} else {
		if err := tagflag.ParseErr(&config.Args, strings.Fields(exit.Args)); err != nil {
			fmt.Printf("Error parsing CLI arguments: %s", err)
			exit.Exit(exit.ExitCodeError)
			return
		}
	}

	// Make sure we are properly multithreaded.
	runtime.GOMAXPROCS(runtime.NumCPU())

	setupLogging()

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Got a panic: %s", r)
			log.Errorf("Stacktrace: \n" + string(debug.Stack()))
			exit.Exit(exit.ExitCodeError)
		}
	}()

	if exit.IsShared {
		log.Infof("Starting Elementum daemon in shared library mode")
	} else {
		log.Infof("Starting Elementum daemon")
	}
	log.Infof("Version: %s LibTorrent: %s Go: %s, Threads: %d", util.GetVersion(), util.GetTorrentVersion(), runtime.Version(), runtime.GOMAXPROCS(0))

	// Init default XBMC connections
	xbmc.Init()

	conf, err := config.Reload()
	if err != nil || conf == nil {
		log.Errorf("Could not get addon configuration: %s", err)
		exit.Exit(exit.ExitCodeError)
		return
	}

	xbmc.KodiVersion = conf.Platform.Kodi

	log.Infof("Addon: %s v%s", conf.Info.ID, conf.Info.Version)

	lock, err := ensureSingleInstance(conf)
	if err != nil {
		log.Warningf("Unable to acquire lock %q: %v, exiting...", lock.File, err)
		exit.Exit(exit.ExitCodeError)
	}
	if lock != nil {
		defer lock.Unlock()
	}

	db, err := database.InitStormDB(conf)
	if err != nil {
		log.Errorf("Could not open application database: %s", err)
		exit.Exit(exit.ExitCodeError)
		return
	}

	cacheDb, errCache := database.InitCacheDB(conf)
	if errCache != nil {
		log.Errorf("Could not open cache database: %s", errCache)
		exit.Exit(exit.ExitCodeError)
		return
	}

	s := bittorrent.NewService()

	var shutdown = func(code int) {
		if s == nil || s.Closer.IsSet() {
			return
		}

		// Set global Closer
		broadcast.Closer.Set()

		s.Closer.Set()

		log.Infof("Shutting down with code %d ...", code)
		scrape.Stop()
		library.CloseLibrary()
		s.Close(true)

		db.Close()
		cacheDb.Close()

		// Wait until service is finally stopped
		<-s.CloserNotifier.C()

		log.Info("Goodbye")

		// If we don't give an exit code - python treat as well done and not
		// restarting the daemon. So when we come here from Signal -
		// we should properly exit with non-0 exitcode.
		exit.Exit(code)
	}

	var watchParentProcess = func() {
		for {
			if os.Getppid() == 1 {
				log.Warning("Parent shut down, shutting down too...")
				go shutdown(exit.ExitCodeSuccess)
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
	go watchParentProcess()

	// Make sure HTTP mux is empty
	http.DefaultServeMux = new(http.ServeMux)
	http.Handle("/", api.Routes(s, shutdown))

	http.Handle("/files/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		handler := http.StripPrefix("/files/", http.FileServer(bittorrent.NewTorrentFS(s, r.Method)))
		handler.ServeHTTP(w, r)
	}))

	if config.Get().GreetingEnabled {
		if xbmcHost, _ := xbmc.GetLocalXBMCHost(); xbmcHost != nil {
			xbmcHost.Notify("Elementum", "LOCALIZE[30208]", config.AddonIcon())
		}
	}

	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	signal.Ignore(syscall.SIGPIPE, syscall.SIGILL)
	defer close(sigc)

	go func() {
		closer := s.Closer.C()

		for {
			select {
			case <-closer:
				return
			case <-exit.Closer.C():
				shutdown(exit.ExitCodeSuccess)
			case <-sigc:
				shutdown(exit.ExitCodeError)
			}
		}
	}()

	go func() {
		xbmcHost, _ := xbmc.GetLocalXBMCHost()
		if xbmcHost == nil || !xbmcHost.Ping() {
			return
		}

		if checkRepository(xbmcHost) {
			log.Info("Updating Kodi add-on repositories... ")
			xbmcHost.UpdateAddonRepos()
		}

		xbmcHost.DialogProgressBGCleanup()
		xbmcHost.ResetRPC()
	}()

	go library.Init()
	go trakt.TokenRefreshHandler()
	go db.MaintenanceRefreshHandler()
	go cacheDb.MaintenanceRefreshHandler()
	go scrape.Start()
	go util.FreeMemoryGC()

	localAddress := fmt.Sprintf("%s:%d", config.Args.LocalHost, config.Args.LocalPort)
	log.Infof("Prepared in %s", time.Since(now))
	log.Infof("Starting HTTP server at %s", localAddress)

	exit.Server = &http.Server{
		Addr:    localAddress,
		Handler: nil,
	}

	if err = exit.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		exit.Panic(err)
		return
	}
	if !exit.IsShared {
		os.Exit(exit.Code)
	}
}
