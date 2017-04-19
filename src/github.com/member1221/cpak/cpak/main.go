package main

import (
	"fmt"
	"github.com/member1221/cpak/libcpak"
	"os"
	"sync"
	"bufio"
	"strings"
)

var VERSION libcpak.Version = libcpak.MakeVersion(0,1,0)

var CONFIGLOCATION string = libcpak.CPAK_DEFAULT_CFG_DIR

var HELPTEXT string = "Arguments:\n" +
	"	list - Lists applications in specified category or with specified prefix.\n" +
	"	install - Installs an application\n" +
	"	reinstall - \"Nukes\" an application and reinstalls it.\n" +
	"	remove - Removes an application\n" +
	"	purge - Purges an applications configuration files\n" +
	"	nuke - Completely removes ALL files for an application\n" +
	"	help/-h - This screen.\n\n" +
	"Modifiers:" +
	"	updates - Combined with [install], installs updates to all applications if found.\n"

func main() {
	fmt.Println("[CPAK PAcKager " + VERSION.ToString() + "]")
	if os.Geteuid() == 0 {
		args := os.Args[1:]
		if args != nil && len(args) > 0 {
			action := args[0]

			if action == "gencache" {
				fmt.Println("Generating a new application cache...")
				libcpak.ClearPackageCache()
				err := libcpak.SavePackageCache(CONFIGLOCATION)
				if err != nil {
					fmt.Println("[Error] " + err.Error())
				}
				//TODO: Generate cache
				return;
			} else if action == "repo" {
				if (len(args) > 1) {
					action2 := args[1]
					if action2 == "generate" {
						libcpak.SaveRepoCache(CONFIGLOCATION)
						return
					} else if action2 == "add" {

					} else if action2 == "remove" {

					}
				}
				fmt.Println("No repository action was defined!\n" +
					"Actions:\n" +
					"	generate - Generate a new repository list.\n" +
					"	add (repository) - Add a repository to the list.\n" +
					"	remove (repository) - Remove a repository from the list.")
				return;
			}

			err := libcpak.LoadPackageCache(CONFIGLOCATION)
			if err != nil && err.Error() != "unexpected end of JSON input" {
				fmt.Println("\n[Error] " + err.Error())
			}

			err = libcpak.LoadRepoCache(CONFIGLOCATION)
			if err != nil && err.Error() != "unexpected end of JSON input"  {
				fmt.Println("\n[Error] " + err.Error())
			}

			if action == "install" {

				if len(args) > 1 {
					handleInstall(args[1:])
					return
				}
				fmt.Println("No applications specified to install!")


			} else if action == "pull" {
				//TODO: If needed in future, this will pull repo updates.
			} else if action == "reinstall" {

				if len(args) > 1 {

				}
				fmt.Println("No applications specified to reinstall!")
			} else if action == "remove" {

				if len(args) > 1 {

				}
				fmt.Println("No applications specified to remove!")
			} else if action == "purge" {

				if len(args) > 1 {

				}
				fmt.Println("No applications specified to purge!")
			} else if action == "nuke" {

				if len(args) > 1 {

				}
				fmt.Println("No applications specified to nuke!")
			} else if action == "help" || action == "-h" {
				fmt.Println(HELPTEXT)
			} else {
				fmt.Println("Invalid argument " + action + "!\n\nRun [cpak help] for help.")
			}
			return
		}
	} else {
		fmt.Println("Please run CPAK as root/administrator.")
		return
	}
	fmt.Println("No arguments was passed!\n\n" + HELPTEXT)
}

//Handles the action of installing a package.
func handleInstall(apps []string) {
	fmt.Println("Requesting packages...")
	pkgs := libcpak.RequestPackages(apps)
	var predependancies []libcpak.Package = make([]libcpak.Package, 0)
	var dependancies []libcpak.Package = make([]libcpak.Package, 0)
	max := 10
	i := 0
	mut := sync.Mutex{}
	fmt.Println("Requesting (pre)dependancies...")
	for _, pkg := range pkgs {
		p, d, err := libcpak.RequestPackageDependancies(pkg, predependancies, dependancies)
		if err != nil {
			return;
		}
		predependancies = append(predependancies, p...)
		dependancies = append(dependancies, d...)
	}

	if len(pkgs) == 0 {
		fmt.Println("Requested package(s) was not found in repositories, sorry.")
		return;
	}

	fmt.Println("Here's the packages you'll install:\n\nPackages:")
	a, b, c := libcpak.PackageListToString(pkgs, 0)
	fmt.Println(a + "\n\n" +
		"Pre-dependancies:\n" + b + "\n\n" +
		"Dependancies:\n" + c + "\n\n")

	cont := false
	for !cont {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Are you sure you want these packages installed? [Y/N] ")
		text, _ := reader.ReadString('\n')
		if strings.ToLower(text[:1]) == "y" {
			cont = true
		} else if strings.ToLower(text[:1]) == "n" {
			fmt.Println("Exited by user...")
			return
		}
	}


	done := false
	go func() {
		for _, dep := range predependancies {
			libcpak.DownloadPackage(dep)
		}

		for _, dep := range dependancies {
			go func() {
				mut.Lock()
				i++
				mut.Unlock()

				libcpak.DownloadPackage(dep)

				mut.Lock()
				i--
				mut.Unlock()

			}()
			for (i >= max) {
				//Wait...
			}
		}

		for _, pkg := range pkgs {
			go func() {
				mut.Lock()
				i++
				mut.Unlock()

				libcpak.DownloadPackage(pkg)

				mut.Lock()
				i--
				mut.Unlock()

			}()
			for (i >= max) {
				//Wait...
			}
		}
		done = true
	}()
	for !done {
		if libcpak.LogLen(0) > 0 {
			fmt.Println(libcpak.PullLog(0));
		}
	}
	return
}