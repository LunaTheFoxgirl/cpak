package libcpak

import (
	"errors"
	"os"
	"io/ioutil"
	"encoding/json"
	"github.com/cavaliercoder/grab"
	"strconv"
	"encoding/binary"
	"archive/tar"
	"compress/gzip"
)

var pkgcachefile string = "cpakcache.json"
var pkgcache map[string]Package

type Package struct {
	Name         	string `json:"name"`
	Origin       	string `json:"origin"`
	Version      	Version `json:"version"`
	Files 	     	map[string]string `json:"files"`
	Dependencies 	[]Package `json:"dependencies"`
	PreDependencies []Package `json:"pre-dependencies"`
}



func PreparePackageInstallation(file string) (PackageFile, error) {
	sep := PATH_SEP
	root := PATH_ROOT
	var err error
	var fl PackageFile = PackageFile{}
	var pk Package
	//If package file exists
	if _, err = os.Stat(file); err == nil {
		file, err := os.Open(file)
		if err != nil {
			return PackageFile{}, err
		}
		defer file.Close()

		//Check for magic bytes of "CPAK"
		b := make([]byte, len([]byte("CPAK")))
		file.Read(b)
		headerdef := string(b)
		if headerdef != "CPAK" {
			return PackageFile{}, errors.New("File " + file.Name() + " not an cpak package! (Invalid magic bytes)")
		}

		//Get length of and unmarshal json header
		b = make ([]byte, binary.MaxVarintLen64)
		file.Read(b)
		rlen := binary.BigEndian.Uint64(b)
		b = make([]byte, rlen)
		file.Read(b)

		err = json.Unmarshal(b, pk)
		if err != nil {
			return PackageFile{}, errors.New("File " + file.Name() + " not an cpak package! (Unable to unmarshal header)")
		}
		fl.Header = pk

		//Get prerecipe lua script
		b = make ([]byte, binary.MaxVarintLen64)
		file.Read(b)
		rlen = binary.BigEndian.Uint64(b)
		if rlen > 0 {
			b = make([]byte, rlen)
			file.Read(b)
			fl.PreRecipe = b
		}

		//Get postrecipe lua script
		b = make ([]byte, binary.MaxVarintLen64)
		file.Read(b)
		rlen = binary.BigEndian.Uint64(b)
		if rlen > 0 {
			b = make([]byte, rlen)
			file.Read(b)
			fl.PostRecipe = b
		}

		//Extract gzip tar archive to /tmp for installation
		b = make ([]byte, binary.MaxVarintLen64)
		file.Read(b)
		rlen = binary.BigEndian.Uint64(b)
		b = make([]byte, rlen)
		file.Read(b)
		gr, err := gzip.NewReader(file)
		if err != nil {
			return PackageFile{}, errors.New("File " + file.Name() + " not an cpak package! (Invalid gzip container)")
		}
		defer gr.Close()
		Untar(tar.NewReader(gr), root + "tmp" + sep + fl.Header.Name)
		fl.TmpFiles = root + "tmp" + sep + fl.Header.Name
		return fl, nil
	}
	return PackageFile{}, err
}

type PackageFile struct {
	Header Package
	TmpFiles string
	PreRecipe []byte
	PostRecipe []byte
}

type WebPackage struct {
	Name string `json:"name"`
	Origin string `json:"origin"`
	Category string `json:"category"`
	Version string `json:"version"`
}

func LoadPackageCache(root string) error {
	cache := root + "/" + pkgcachefile
	// Make sure file exists.
	if _, err := os.Stat(cache); err == nil {

		//Read file.
		file, err := ioutil.ReadFile(cache);
		if err != nil {
			return err
		}

		//Unmarshal file.
		err = json.Unmarshal(file, pkgcache)

		if err != nil {
			return err
		}

		//Return error result.
		return nil
	}

	//Return error if file doesn't exist.
	return errors.New("Application cache was not found!\n" +
		"If this issue persists, try running cpak gencache.\n" +
		"WARNING: cpak gencache will clear previous cache elements.\n" +
		"Software which is listed in the old cache will not be able to be managed without reinstalling the applications.")
}

func ClearPackageCache() {
	pkgcache = make(map[string]Package, 0)
}

func SavePackageCache(root string) error {
	cache := root + "/" + pkgcachefile
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

	//Unmarshal file.
	out, err := json.Marshal(pkgcache)

	if err != nil {
		return err
	}

	file.Write(out)
	//Return error result.
	return nil
}

