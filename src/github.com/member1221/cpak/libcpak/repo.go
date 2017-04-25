package libcpak

import (
	"errors"
	"io/ioutil"
	"os"
	"encoding/json"
	"strings"
	"strconv"
	"fmt"
)

var repocachefile string = "repos.json"
var repocache []Repository

type Repository struct {
	Name string `json:"name"`
	Origin string `json:"origin"`
	Version Version `json:"version"`
}

func LoadRepoCache(root string) error {
	cache := root + "/" + repocachefile
	repol, err := GetRepositories(cache)
	if err != nil {
		return err
	}
	repocache = repol
	return nil
}

func FetchRepo(origin string) (Repository, string, error) {
	var repo Repository
	if !strings.Contains(origin, ":") {
		origin = ":" + strconv.Itoa(CPAK_DEFAULT_SERVE_PORT)
	}
	res, err := rest.Request(RestRequest{
			RequestType: R_POST,
			Directory: "http://" + origin + "/cpak/repo",
		}, []RestRequestData{})

	if err != nil {
		return Repository{}, "", err
	}
	err = json.Unmarshal([]byte(res), &repo)

	if err != nil {
		return Repository{}, "", err
	}
	return repo, origin, nil
}

func SaveRepoCache(root string) error {
	cache := root + "/" + repocachefile
	if _, err := os.Stat(cache); err != nil {
		c, err := os.Create(cache)
		if err != nil {
			return err
		}
		c.Close()
	}

	file, err := os.OpenFile(cache, os.O_RDWR, 0)
	defer file.Close()

	if err != nil {
		return err
	}

	//Marshal file.
	out, err := json.MarshalIndent(repocache, "", "\t")

	if err != nil {
		return err
	}

	_, err = file.Write([]byte(string(out) + "\n"))
	if err != nil {
		return err
	}

	//Return error result.
	return nil
}

func GetCache() *[]Repository {
	return &repocache
}

func AddRepoToCache(repository Repository) {
	repocache = append(repocache, repository)
}

func RemoveRepoFromCache(repositoryName string) {
	//TODO: make this do stuff.
}

type pkgRequestResponse struct {
	error string
	pkg Package
}

var rest = New()
func (r Repository) RequestPackage(name string) (Package, error) {
	var pkg pkgRequestResponse

	res, err := rest.Request(RestRequest{
		RequestType: R_POST,
		Directory: r.Origin + "/cpak/repolist",
	}, []RestRequestData{
		RestRequestData{
			"pkg",
			name,
		},
	})

	if err != nil {
		return Package{}, err
	}

	json.Unmarshal([]byte(res), pkg)

	if pkg.error != "" {
		return Package{}, errors.New(pkg.error)
	}

	return pkg.pkg, nil
}

func GetRepositories(list string) ([]Repository, error) {

	// Make sure file exists.
	if _, err := os.Stat(list); err == nil {

		//Read file.
		file, err := ioutil.ReadFile(list);
		if err != nil {
			return nil, err
		}
		fmt.Println(string(file))
		//Unmarshal file.
		var repos []Repository
		err = json.Unmarshal(file, repos)

		if err != nil {
			return nil, err
		}

		//Return unmarshalled result.
		return repos, nil
	}

	//Return error if file doesn't exist.
	return nil, errors.New("Repository list was not found!\n" +
		"Please run cpak repo generate to generate an repository list.\n" +
		"Afterwards add a repository with cpak repo add (link)")
}