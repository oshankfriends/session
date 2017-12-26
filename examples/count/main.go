package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oshankfriends/session"
	_ "github.com/oshankfriends/session/plugins/memory"
)

var (
	port                    = flag.String("port", "8080", "app will listen on this port")
	globalSessionManager, _ = session.NewManager("gosessionid", "memory", time.Second*10)
)

func count(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(os.Stdout, "enters count handler")
	session, _ := globalSessionManager.StartSession(w, r)
	ct := session.Get("countnum")
	if ct == nil {
		session.Set("countnum", 1)
	} else {
		session.Set("countnum", (ct.(int) + 1))
	}
	fmt.Fprintln(w, session.Get("countnum"))
}

func main() {
	flag.Parse()
	var errCh = make(chan error)
	http.HandleFunc("/count", count)
	go func() {
		fmt.Printf("started listening on %s\n", *port)
		errCh <- http.ListenAndServe(":"+*port, nil)
	}()

	go func() {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		errCh <- fmt.Errorf("%s", <-sigChan)
	}()

	fmt.Printf("\nexit : [ %s ]\n", <-errCh)
}
func init() {
	go globalSessionManager.GC()
}
