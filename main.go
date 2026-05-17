package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

const discordURL = "https://discord.com/api/download/stable?platform=linux&format=deb"

func getInstalledVersion() (string, error) {
	// Runs a shell command and captures its stdout as a []byte
	out, err := exec.Command("dpkg", "-s", "discord").Output()
	if err != nil {
		//Discord is probably not installed yet
		return "", nil
	}
	// string(out) converts []byte to a string for it to used by strings.Split()
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Version:") {
			// "Version: 0.0.77 => "0.0.77"
			version := strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
			return version, nil
		}
	}
	return "", fmt.Errorf("could not parse installed version")
}

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

	latestVersion, err := extractVersion(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Latest version:", latestVersion)

	installedVersion, err := getInstalledVersion()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if installedVersion == "" {
		fmt.Println("Discord is not installed yet")
	} else {
		fmt.Println("Installed version:", installedVersion)
	}

	if installedVersion == latestVersion {
		fmt.Println("Already up to date! Launching Discord...")
		// Launch Discord here
		return
	}

	fmt.Println("Update available! Will now download and install", latestVersion)
}
