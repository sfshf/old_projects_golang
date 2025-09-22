package simplehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	gateway_api "github.com/nextsurfer/oracle/api/gateway"
	"github.com/nextsurfer/oracle/api/response"
	"google.golang.org/grpc"
)

var (
	ErrResponseStatusCodeNotEqualTo200 = errors.New("http response status code not equal to 200")
	ErrResponseDataCodeNotEqualToZero  = errors.New("http response data code not equal to 0")
)

func Get(url string, headers map[string]string, respData interface{}) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if respData != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, respData); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func PostJsonRequest(location string, reqData interface{}, cookie *http.Cookie, respData interface{}) (*http.Response, error) {
	var body io.Reader
	if reqData != nil {
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}
	req, err := http.NewRequest(http.MethodPost, location, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if respData != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, respData); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func UnmarshalRequestBody(r *http.Request, req interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, req); err != nil {
		return err
	}
	return nil
}

func GetRealIP(r *http.Request) string {
	var realIP string
	if realIP = r.Header.Get("x-forwarded-for"); realIP != "" {
		return realIP
	}
	// ...
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func UserAgent(r *http.Request) string {
	var res string
	for key, value := range r.Header {
		if strings.ToLower(key) == "user-agent" {
			for k, v := range value {
				res += v
				if k != len(value)-1 {
					res += ";;"
				}
			}
			break
		}
	}
	return res
}

func NotifyGatewayRefreshService(gatewayAppIpv4 string, gatewayGrpcPort int32, serviceName string) error {
	addr := fmt.Sprintf("%s:%d", gatewayAppIpv4, gatewayGrpcPort)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	gatewayServiceClient := gateway_api.NewGatewayServiceClient(conn)
	respData, err := gatewayServiceClient.RefreshService(context.Background(), &gateway_api.RefreshServiceRequest{
		Name: serviceName,
	})
	if err != nil {
		return err
	}
	if respData.Code != response.StatusCodeOK {
		return fmt.Errorf("notify gateway [addr=%s] refresh service failed", addr)
	}
	return nil
}

func NotifyGatewayRefreshCertificate(gatewayAppIpv4 string, gatewayGrpcPort int32, domain string) error {
	addr := fmt.Sprintf("%s:%d", gatewayAppIpv4, gatewayGrpcPort)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	gatewayServiceClient := gateway_api.NewGatewayServiceClient(conn)
	respData, err := gatewayServiceClient.RefreshCertificate(context.Background(), &gateway_api.RefreshCertificateRequest{
		Domain: domain,
	})
	if err != nil {
		return err
	}
	if respData.Code != response.StatusCodeOK {
		return fmt.Errorf("call gateway [addr=%s] RefreshCertificate failed", addr)
	}
	return nil
}

func NotifyGatewayRefreshProxy(gatewayAppIpv4 string, gatewayGrpcPort int32, domain string) error {
	addr := fmt.Sprintf("%s:%d", gatewayAppIpv4, gatewayGrpcPort)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	gatewayServiceClient := gateway_api.NewGatewayServiceClient(conn)
	respData, err := gatewayServiceClient.RefreshProxy(context.Background(), &gateway_api.RefreshProxyRequest{
		Domain: domain,
	})
	if err != nil {
		return err
	}
	if respData.Code != response.StatusCodeOK {
		return fmt.Errorf("call gateway [addr=%s] RefreshCertificate failed", addr)
	}
	return nil
}

func GetLocalIPv4() string {
	ips, _ := GetLocalIPv4s()
	if len(ips) > 0 {
		return ips[0]
	}
	return ""
}

func GetLocalIPv4s() ([]string, error) {
	ips := make([]string, 0)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}
	return ips, nil
}
