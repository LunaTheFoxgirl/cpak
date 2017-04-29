package libcpak

import (
	"sync"
	"fmt"
	"encoding/json"
	"strings"
	"strconv"
)

var translatedStrings map[string]string
var mtx sync.Mutex

func Translate(transval, defaultval string, inputs ...string) string {
	mtx.Lock()
	t := translatedStrings[transval]
	if t == "" {
		t = defaultval
	}

	for i, txt := range inputs {
		strings.Replace(t,"{" + strconv.Itoa(i) + "}", txt, 0)
	}

	defer mtx.Unlock()
	return t
}

func LoadLang(langname string) {
	lng, err := ReadFromFile(CPAK_DEFAULT_CFG_DIR + PATH_SEP + "langs" + PATH_SEP + langname + ".lang")
	if err != nil {
		fmt.Println("[CPAK_LANG ERROR] " + err.Error() + "\nCPAK will default to internal english.")
	}
	err = json.Unmarshal(lng, translatedStrings)
	if err != nil {
		fmt.Println("[CPAK_LANG JSON ERROR] " + err.Error() + "\nCPAK will default to internal english.")
	}
}