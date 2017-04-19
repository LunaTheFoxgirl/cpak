package libcpak

import "sync"

var logdata map[int][]string = make(map[int][]string)
var mutex sync.Mutex

func PushLog(verbosity int, message string) {
	mutex.Lock()
	logdata[verbosity] = append(logdata[verbosity], message)
	mutex.Unlock()
}

func LogLen(verbosity int) int {
	mutex.Lock()
	defer mutex.Unlock()
	return len(logdata[verbosity])
}

func PullLog(verbosity int) string {
	mutex.Lock()

	//HACK: HACKY CODE HERE. xD
	defer func () {
		logdata[verbosity] = logdata[verbosity][1:]
	}()
	defer mutex.Unlock()

	//return the thing n stuff
	return logdata[verbosity][0]

}