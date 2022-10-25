package consulservice

import (
//	"fmt"
        "bytes"
        "crypto/tls"
        "encoding/json"
        "net/http"
	"net"
	"strconv"
	"strings"
	"time"
)

type serviceStruct struct {
	Id		string		`json:"ID"`
	Name		string		`json:"Name"`
	Tags		[]string	`json:"Tags"`
	Addr		string		`json:"Address"`
	Port		int		`json:"Port"`
	Check		checkStruct	`json:"Check"`
}

type checkStruct struct {
	Id		string		`json:"ID"`
	Name		string		`json:"Name"`
	UnregAfter	string		`json:"DeregisterCriticalServiceAfter,omitempty"`
	Interval	string		`json:"Interval,omitempty"`
	Ttl		string		`json:"TTL,omitempty"`
}

func RegisterService(serviceName string, port int, tags []string) bool {
	var service	serviceStruct

	service.Id = strings.Replace(serviceName, "/", ".", -1) + ":" + strconv.Itoa(port)
	service.Name = serviceName
	service.Tags = tags
	service.Addr = getIpAddr()
	service.Port = port
	service.Check.Id = service.Id + ":check"
	service.Check.Name = service.Name + " health check"
	service.Check.Ttl = "60s"

	// PUT	v1/agent/service/register
        tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},}

        client := &http.Client{Transport: tr}

        buf, err := json.Marshal(service)

        req, err := http.NewRequest("PUT", "http://localhost:8500/v1/agent/service/register", bytes.NewBufferString(string(buf)))

        if err != nil {
                return false
        }

	req.Header.Add("Content-Type", "application/json")

        response, err := client.Do(req)

        if err != nil {
                return false
        }
	// fmt.Println(response)

        defer response.Body.Close()

	healthCheck(service.Id)

	return true
}

func healthCheck(serviceId string) {

        tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},}

        client := &http.Client{Transport: tr}

	// PUT	v1/agent/check/pass/:check_id
	req, _ := http.NewRequest("PUT", "http://localhost:8500/v1/agent/check/pass/" + "service:" + serviceId + "?note=passed", nil)

	req.Header.Add("Content-Type", "application/json")

	go func() {
		for {
		        _, err := client.Do(req)
			if err != nil {
				// fmt.Println("error", err)
				break
			}
			// fmt.Println(resp)
			time.Sleep(45 * time.Second)
		}
	}()
}

func UnregisterService(serviceName string, port int) bool {

	serviceId := strings.Replace(serviceName, "/", ".", -1) + ":" + strconv.Itoa(port)

	// PUT	v1/agent/service/deregister/:service.id
        tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},}

        client := &http.Client{Transport: tr}

        req, err := http.NewRequest("PUT", "http://localhost:8500/v1/agent/service/deregister/" + serviceId, nil)

	req.Header.Add("Content-Type", "application/json")

        if err != nil {
                return false
        }

        response, err := client.Do(req)

	// fmt.Println("unregister", response)
        if err != nil {
                return false
        }

        defer response.Body.Close()

	return true
}

func getIpAddr() string {
	ipAddr := ""

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ipAddr
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipAddr = ipnet.IP.String()
			}
		}
	}

	return ipAddr
}
