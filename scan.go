package nmap

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
)

// Scan holds one nmap scan. It can be rescanned, diff'ed, and parsed for hosts
type Scan struct {
	DisplayArgs string
	Hosts       []Host
}

func (scan rawScan) cleanScan() Scan {
	output := Scan{scan.DisplayArgs, []Host{}}
	for _, host := range scan.Hosts {
		output.Hosts = append(output.Hosts, host.cleanHost())
	}

	return output
}

// RunScan is used to scan hosts with a list of hosts, ports, and nmap flags
func RunScan(hosts []string, ports []int, opts []string) (*Scan, error) {
	// Parse arguments
	args := []string{"-oX", "-"}

	portList := ""
	if len(ports) != 0 {
		for _, port := range ports {
			portList += strconv.Itoa(port) + ","
		}
	}
	args = append(args, opts...)
	if portList != "" {
		args = append(args, "-p", portList)
	}
	args = append(args, hosts...)

	// Find path for nmap binary
	nmapPath, err := exec.LookPath("nmap")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(nmapPath, args...)

	// Configure output pipes
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Read output
	stdout, err := ioutil.ReadAll(outPipe)
	if err != nil {
		return nil, err
	}

	stderr, err := ioutil.ReadAll(errPipe)
	if err != nil {
		return nil, err
	}

	// Wait on command to be finished
	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		return nil, errors.New(err.Error() + "\n" +
			string(stderr) + "\n" + string(stdout))
	}

	// Parse command output
	rawScan, err := parseXML(stdout)
	if err != nil {
		fmt.Println("Throwing error parsing!")
		return nil, err
	}

	hostScan := rawScan.cleanScan()

	return &hostScan, nil
}

// Rescan will rescan all hosts in the Scan
func (s *Scan) Rescan() (*Scan, error) {
	newScan := new(Scan)
	for _, host := range s.Hosts {
		host, err := host.Rescan()
		if err != nil {
			return nil, err
		}
		newScan.Hosts = append(newScan.Hosts, *host)
	}

	return newScan, nil
}

// RescanWithOptions will rescan all hosts in the Scan
func (s *Scan) RescanWithOptions(opts []string) (*Scan, error) {
	newScan := new(Scan)
	for _, host := range s.Hosts {
		host, err := host.RescanWithOptions(opts)
		if err != nil {
			return nil, err
		}
		newScan.Hosts = append(newScan.Hosts, *host)
	}

	return newScan, nil
}
