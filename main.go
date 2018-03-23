// noip is a utility to update a no-ip.com DNS records.
//
// It reads a JSON-formatted config file like this:
//
//	{
//		"username": "user",
//		"password": "secret",
//		"hostname": "host1.domain.com",
//		"interface": "eth1"
//	}
//
// If updating multiple hostnames or groups use a comma separated list:
//
//	{
//		"username": "user",
//		"password": "secret",
//		"hostname": "host1.domain.com,group1,host2.domain.com",
//		"interface": "eth1"
//	}
//
// A successful update will produce no output, as they say "no news is
// good news".
//
// Usage of noip:
//	noip
//	noip -config /home/user/noip.json
//
// Use Crontab to run the noip on a scheduled basis:
//
//	$ crontab -e
//	# Run every 10 minutes
//	10 * * * * /usr/sbin/noip
//
package main // import "mazebuhu.io/noip"

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
)

const (
	updateBaseURL = "https://dynupdate.no-ip.com/nic/update"
	userAgent     = "noip/v1.0.0 mazebuhu.io/noip"
)

func main() {
	var (
		confFile = flag.String("config", "/etc/noip/noip.json", "JSON-formatted config file.")
	)
	flag.Parse()

	data, err := ioutil.ReadFile(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	var cfg struct {
		Username, Password, Hostname, Interface string
	}

	if err = json.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}

	ip, err := IPv4(cfg.Interface)
	if err != nil {
		log.Fatal(err)
	}

	u, err := url.Parse(updateBaseURL)
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	q.Set("hostname", cfg.Hostname)
	q.Set("myip", ip)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(cfg.Username, cfg.Password)
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error in response: %q\n", resp.Status)
	}
}

// IPv4 tries to determine the IPv4 address of the given
// network interface name.
func IPv4(name string) (string, error) {
	i, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}

	addrs, err := i.Addrs()
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		if ipn, ok := a.(*net.IPNet); ok {
			if ipn.IP.To4() != nil {
				return ipn.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no IPv4 found for interface: %q", name)
}
