package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
)

//go:embed templates/* static/*
var content embed.FS

func startServer() {
	// Serve static files
	staticFS, _ := fs.Sub(content, "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// API routes
	http.HandleFunc("/api/todos", handleAPITodos)
	http.HandleFunc("/api/todos/", handleAPITodo)
	http.HandleFunc("/api/categories", handleAPICategories)

	// Serve main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := content.ReadFile("templates/index.html")
		if err != nil {
			http.Error(w, "Template not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

	// Get local IP for display
	ip := getLocalIP()
	fmt.Println("Todo Web UI starting...")
	fmt.Println()
	fmt.Printf("  Local:   http://localhost:8080\n")
	if ip != "" {
		fmt.Printf("  Network: http://%s:8080\n", ip)
	}
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")

	http.ListenAndServe(":8080", nil)
}

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			ipStr := ip.String()
			if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") || strings.HasPrefix(ipStr, "172.") {
				return ipStr
			}
		}
	}
	return ""
}
