package main

import (
	marionette "github.com/njasm/marionette_client"
	webview "github.com/zserge/webview"
	"github.com/thedevsaddam/gojsonq"
	"github.com/mitchellh/go-homedir"
	"github.com/rakyll/statik/fs"
	_ "./statik" // TODO: Replace with the absolute import path
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var client = marionette.NewClient()
var userHome, err = homedir.Dir()


func main() {
	fmt.Println(userHome)
	logFile, err := os.OpenFile(userHome + "/.hyperwalker/logs/hyperwalker.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("HyperWalker is running")
	go spawnFf()
	go serveScript()
	initClient()
	// Will block if we don't run concurrently
    execute()
	quit()
}

// TODO: Check whether Firefox with marrionette is already running
func spawnFf() {
	ffProfile := exec.Command("firefox", "-no-remote", "-CreateProfile","hyperwalker")
	ffProfile.Start()
	ffCmd := exec.Command("firefox", "-P", "hyperwalker", "-no-remote", "-headless"," -private-window", "-marionette")
	ffCmd.Start()
}

func initClient() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Navigate to (URI): ")
	uri, _ := reader.ReadString('\n')
	client.Connect("127.0.0.1", 2828) // this are the default marionette values for hostname, and port 
	client.NewSession("", nil) // let marionette generate the Session ID with it's default Capabilities
	client.Navigate(uri)
}

// We have to serve the JS scripts via HTTP
// TODO: Pick a better port
// TODO: Figure out how to embed the files in the binary
func serveScript() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(statikFS)))
	if err := http.ListenAndServe(":61628", nil); err != nil {
		log.Fatalf("Problem starting HTTP server. Go says: ", err)
	}
}


func execute() {
	resp, err := http.Get("http://127.0.0.1:61628/js/exec.js")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Something bad", err)
	}

	scriptstr := string(bodyBytes)
	args := []interface{}{}  // arguments to be passed to the function
	timeoutint := 50000     // milliseconds
	timeout := uint(timeoutint)     // milliseconds
	sandbox := false    // new Sandbox
	snap, err := client.ExecuteScript(scriptstr, args, timeout, sandbox)
	if err != nil {
		log.Fatal("Error executing script", err)
	}
	// HTML is returned as a JSON blob, so we must parse it
	data := gojsonq.New().JSONString(snap.Value).Find("value")
	//println(data.(string))

	f, err := ioutil.TempFile(os.TempDir(), "hyperwalker-*.html")
	if err != nil {
	    log.Fatal("Cannot create temporary file", err)
	}
	defer f.Close()
	if _, err := f.WriteString(data.(string)); err != nil {
	    log.Fatal("Cannot write to temporary file", err)
	}
    fmt.Printf("wrote snapshot to %s\n", f.Name())
	f.Sync()
	// Open up the html file
	// TODO: remove the i/o operation and open HTML from memory
	webview.Open("Minimal webview example",
		"file:///" + f.Name(), 800, 600, true)
}

// For saving a screenshot.
// TODO: Save a screenshot with every HTML snapshot. Make sure to run concurrently.
func screenshot() {
	screenshot, err := client.Screenshot()
	if err != nil {
		log.Fatal("wtf", err)
	}
	fmt.Println(screenshot)
}

// This makes Firefox quit
func quit() {
	client.Quit()
}
