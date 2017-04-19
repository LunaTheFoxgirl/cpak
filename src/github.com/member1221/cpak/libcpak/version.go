package libcpak

import (
	"strconv"
	"strings"
	"errors"
)

type Version struct {
	Patch int
	Minor int
	Major int
	Post string
}

func MakeVersion(major, minor, patch int) Version {
	return MakeVersionWithPost(major, minor, patch, "")
}

func MakeVersionWithPost(major, minor, patch int, post string) Version {
	return Version{patch, minor, major, post}
}

func (v Version) ToString() string {
	return strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor) + "." + strconv.Itoa(v.Patch)
}

func ParseVersionString(str string) (Version, error) {
	a := strings.Split(str, "-")
	s := strings.Split(a[0], ".")
	if len(s) == 3 {

		major, err := strconv.Atoi(s[0])
		if err != nil {
			return Version{}, err
		}

		minor, err := strconv.Atoi(s[1])
		if err != nil {
			return Version{}, err
		}

		patch, err := strconv.Atoi(s[2])
		if err != nil {
			return Version{}, err
		}

		if len(a) == 2 {
			return MakeVersionWithPost(major, minor, patch, a[1]), nil
		}
		return MakeVersion(major, minor, patch), nil
	}
	return Version{}, errors.New("Version string contained too many indexes.")
}