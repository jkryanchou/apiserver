package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/bouk/httprouter"
	"github.com/cloudflare/golibs/lrucache"
	"github.com/phuslu/glog"
	"github.com/phuslu/net/http2"
	"golang.org/x/sync/singleflight"
)

var (
	version = "r9999"
)

func main() {
	var err error

	rand.Seed(time.Now().UnixNano())
	OLDPWD, _ := os.Getwd()

	if len(os.Args) > 1 && os.Args[1] == "-version" {
		fmt.Println(version)
		return
	}

	var pidfile string
	flag.StringVar(&pidfile, "pidfile", "", "pid file name")
	flag.Parse()

	config, err := NewConfig(flag.Arg(0))
	if err != nil {
		glog.Fatalf("NewConfig(%#v) error: %+v", flag.Arg(0), err)
	}

	// see http.DefaultTransport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ClientSessionCache: tls.NewLRUClientSessionCache(2048),
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    false,
	}

	ipinfo := &IpinfoHandler{
		URL:          config.Ipinfo.Url,
		Regex:        regexp.MustCompile(config.Ipinfo.Regex),
		CacheTTL:     time.Duration(config.Ipinfo.CacheTtl) * time.Second,
		Cache:        lrucache.NewLRUCache(10000),
		Singleflight: &singleflight.Group{},
		Transport:    transport,
	}

	ln, err := ReusePortListen("tcp", config.Default.ListenAddr)
	if err != nil {
		glog.Fatalf("TLS Listen(%s) error: %s", config.Default.ListenAddr, err)
	}

	glog.Infof("apiserver %s ListenAndServeTLS on %s\n", version, ln.Addr().String())

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/ipinfo/:ip", ipinfo.Ipinfo)
	router.GET("/metrics", Metrics)
	router.GET("/debug/pprof/profile", Pprof)

	server := &http.Server{
		Handler: router,
	}

	http2.ConfigureServer(server, &http2.Server{})

	go server.Serve(ln)

	if pidfile != "" {
		ioutil.WriteFile(pidfile, []byte(strconv.Itoa(os.Getpid())), 0644)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	switch <-c {
	case syscall.SIGHUP:
	default:
		glog.Infof("apiserver server closed.")
		os.Exit(0)
	}

	glog.Infof("apiserver start new child server")
	exe, err := os.Executable()
	if err != nil {
		glog.Fatalf("os.Executable() error: %+v", exe)
	}

	_, err = os.StartProcess(exe, os.Args, &os.ProcAttr{
		Dir:   OLDPWD,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})
	if err != nil {
		glog.Fatalf("os.StartProcess(%+v, %+v) error: %+v", exe, os.Args, err)
	}

	glog.Warningf("apiserver start graceful shutdown...")
	SetProcessName("apiserver: (graceful shutdown)")

	timeout := 5 * time.Minute
	if config.Default.GracefulTimeout > 0 {
		timeout = time.Duration(config.Default.GracefulTimeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		glog.Errorf("%T.Shutdown() error: %+v", server, err)
	}

	glog.Infof("apiserver server shutdown.")
}
