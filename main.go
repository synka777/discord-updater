package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const discordURL = "https://discord.com/api/download/stable?platform=linux&format=deb"

func launchDiscord() error {
	cmd := exec.Command("discord")
	err := cmd.Start() // Start() launches and returns immediately, Run() would wait
	if err != nil {
		return err
	}
	fmt.Println("Discord launched!")
	return nil
}

func installPackage(debPath string) error {
	cmd := exec.Command("sudo", "dpkg", "-i", debPath)

	// Instead of capturing output like we did with .Output() before, we're wiring the command's
	// output stream directly to the terminal's output stream.
	// Whatever dpkg prints, the user sees it in real time.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run() // Runs the command and waits for it to finish
}

func downloadFile(url string, version string) (string, error) {
	// fmt.Printf() returns a string instead of printing it.
	// Great for building strings with variables, like our file path.
	destPath := fmt.Sprintf("/tmp/discord-%s.deb", version)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	//  Creates (or truncates) a file on disk and returns a handle to write into it.
	file, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// resp.Contentlength is the size of the download in bytes
	// taken from the HTTP Content-Length header.
	totalBytes := resp.ContentLength // total size in bytes, -1 if unknown
	downloaded := int64(0)

	// Here make() allocates a 32KB byte slice to use as a read buffer.
	// make() is how you allocate slices and maps in Go.
	buf := make([]byte, 32*1024)
	for { // infinite loop, break when EOF is reached
		n, err := resp.Body.Read(buf)
		if n > 0 {
			// writes only the n bytes that were actually read - no junk data created
			_, writeErr := file.Write(buf[:n])
			if writeErr != nil {
				return "", writeErr
			}
			downloaded += int64(n)
			if totalBytes > 0 {
				// \r moves the cursor back to the start of the line
				//
				percent := float64(downloaded) / float64(totalBytes) * 100
				// %f => float; .1 => one decimal, to place IN %f; % => escape the next character
				fmt.Printf("\rDownloading... %.1f%%", percent)
			} else {
				fmt.Printf("\rDownloaded %d bytes", downloaded)
			}
		}
		if err == io.EOF {
			break // Download complete
		}
		if err != nil {
			return "", err
		}
	}
	fmt.Println() // newline after progress line
	return destPath, nil
}

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
		launchDiscord()
		return
	}

	fmt.Println("Update available! Download version", latestVersion)

	debPath, err := downloadFile(url, latestVersion)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Downloaded to:", debPath)

	fmt.Println("Installing update...")
	// We're using = instead of := here.
	// That's because err is already declared from a previous line
	err = installPackage(debPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Installed successfully!")

	os.Remove(debPath)
	fmt.Println("Cleaned up", debPath)

	print("Launching Discord...")
	err = launchDiscord()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
