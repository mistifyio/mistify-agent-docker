package mdocker

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/client"
)

// addInterfaces adds network interfaces to a guest container
func addInterfaces(g *client.Guest) error {
	command := "ovs-docker"
	for _, nic := range g.Nics {
		args := []string{"add-port",
			nic.Network,
			nic.Name,
			g.Id,
			"--macaddress=" + nic.Mac, // ovs-docker errors if separate
		}
		if output, err := exec.Command(command, args...).CombinedOutput(); err != nil {
			e := fmt.Errorf("failed to add interface %s", nic.Name)
			log.WithFields(log.Fields{
				"error":   err,
				"command": command,
				"args":    args,
				"output":  string(output),
			}).Error(e)
			return e
		}
	}
	return nil
}

// removeInterfaces removes network interfaces from a guest container
func removeInterfaces(g *client.Guest) error {
	command := "ovs-docker"
	for _, nic := range g.Nics {
		args := []string{
			"del-port",
			nic.Network,
			nic.Name,
			g.Id,
		}
		if output, err := exec.Command(command, args...).CombinedOutput(); err != nil {
			// Ignore errors when trying to remove interface that is already gone
			if !strings.Contains(strings.ToLower(string(output)), "failed to find any attached port") {
				e := fmt.Errorf("failed to remove interface %s", nic.Name)
				log.WithFields(log.Fields{
					"error":   err,
					"command": command,
					"args":    args,
					"output":  string(output),
				}).Error(e)
				return e
			}
		}
	}
	return nil
}
