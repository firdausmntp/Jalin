package speedtest

// DefaultServers lists well-known, high-bandwidth global endpoints.
var DefaultServers = []ServerConfig{
	{
		ID:          "cloudflare-global",
		Name:        "Cloudflare Edge",
		Location:    "Closest Anycast CDN",
		PingURL:     "https://speed.cloudflare.com/__down?bytes=0",
		DownloadURL: "https://speed.cloudflare.com/__down?bytes=25000000", // 25MB chunks
		UploadURL:   "https://speed.cloudflare.com/__up",
	},
	{
		ID:          "cloudflare-fallback",
		Name:        "Cloudflare CDN 1.1.1.1",
		Location:    "Global Anycast",
		PingURL:     "https://1.1.1.1/cdn-cgi/trace",
		DownloadURL: "https://speed.cloudflare.com/__down?bytes=10000000", // 10MB chunks
		UploadURL:   "https://speed.cloudflare.com/__up",
	},
}

// GetServerOrDefault returns a server config matching id, or the default first server.
func GetServerOrDefault(id string) ServerConfig {
	for _, s := range DefaultServers {
		if s.ID == id {
			return s
		}
	}
	return DefaultServers[0]
}