func GetPackageDependanciesInstalled (p Package) ([]Package, []Package) {
	var pkgs []Package = make([]Package, 0)
	var ppkgs []Package = make([]Package, 0)
	for _, app := range p.Dependencies {

		//Check if pkgcache contains dependency.
		if _, is := pkgcache[app.Name]; is {
			pkgs = append(pkgs, app)
		}
	}
	for _, app := range p.PreDependencies {

		//Check if pkgcache contains pre-install dependency.
		if _, is := pkgcache[app.Name]; is {
			ppkgs = append(pkgs, app)
		}
	}
	return ppkgs, pkgs
}

func PackageListToString(pkgs []Package, verbosity int) (string, string, string) {
	pkl := ""
	depl := ""
	pdepl := ""
	for _, pkg := range pkgs {
		pkl += " " + pkg.Name
		if verbosity > 0 {
			pkl += "#" + pkg.Version.ToString()
		}
		for _, pdep := range pkg.PreDependencies {
			pdepl += " " + pdep.Name
			if verbosity > 0 {
				pdepl += "#" + pdep.Version.ToString()
			}
		}

		for _, dep := range pkg.Dependencies {
			depl += " " + dep.Name
			if verbosity > 0 {
				depl += "#" + dep.Version.ToString()
			}
		}
	}
	return pkl, depl, pdepl

}

func RequestPackages(pkgs []string) []Package {
	var found []Package = make([]Package, 0)
	for _, r := range *GetCache() {
		for _, pkg := range pkgs {
			if !PkgListContainsName(found, pkg) {
				pakg, err := r.RequestPackage(pkg)
				if err != nil {
					PushLog(0, err.Error())
					continue
				}
				found = append(found, pakg)
			}
		}
	}
	return found
}

func RequestPackageDependancies(rootPackage Package, preprecache []Package, precache []Package) ([]Package,[]Package, error) {
	var prepkgs []Package = preprecache
	var pkgs []Package = precache
	var req Package
	var err error
	length := len(rootPackage.Dependencies) + len(rootPackage.PreDependencies)
	pre := false
	got := 0
	for _, r := range *GetCache() {
		if got >= length {
			break
		}
		if !pre {
			for _, i := range rootPackage.PreDependencies {
				req, err = r.RequestPackage(i.Name + "#" + i.Version.ToString())
				if err != nil {
					PushLog(0, err.Error())
					continue
				}
				if (!PkgListContains(prepkgs, req)) {
					pkgs = append(prepkgs, req)
					got++
				}
			}
		} else {
			for _, i := range rootPackage.Dependencies {
				req, err = r.RequestPackage(i.Name + "#" + i.Version.ToString())
				if err != nil {
					PushLog(0, err.Error())
					continue
				}
				if (!PkgListContains(pkgs, req)) {
					pkgs = append(pkgs, req)
					got++
				}
			}
		}
	}
	return prepkgs, pkgs, nil
}

func PkgListContains(list []Package, pkg Package) bool {
	for _, i := range list {
		if i.Name == pkg.Name {
			return true
		}
	}
	return false
}

func PkgListContainsName(list []Package, pkgname string) bool {
	for _, i := range list {
		if i.Name == pkgname {
			return true
		}
	}
	return false
}

func strSliceContains(list []string, test string) bool {
	for _, i := range list {
		if i == test {
			return true
		}
	}
	return false
}

func DownloadPackage(pkg Package) (Package, error) {
	_, err := grab.NewRequest(pkg.Origin)
	if !grab.IsBadDestination(err) {
		PushLog(0, "Downloading package "+pkg.Name+"...")
		p, err := grab.GetAsync("/tmp/" + pkg.Name + ".cpak", pkg.Origin)
		if err != nil {
			return Package{}, nil
		}

		//Begin Download progress listing
		pp := <-p
		for (!pp.IsComplete()) {
			pp := <-p
			PushLog(0, "["+pkg.Name+"] Progress: "+strconv.FormatFloat(pp.Progress()*100, 'E', -1, 64)+"%")
		}
		PushLog(0, "["+pkg.Name+"] Package download completed.")
	}
	return Package{}, errors.New("Package not found!\n\n" +
		"If you want package " + pkg.Name + " to exist, try creating a cpak package for the application.\n" +
		"Or run cpak list to see a list of applications.")
}