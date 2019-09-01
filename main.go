//go:generate statik -src=./js

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

import (
	"golang.org/x/net/html"
	marionette "github.com/njasm/marionette_client"
	webview "github.com/zserge/webview"
	"github.com/thedevsaddam/gojsonq"
	"github.com/mitchellh/go-homedir"
	"github.com/rakyll/statik/fs"
	"github.com/kennygrant/sanitize"
	_ "./statik" // TODO: Replace with the absolute import path
)

var client = marionette.NewClient()
var userHome, err = homedir.Dir()


func main() {
	logFile, err := os.OpenFile(userHome + "/.hyperwalker/logs/hyperwalker.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("HyperWalker is running")
	go spawnFf()
	// Will block if we don't run concurrently
	go serveScript()
	initClient()
	openView()
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

func continuedNav(uri string) {
	client.Navigate(uri)
	openView()
}

// We have to serve the JS scripts via HTTP
func serveScript() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(statikFS)))
	if err := http.ListenAndServe("127.0.0.1:61628", nil); err != nil {
		log.Fatalf("Problem starting HTTP server. Go says: ", err)
	}
}


func execute() (string, string) {
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

	// Get title
	reader := strings.NewReader(data.(string))
	title, true := GetHtmlTitle(reader);
	if true{
		fmt.Println("Displaying Page: " + title)
	}
	sanTitle := sanitize.Path(title)
	/*; true {
		fmt.Println(title)
	} else {
		fmt.Println("Fail to get HTML title"); log.Printf("Fail to get HTML title")
	}
*/
	f, err := ioutil.TempFile(os.TempDir(), sanTitle + "-hyperwalker-*.html")
	if err != nil {
	    log.Fatal("Cannot create temporary file", err)
	}
	defer f.Close()
	if _, err := f.WriteString(data.(string)); err != nil {
	    log.Fatal("Cannot write to temporary file", err)
	}
	fileName := f.Name()
    fmt.Printf("wrote snapshot to %s\n", fileName)
	f.Sync()
	return fileName, title
}

func handleRPC(w webview.WebView, data string) {
	continuedNav(data)
}

func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title" 
}

func traverse(n *html.Node) (string, bool) {
	if isTitleElement(n) {
		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := traverse(c)
		if ok {
			return result, ok
		}
	}

	return "", false
}

func GetHtmlTitle(r io.Reader) (string, bool) {
	doc, err := html.Parse(r)
	if err != nil {
		panic("Fail to parse html")
	}

	return traverse(doc)
}

func cspMod() (string, string){
	snapshot, title := execute()
	file, err := ioutil.ReadFile(snapshot)
	if err != nil {
		log.Printf("Failed to open snapshot for CSP modification")
	}
	new := strings.Replace(string(file), "none", "unsafe-inline", 1)
	err = ioutil.WriteFile(snapshot, []byte(new), 0600)
	if err != nil {
		log.Printf("Failed to open snapshot for CSP modification")
	}
	return snapshot, title
}

func openView() {

	snapshot, title := cspMod()
	w := webview.New(webview.Settings{
		URL: "file:///" + snapshot,
		Title: title,
		Resizable: true,
		Debug: true,
		ExternalInvokeCallback: handleRPC,
	})
	webview.Debug()
	defer w.Exit()
	w.Dispatch(func() {
		// Inject JS
		bean, err := http.Get("http://127.0.0.1:61628/js/intercept.js")
		if err != nil {
			panic(err)
		}
		defer bean.Body.Close()
		boBytes, err := ioutil.ReadAll(bean.Body)
		if err != nil {
			log.Fatalf("Something bad", err)
		}

		scriptstr := string(boBytes)
		w.Eval(scriptstr)
	})
	w.Run()
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
