package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ti/router/tibrain/internal/ticonsole"
)

func main() {
	refresh := flag.Duration("refresh", 3*time.Second, "dashboard refresh interval")
	once := flag.Bool("once", false, "probe once and print JSON")
	noColor := flag.Bool("no-color", false, "disable ANSI colors")
	timeout := flag.Duration("timeout", 2*time.Second, "per-service HTTP timeout")
	flag.Parse()

	clientTimeout := *timeout
	if clientTimeout <= 0 {
		clientTimeout = 2 * time.Second
	}
	backend := ticonsole.NewHTTPBackend(
		&http.Client{Timeout: clientTimeout},
		ticonsole.DefaultServices(),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *once {
		probeCtx, probeCancel := context.WithTimeout(ctx, clientTimeout+time.Second)
		defer probeCancel()
		snapshot := backend.Snapshot(probeCtx)
		encoded, err := json.MarshalIndent(snapshot, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "ti-console:", err)
			os.Exit(1)
		}
		fmt.Println(string(encoded))
		return
	}

	app := &ticonsole.App{
		Backend:         backend,
		In:              os.Stdin,
		Out:             os.Stdout,
		RefreshInterval: *refresh,
		UseColor:        !*noColor,
	}
	if err := app.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "ti-console:", err)
		os.Exit(1)
	}
}
