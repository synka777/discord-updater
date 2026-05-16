package main

import (
	"fmt"
	"net/http"
	"strings"
)

const discordURL = "https://discord.com/api/download/stable?platform=linux&format=deb"

func extractVersion(url string) (string, error) {
	// Split "https://dl.discordapp.net/apps/linux/0.0.77/discord-0.0.77.deb"
	// into ["https:", "", "dl.discordapp.net", "apps", "linux", "0.0.77", "discord-0.0.77.deb"]
	parts := strings.Split(url, "/") // Returns a slice of strings ([]string)

	for _, part := range parts {
		// The version segment looks like "O.O.77" - three numbers separated by dots
		if strings.Count(part, ".") == 2 {
			return part, nil
		}
	}
	return "", fmt.Errorf("could not extract version from URL: %s", url)
}

func getLatestURL() (string, error) {
	// Create an HTTP client that does NOT follow redirects
	client := &http.Client{
		// CheckRedirect = hook that fires when the client is about to follow a redirect.
		// Returning http.ErrUseLastResponse tells it: "stop here, give me back the redirect response as-is".
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// "stop here, give me back the redirect response as-is".
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(discordURL)
	if err != nil {
		return "", err
	}
	// defer means "run this line when the function exits", no matter what. It's how Go ensures you don't leak open connections.
	// You'll write defer + Close() almost every time you open something.
	defer resp.Body.Close()

	// The redirect destination is in the "Location" header
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no redirect location found")
	}
	return location, nil
}

func main() {
	fmt.Println("Discord updater starting...")
	url, err := getLatestURL()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Latest Discord URL:", url)

	version, err := extractVersion(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Latest version:", version)

}
