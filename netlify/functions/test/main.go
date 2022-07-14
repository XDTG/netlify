package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var result string

func printEnvs() {
	envs := os.Environ()
	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		} else {
			result += fmt.Sprintf("%s = %s\n", parts[0], parts[1])
		}
	}
	result += "\n\n"
}

func readResolv() {
	filepath := "/etc/resolv.conf"
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println("failed to read resolv.conf!")
		panic(err)
	}
	result += string(content)
	result += "\n\n"
}

func digServices() {
	cmd := exec.Command("dig", "baidu.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("failed to exec \"dig\" command!")
		result += "non dig info\n"
	} else {
		result += string(out)
	}

	result += "\n\n"
}

func printServices() {
	readResolv()
	//digServices()
}

func coreMetrics(dnsIp string) {
	if dnsIp == "" {
		return
	}

	url := "http://" + dnsIp + ":9153/metrics"
	fmt.Println("request url: " + url)

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("failed to get dns metrics")
		fmt.Println(err)
		result += "not get \"" + url + "\" (" + err.Error() + ")\n"
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	result += "CoreDNS response: \n" + string(body) + "\n"
	result += "StatusCode: " + string(rune(resp.StatusCode)) + "\n"
}
func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	result = ""

	printEnvs()

	printServices()

	dnsIp := "169.254.78.1"
	coreMetrics(dnsIp)

	return &events.APIGatewayProxyResponse{
		StatusCode:        200,
		Headers:           map[string]string{"Content-Type": "text/plain"},
		MultiValueHeaders: http.Header{"Set-Cookie": {"Ding", "Ping"}},
		Body:              result,
		IsBase64Encoded:   false,
	}, nil
}

func main() {
	// Make the handler available for Remote Procedure Call
	lambda.Start(handler)
}
