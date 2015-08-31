package mdocker

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/client"
)

func getPortForContainerInterface(guestID, ifaceName string) (string, error) {
	command := "ovs-vsctl"
	args := []string{
		"--data=bare",
		"--no-heading",
		"--columns=name",
		"find",
		"interface",
		"external_ids:container_id=" + guestID,
		"external_ids:container_iface=" + ifaceName,
	}
	output, err := exec.Command(command, args...).CombinedOutput()
	if err != nil {
		e := fmt.Errorf("failed to look up name of interface %s for guest %s",
			ifaceName,
			guestID,
		)
		log.WithFields(log.Fields{
			"error":   err,
			"command": command,
			"args":    args,
			"output":  string(output),
		}).Error(e)
		return "", e
	}
	return strings.TrimSpace(string(output)), nil
}

func addPort(g *client.Guest, nic client.Nic) (string, error) {
	command := "ovs-docker"
	args := []string{"add-port",
		nic.Network,
		nic.Name,
		g.ID,
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
		return "", e
	}
	return getPortForContainerInterface(g.ID, nic.Name)
}

func tagPort(port string, vlanInts []int) error {
	command := "ovs-vsctl"

	if len(vlanInts) == 0 {
		return nil
	}

	vlans := make([]string, len(vlanInts), len(vlanInts))
	for i := 0; i < len(vlanInts); i++ {
		vlans[i] = strconv.Itoa(vlanInts[i])
	}

	args := []string{
		"set",
		"port",
		port,
		"trunks=" + strings.Join(vlans, ","),
	}

	if output, err := exec.Command(command, args...).CombinedOutput(); err != nil {
		e := fmt.Errorf("failed to tag interface %s", port)
		log.WithFields(log.Fields{
			"error":   err,
			"command": command,
			"args":    args,
			"output":  string(output),
		}).Error(e)
		return e
	}

	return nil
}

// addInterfaces adds network interfaces to a guest container
func addInterfaces(g *client.Guest) error {
	for _, nic := range g.Nics {
		port, err := addPort(g, nic)
		if err != nil {
			return err
		}
		if err := tagPort(port, nic.VLANs); err != nil {
			return err
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
			g.ID,
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
