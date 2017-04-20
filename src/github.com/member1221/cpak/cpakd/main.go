package main

import (
	"net/http"
	"github.com/member1221/cpak/libcpak"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"sync"
	"strings"
	"strconv"
)

var CONFIGDIR string = libcpak.CPAK_DEFAULT_CFG_DIR
var DEFAULT_SERVEDIR string = libcpak.CPAK_DEFAULT_SERVE_DIR
var DEFAULT_NAME string = libcpak.CPAK_DEFAULT_REPO_NAME
var DEFAULT_PORT int = libcpak.CPAK_DEFAULT_SERVE_PORT
var DEFAULT_CATEGORIES []string = libcpak.CPAK_DEFAULT_REPO_CATS()

var running bool = true

var config ServerConfig

var CATEGORIES map[string][]libcpak.WebPackage = make(map[string][]libcpak.WebPackage)
var APPS map[string]libcpak.WebPackage = make(map[string]libcpak.WebPackage)

var MTX sync.Mutex



func main() {

	fmt.Println("[CPAK PAcKager Server]")
	fmt.Println( "Loading configuration...")
	file, err := ioutil.ReadFile(CONFIGDIR+"/server.json")

	if err != nil {
		config = ServerConfig{DEFAULT_NAME, DEFAULT_SERVEDIR, DEFAULT_CATEGORIES}
	} else {
		err := json.Unmarshal([]byte(file), config)
		if err != nil {
			fmt.Println("[Json Unmarshalling Error] " + err.Error() + "!")
		}
	}

	go func() {

		libcpak.PushLog(0, "Updating package list...")
		updatepkglist()

		libcpak.PushLog(0, "Added repolist handler...")
		http.HandleFunc("/cpak/repolist", handlerepolist)

		libcpak.PushLog(0, "Added repo handler...")
		http.HandleFunc("/cpak/repo/", handlerepo)

		port := strconv.Itoa(DEFAULT_PORT)
		libcpak.PushLog(0, "Serving on port " + port + "...")
		http.ListenAndServe(":" + port, nil)
	}()
	for running {
		if libcpak.LogLen(0) > 0 {
			fmt.Println(libcpak.PullLog(0))
		}
	}
}

func updatepkglist (){
	folders, err := ioutil.ReadDir(DEFAULT_SERVEDIR)
	if err != nil {

		libcpak.PushLog(0, "[Error] " + err.Error())
	}
	for _, folder := range folders {
		if folder.IsDir() {
			libcpak.PushLog(0, "Added category directory " + folder.Name() + "...")
			MTX.Lock()
			CATEGORIES[folder.Name()] = make([]libcpak.WebPackage, 0)
			MTX.Unlock()
			folders, err := ioutil.ReadDir(DEFAULT_SERVEDIR + "/" + folder.Name())
			if err != nil {
				libcpak.PushLog(0, "[Error] " + err.Error())
			}
			for _, appf := range folders {
				app := libcpak.WebPackage{
					appf.Name(),
					folder.Name() + "/" + appf.Name(),
					folder.Name(),
					"",
				}
				libcpak.PushLog(0, "Added package " + appf.Name() + " from category " + folder.Name() + "...")
				MTX.Lock()
				CATEGORIES[folder.Name()] = append(CATEGORIES[folder.Name()], app)
				APPS[appf.Name()] = app
				MTX.Unlock()
			}

		}
	}
}

func handlerepolist(w http.ResponseWriter, r *http.Request) {
	libcpak.PushLog(0, "Connection from " + r.Host + "...")
	var output []byte
	var err error
	pack := r.FormValue("pkg")
	if strings.TrimSpace(pack) != "" {
		app := APPS[pack]
		if app.Name != "" {
			output, err = json.Marshal(app)
			if err != nil {
				libcpak.PushLog(0, "[Error (Package Listing)] "+err.Error())
			}
		} else {
			output, err = json.Marshal(libcpak.WebError{
				"Package " + pack + " not found!",
				404,
			})
			if err != nil {
				libcpak.PushLog(0, "[Error (Package Listing)] "+err.Error())
			}
		}
	} else {
		output, err = json.Marshal(libcpak.WebError {
			"Invalid package parameter",
			404,
		})
		if err != nil {
			libcpak.PushLog(0, "[Error (Package Listing)] "+err.Error())
		}
	}
	w.Write([]byte(output))
	return;
}


func handlerepo(w http.ResponseWriter, r *http.Request) {
	pack := r.FormValue("pkg")
	version := r.FormValue("ver")
}