package libcpak

import (
	"errors"
	"io/ioutil"
	"os"
	"encoding/json"
)

var repocachefile string = "cpakrepos.json"
var repocache []Repository

type Repository struct {
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

func SaveRepoCache(root string) error {
	cache := root + "/" + repocachefile
	if _, err := os.Stat(cache); err != nil {
		c, err := os.Create(cache)
		if err != nil {
			return err
		}
		c.Close()
	}

	file, err := os.Open(cache)
	defer file.Close()

	if err != nil {
		return err
	}

	//Marshal file.
	out, err := json.Marshal(repocache)

	if err != nil {
		return err
	}

	file.Write(out)
	//Return error result.
	return nil
}

func GetCache() []Repository {
	return repocache
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
		Directory: r.Origin + "/cpak/repo",
	}, []RestRequestData{
		RestRequestData{
			"package",
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