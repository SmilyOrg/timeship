package network

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
)

type ListenURL struct {
	Local bool
	IPv6  bool
	URL   string
}

// GetListenURLs returns all URLs that a listener is available on
func GetListenURLs(addr net.Addr) ([]ListenURL, error) {
	var urls []ListenURL
	switch vaddr := addr.(type) {
	case *net.TCPAddr:
		if vaddr.IP.IsUnspecified() {
			ifaces, err := net.Interfaces()
			if err != nil {
				return urls, fmt.Errorf("unable to list interfaces: %v", err)
			}
			for _, i := range ifaces {
				addrs, err := i.Addrs()
				if err != nil {
					return urls, fmt.Errorf("unable to list addresses for %v: %v", i.Name, err)
				}
				for _, a := range addrs {
					switch v := a.(type) {
					case *net.IPNet:
						urls = append(urls, ListenURL{
							Local: v.IP.IsLoopback(),
							IPv6:  v.IP.To4() == nil,
							URL:   fmt.Sprintf("http://%v", net.JoinHostPort(v.IP.String(), strconv.Itoa(vaddr.Port))),
						})
					default:
						urls = append(urls, ListenURL{
							URL: fmt.Sprintf("http://%v", v),
						})
					}
				}
			}
		} else {
			urls = append(urls, ListenURL{
				Local: vaddr.IP.IsLoopback(),
				URL:   fmt.Sprintf("http://%v", vaddr.AddrPort()),
			})
		}
	default:
		urls = append(urls, ListenURL{
			URL: fmt.Sprintf("http://%v", addr),
		})
	}
	return urls, nil
}

// PrintListenURLs prints all URLs that a listener is available on
func PrintListenURLs(addr net.Addr, apiPrefix string) error {
	urls, err := GetListenURLs(addr)
	if err != nil {
		return err
	}
	// Sort by ipv4 first, then local, then url
	sort.Slice(urls, func(i, j int) bool {
		if urls[i].IPv6 != urls[j].IPv6 {
			return !urls[i].IPv6
		}
		if urls[i].Local != urls[j].Local {
			return urls[i].Local
		}
		return urls[i].URL < urls[j].URL
	})

	for _, url := range urls {
		prefix := "network"
		if url.Local {
			prefix = "local"
		}
		log.Printf("  %-8s %s%s\n", prefix, url.URL, apiPrefix)
	}
	return nil
}
