package discovery

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func ExampleDiscoverer() {
	localhost := net.IP([]byte{127, 0, 0, 1})
	localhost6 := net.IP([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})

	// utility for parsing port
	getPort := func(addr string) int {
		_, p, _ := net.SplitHostPort(addr)
		p64, _ := strconv.ParseInt(p, 10, 64)
		return int(p64)
	}

	// localDiscoverer will always return localhost regardless of address
	localDiscoverer := func(network, addr string) (net.Addr, error) {
		switch network {
		case "tcp", "tcp4":
			return &net.TCPAddr{
				IP:   localhost,
				Port: getPort(addr),
			}, nil
		case "udp", "udp4":
			return &net.UDPAddr{
				IP:   localhost,
				Port: getPort(addr),
			}, nil
		case "tcp6":
			return &net.UDPAddr{
				IP:   localhost6,
				Port: getPort(addr),
			}, nil
		case "udp6":
			return &net.UDPAddr{
				IP:   localhost6,
				Port: getPort(addr),
			}, nil
		default:
			return nil, ErrUndiscoverable
		}
	}

	// Dialer will always dial localhost regardless of the host address
	d := NewDialer(DiscovererFunc(func(network, addr string) (net.Addr, error) {
		netaddr, err := localDiscoverer(network, addr)
		if err != nil {
			return nil, err
		}
		fmt.Printf("discover: %s\n", netaddr)
		return netaddr, nil
	}))
	// Use the DialContext of the discovery.Dialer in an HTTP Client.Transport
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: d.DialContext,
		},
	}
	// This will get from localhost:80
	client.Get("http://host.doesnotexist")
	// This will get from localhost:443
	client.Get("https://host.doesnotexist")
	// This should try to connect to unix:/file/does/not/exist as normal (and fail)
	_, err := d.Dial("unix", "/file/does/not/exist")
	fmt.Println(err)
	// Output:
	// discover: 127.0.0.1:80
	// discover: 127.0.0.1:443
	// dial unix /file/does/not/exist: connect: no such file or directory
}

func TestDiscoveryDialer(t *testing.T) {
	port := 8080
	srv := http.Server{
		Addr: "localhost:8080",
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			resp.WriteHeader(200)
			resp.Write([]byte("OK\n"))
		}),
	}
	go func() {
		srv.ListenAndServe()
	}()
	doneCh := make(chan struct{})
	go func() {
		<-doneCh
		srv.Close()
	}()
	defer close(doneCh)
	time.Sleep(100 * time.Millisecond) // Let the server run
	d := NewDialer(DiscovererFunc(func(network, host string) (net.Addr, error) {
		t.Logf("Discover: %s / %s", network, host)
		return &net.TCPAddr{
			IP:   net.IP([]byte{127, 0, 0, 1}),
			Port: port,
		}, nil
	}))
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: d.DialContext,
		},
	}
	resp, err := client.Get("http://host.doesnotexist")
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("BODY: %s", string(body))
}
