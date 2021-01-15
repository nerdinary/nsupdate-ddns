package main

import (
	"context"
	"ddns/web"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

const (
	v4Check = "https://ipv4.nsupdate.info/myip"
	v6Check = "https://ipv6.nsupdate.info/myip"

	v4Update = "https://ipv4.nsupdate.info/nic/update"
	v6Update = "https://ipv6.nsupdate.info/nic/update"
)

// ips holds the v6 and v6 metadata.
type ips struct {
	v4  ipMeta
	v6  ipMeta
	cfg *Config
}

// ipMeta holds the resolved and public addresses of the host.
type ipMeta struct {
	current string
	public  string
	diff    bool
}

// Config holds the authentication & hostname info.
type Config struct {
	Username string
	Password string
	Hostname string
}

// checkLocal uses a DNS resolver to get the IP address of the requested host.
func (i *ips) checkLocal(ctx context.Context) error {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
	currV4, err := r.LookupIP(ctx, "ip4", i.cfg.Hostname)
	if err != nil {
		// If no address is set, we don't want to fail as it'll require an update.
		fmt.Println(err)
	}
	for _, c := range currV4 {
		i.v4.public = c.To4().String()
	}

	currV6, err := r.LookupIP(ctx, "ip6", i.cfg.Hostname)
	if err != nil {
		// If no address is set, we don't want to fail as it'll require an update.
		fmt.Println(err)
	}
	for _, c := range currV6 {
		i.v6.public = c.To16().String()
	}
	return nil
}

// checkPublic gets the public v4 and v6 addressess of this current host.
func (i *ips) checkPublic() error {
	var err error
	i.v4.current, err = web.MakeRequest(v4Check)
	if err != nil {
		return err
	}
	i.v6.current, err = web.MakeRequest(v6Check)
	if err != nil {
		return err
	}
	return nil
}

// updateDNS sends a request to update the host IP.
func (i *ips) updateDNS() error {
	client := web.New(i.cfg.Username, i.cfg.Password)
	if i.v4.diff {
		if err := client.UpdateIP(v4Update); err != nil {
			return err
		}
	}
	if i.v6.diff {
		if err := client.UpdateIP(v6Update); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	cfgFile := flag.String("config", "config.json", "Config JSON file")
	flag.Parse()

	file, err := os.Open(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	if err := decoder.Decode(&config); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ip := &ips{cfg: &config}
	if err := ip.checkLocal(ctx); err != nil {
		log.Fatal(err)
	}
	if err := ip.checkPublic(); err != nil {
		log.Fatal(err)
	}

	noChg := true
	if ip.v4.current != ip.v4.public {
		ip.v4.diff = true
		noChg = false
	}
	if ip.v6.current != ip.v6.public {
		ip.v6.diff = true
		noChg = false
	}

	if noChg {
		fmt.Printf("No update required, IPs are [%v, %v]\n", ip.v4.public, ip.v6.public)
		return
	}

	if err := ip.updateDNS(); err != nil {
		log.Fatal(err)
	}
}
