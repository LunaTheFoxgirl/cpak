package main

import "github.com/member1221/cpak/libcpak"

type ServerConfig struct {
	Name string `json:"name"`
	Port int `json:"port"`
	ServeDir string `json:"serve_dir"`
	Version libcpak.Version `json:"version"`
	Categories []string `json:"serve_categories"`
}
