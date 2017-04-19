package libcpak

const (
	CPAK_DEFAULT_CFG_DIR = "/etc/cpak.d"
	CPAK_DEFAULT_SERVE_DIR = "/opt/cpak/serve"
	CPAK_DEFAULT_REPO_NAME = "Server ECLIPSED_TIDES"
	CPAK_DEFAULT_SERVE_PORT = 80
)

func CPAK_DEFAULT_REPO_CATS() []string {
	return []string { "games", "fun", "system", "drawing", "security", "tools"}
}


type WebError struct {
	Message string `json:"message"`
	Code int `json:"code"`
}