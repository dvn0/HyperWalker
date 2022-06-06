package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

import (
	marionette "github.com/njasm/marionette_client"
	"github.com/thedevsaddam/gojsonq"
	"golang.org/x/net/html"
)

var client = marionette.NewClient()
var userHome = getUserHomePath()

func usage() {
	fmt.Fprintln(os.Stderr, "hyperwalker -- fetch and freeze-dry webpages\nOptions:")
	flag.PrintDefaults()
}

func main() {
	logFile, err := os.OpenFile(userHome+"/.hyperwalker/logs/hyperwalker.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("HyperWalker is running")
	spawnFf()
	// Will block if we don't run concurrently
	go serveScript()

	var (
		urlString = flag.String("url", "", "URL to fetch")
	)
	flag.Usage = usage
	flag.Parse()

	if *urlString != "" {
		fmt.Printf("Retrieving %s\n", *urlString)
		initClient(*urlString)
		return
	} else {
		fmt.Println("URL required.")
		return
	}
}

// TODO: Check whether Firefox with marrionette is already running
func spawnFf() {
	// Check if Firefox profile exists:

	checkProfile, err := filepath.Glob(userHome + "/.mozilla/firefox/*.hyperwalker")
	if err != nil {
		log.Fatalf(err.Error())
	}
	if len(checkProfile) == 0 {
		// Create a new profile:
		log.Printf("Creating new Firefox profile in: %s", userHome+"/.mozilla/firefox/")
		ffProfile := exec.Command("firefox", "--headless", "--CreateProfile", "hyperwalker")
		ffProfile.Start()
		// Execute firefox with these arguments
		// TODO: make headless mode toggled by debug flag
		log.Println("Opening a headless Firefox.")
		time.Sleep(2 * time.Second) // Sleep to wait for profile creation to complete
		ffCmd := exec.Command("firefox", "--marionette", "--headless", "--private-window", "-P", "hyperwalker")
		ffCmd.Start()
	} else {
		// Execute firefox with these arguments:
		// TODO: make headless mode toggled by debug flag
		log.Println("Opening a headless Firefox.")
		ffCmd := exec.Command("firefox", "--marionette", "--headless", "--private-window", "-P", "hyperwalker")
		ffCmd.Start()
	}
}

func raw_connect(host string, port string) error {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if conn != nil {
		defer conn.Close()
		log.Println("Connected to", net.JoinHostPort(host, port))
	}
	return err
}

func initClient(uri string) {
	for connectStatus := raw_connect("127.0.0.1", "2828"); connectStatus != nil; time.Sleep(1 * time.Second) {
		connectStatus := raw_connect("127.0.0.1", "2828")
		if connectStatus != nil {
			log.Println("Waiting for Firefox to ready the marionette server...")
		} else {
			break
		}
	}
	client.Connect("127.0.0.1", 2828) // this are the default marionette values for hostname, and port
	log.Println("Marrionette connected.")
	client.NewSession("", nil) // let marionette generate the Session ID with its default Capabilities
	log.Println("New Marionette session generated.")
	client.Navigate(uri)
	log.Printf("Navigating to %s", uri)
	execute()
	client.Navigate("about:newtab")
	quit()
}

// We have to serve the JS scripts via HTTP

//go:embed js
var embededFiles embed.FS

func cors(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// do your cors stuff
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// return if you do not want the FileServer handle a specific request

		fs.ServeHTTP(w, r)
	}
}

func serveScript() {
	fsys, err := fs.Sub(embededFiles, ".")
	if err != nil {
		panic(err)
	}

	http.Handle("/js/", cors(http.FileServer(http.FS(fsys))))
	if err := http.ListenAndServe("127.0.0.1:61628", nil); err != nil {
		log.Fatalf("Problem starting HTTP server. Go says: ", err)
	}
}

func execute() (string, string) {
	resp, err := http.Get("http://127.0.0.1:61628/js/dist/freeze-dry.umd.js")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Something bad", err)
	}

	scriptstr := fmt.Sprintf("console.log('Here comes the big script...');\n%s\nconsole.log('Running freezeDry()'); return freezeDry.default();", string(bodyBytes))
	args := []interface{}{}     // arguments to be passed to the function
	timeoutint := 50000         // milliseconds
	timeout := uint(timeoutint) // milliseconds
	sandbox := false            // new Sandbox
	snap, err := client.ExecuteScript(scriptstr, args, timeout, sandbox)
	if err != nil {
		log.Fatal("Error executing script", err)
	}

	// HTML is returned as a JSON blob, so we must parse it
	data := gojsonq.New().JSONString(snap.Value).Find("value")
	//println(data.(string))

	// Get title
	reader := strings.NewReader(data.(string))
	title, true := GetHtmlTitle(reader)
	if true {
		fmt.Println("Saving Page: " + title)
	}
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	sanTitle := reg.ReplaceAllString(title, "-")
	/*; true {
		fmt.Println(title)
	} else {
		fmt.Println("Fail to get HTML title"); log.Printf("Fail to get HTML title")
	}
	*/
	f, err := ioutil.TempFile(os.TempDir(), sanTitle+"-hyperwalker-*.html")
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

// For saving a screenshot.
// TODO: Save a screenshot with every HTML snapshot. Make sure to run concurrently.
func screenshot() {
	screenshot, err := client.Screenshot()
	if err != nil {
		log.Fatal("wtf", err)
	}
	fmt.Println(screenshot)
}

func getUserHomePath() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	return user.HomeDir
}

// This makes Firefox quit
func quit() {
	client.Quit()
}
