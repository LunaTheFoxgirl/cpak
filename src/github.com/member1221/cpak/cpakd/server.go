package main

type ServerConfig struct {
	Name string `json:"name"`
	ServeDir string `json:"serve_dir"`
	Categories []string `json:"serve_categories"`
}
