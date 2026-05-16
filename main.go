package main

import (
	"fmt"
	"net/http"
)

const discordURL = "https://discord.com/api/download/stable?platform=linux&format=deb"

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
}
