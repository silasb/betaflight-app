package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
)

const (
	UPDATE_HOST_BASE = "http://foo.us.to/gui2/"
)

type Version struct {
	Version string `json:"version"`
	File    string `json:"file"`
}

type Versions struct {
	Versions []Version
}

func HasNewerVersion(versionString string) (bool, *Version) {
	resp, err := http.Get(UPDATE_HOST_BASE + "versions.json")
	if err != nil {
		log.Println(err)
		return false, nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var versions []Version
	json.Unmarshal(body, &versions)

	var latestVersion Version

	for _, version := range versions {
		if version.Version > versionString {
			latestVersion = version
			break
		}
	}

	if latestVersion.Version == "" {
		log.Println("We are up to date")
		return false, nil
	}

	log.Println("Can upgrade to", latestVersion.Version)

	return true, &latestVersion
}

func UpdateBinary(version *Version) (err error) {
	log.Println("Upgrading to version", version.Version)

	resp, err := http.Get(UPDATE_HOST_BASE + version.File)
	if err != nil || resp.StatusCode != 200 {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	dest, err := os.Executable()
	if err != nil {
		return err
	}

	// copy current exe to backup so we can try to restore if something fails
	destBackup := dest + ".bak"
	if _, err := os.Stat(dest); err == nil {
		os.Rename(dest, destBackup)
	}

	log.Printf("Downloading new version to %s\n", dest)
	if err := ioutil.WriteFile(dest, data, 0755); err != nil {
		// something failed, let's restore the backup
		os.Rename(destBackup, dest)
		return err
	}

	// we suceeded let's delete the backup
	os.Remove(destBackup)

	log.Printf("Updated to version %s\n", version.Version)
	log.Println("Restarting app")

	var args []string
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		// If the command fails to run or doesn't complete successfully, the
		// error is of type *ExitError. Other error types may be
		// returned for I/O problems.
		log.Println(err)
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The command didn't complete correctly.
			// Exiting while keeping the status code.
			os.Exit(exiterr.Sys().(syscall.WaitStatus).ExitStatus())
		} else {
			return err
		}
	}

	os.Exit(0)

	return nil
}
