package main

import (
	marionette "github.com/njasm/marionette_client"
	webview "github.com/zserge/webview"
	"github.com/thedevsaddam/gojsonq"
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var client = marionette.NewClient()

func main() {
	initClient()
	// Will block if we don't run concurrently
	go serveScript()
    execute()
}

// TODO: Check whether Firefox is running, and start it if not
// Start Firefox in marrionette mode like this:
// /path/to/firefox -P marionette -no-remote -headless -marionette -safe-mode
// "-P" is for selecting a profile, it can be called anything
// You may need to go to about:profiles and create one first
// After the profile exists you need to copy the file config/user.js to:
// $HOME/.mozilla/firefox/<profile-name>/user.js
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
	http.Handle("/", http.FileServer(http.Dir("./freezedry")))
	if err := http.ListenAndServe(":5000", nil); err != nil {
		panic(err)
	}
}

// Trying this out...
func check(e error) {
    if e != nil {
        panic(e)
    }
}

func execute() {
	script, err := ioutil.ReadFile("./js/exec.js")
	scriptstr := string(script)
	args := []interface{}{}  // arguments to be passed to the function
	timeoutint := 50000     // milliseconds
	timeout := uint(timeoutint)     // milliseconds
	sandbox := false    // new Sandbox
	snap, err := client.ExecuteScript(scriptstr, args, timeout, sandbox)
	if err != nil {
	    fmt.Println(err)
	}
	// HTML is returned as a JSON blob, so we must parse it
	data := gojsonq.New().JSONString(snap.Value).Find("value")
	//println(data.(string))

	f, err := os.Create("/tmp/dat1.html")
	check(err)
	defer f.Close()
	html, err := f.WriteString(data.(string))
    fmt.Printf("wrote %d bytes to /tmp/dat1.html\n", html)
	f.Sync()
	// Open up the html file
	// TODO: remove the i/o operation and open HTML from memory
	webview.Open("Minimal webview example",
		"file:///tmp/dat1.html", 800, 600, true)
}

// For saving a screenshot.
// TODO: Save a screenshot with every HTML snapshot. Make sure to run concurrently.
func screenshot() {
	screenshot, err := client.Screenshot()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(screenshot)
}

// This makes Firefox quit
func quit() {
	client.Quit()
}
