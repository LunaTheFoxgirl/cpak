# CPAK Package Manager
CPAK is a package manager that is being written in golang. The goal of it is to be simple to use and develop tools for.
Before building/coding, run `chmod +x configure.sh; ./configure.sh`, to fetch all the neccesary libraries for development.
To build, run `go install github.com/member1221/cpak/cpak`, a file called cpak should now be in the bin directory.


I'm not a lawyer if any of the licensing here is wrong, please tell me (nicely) how to improve it, would be really appriciated.


# TODO:
- Improve server side software
- Package installation
- Fake root environments for installed packages
- Multithreaded package installation
- Add documentation
- Flags
  * --json (output json only)
  * --yes (Always answer yes to questions)
  * --systemwide (Do not install package in a container)
  * --extractonly (Only extract the package)
- Action tags
  * `install updates` (find packages and update them)
- More to be added.
