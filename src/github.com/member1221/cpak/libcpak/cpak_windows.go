package libcpak

const (
	CPAK_DEFAULT_CFG_DIR = "C:\\etc\\cpak.d"
	CPAK_DEFAULT_ICACHE_DIR = "C:\\etc\\cpak.d\\inst"
	CPAK_DEFAULT_SERVE_DIR = "C:\\opt\\cpak\\serve"
	CPAK_DEFAULT_REPO_NAME = "Server ECLIPSED_TIDES"
	//Yes, the default port is 1337, live with it.
	CPAK_DEFAULT_SERVE_PORT = 1337

	PATH_SEP = "\\"
	PATH_ROOT = "C:"
)

func CPAK_DEFAULT_REPO_CATS() []string {
	return []string { "games", "fun", "system", "drawing", "security", "tools"}
}

type WebError struct {
	Message string `json:"message"`
	Code int `json:"code"`
}