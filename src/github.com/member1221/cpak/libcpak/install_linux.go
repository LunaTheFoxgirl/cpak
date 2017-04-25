package libcpak

import "syscall"
import "os/user"
import "fmt"

func (pk Package) Install(sVersion bool) {
	//Load path seperators for the current operating system
	sep := PATH_SEP

	//Set working dir to root.
	syscall.Chdir(sep)

	//Get current installing user.
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}

	//Set the root path string for the application root
	root := sep + "cpak" + sep + pk.Name + "$container"
	if !sVersion {
		//Set special case versioned path string for the application root
		root = sep + "cpak" + sep + pk.Name + "#" + pk.Version.ToString() + "$container"
	}
	fmt.Println("Creating container for application...")
	if !FSExists(root) {
		//Create fake root folders
		CreateDirectory(root)
		CreateDirectory(root + sep + "bin")
		CreateDirectory(root + sep + "lib")
		CreateDirectory(root + sep + "lib64")
		CreateDirectory(root + sep + "bin64")
		CreateDirectory(root + sep + "dev")
		CreateDirectory(root + sep + "etc")
		CreateDirectory(root + sep + "proc")
		CreateDirectory(root + sep + "home" + sep + usr.Username)
		//Link /bin/bash to the new environment
		err = syscall.Link(sep + "bin" + sep + "bash", root + sep + "bin" + sep + "bash")
		if err != nil {
			fmt.Println("Error linking bash: " + err.Error())
		}
	}

	//Mount /proc to environment
	err = syscall.Mount(sep + "proc", root + sep + "proc", "", syscall.MS_BIND, "")
	if err != nil {
		fmt.Println("Error mounting proc: " + err.Error())
	}

	//Mount /dev to environment
	err = syscall.Mount(sep + "dev", root + sep + "dev", "", syscall.MS_BIND, "")
	if err != nil {
		fmt.Println("Error mounting dev: " + err.Error())
	}

	//chroot into the new environment
	syscall.Chdir(root)
	syscall.Chroot(root)
	syscall.Chdir(sep)
}
