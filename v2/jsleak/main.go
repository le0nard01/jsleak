package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"strings"

	"github.com/gijsbers/go-pcre"
)

type JsonReturn struct {
	Url     string
	Pattern string
	Match   string
}

func getLeak(url string, data string, pattern string, jsonArray *[]JsonReturn) {
	re := pcre.MustCompile(pattern, 0)
	matches := re.MatcherString(data, 0).Group(0)
	//fmt.Println(len(matches))
	if len(matches) != 0 {
		fmt.Printf("[+] Url: %v\n[+] Pattern: %v\n[+] Match: %v\n", url, pattern, string(matches))
		jsn := JsonReturn{url, pattern, string(matches)}
		*jsonArray = append(*jsonArray, jsn)
	}
}

func get_inputs() []string {
	reader := bufio.NewReader(os.Stdin)
	var output []rune

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}

	return strings.Fields(string(output))
}

func req(url string, timeout int) string {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{
		Transport: transCfg,
		Timeout: time.Duration(timeout) * time.Second,
	}
	res, err := client.Get(url)

	if err != nil {
		log.Fatal(err)
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return string(data)
}

func main() {
	path := flag.String("pattern", "", "[+] File contains patterns to test")
	verbose := flag.Bool("verbose", false, "[+] Verbose Mode")
	jsonOutput := flag.String("json", "", "[+] Json output file")
	timeout := flag.Int("timeout", 5, "[+] Timeout for request in seconds")
	flag.Parse()
	
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("[+] Use in Pipeline")
		os.Exit(1)
	}

	file, err := os.Open(*path)
	defer file.Close()
	lines := make([]string, 0)

	patterns := bufio.NewScanner(file)
	jsonArray := make([]JsonReturn, 1)
	for patterns.Scan() {
		lines = append(lines, patterns.Text())
	}

	if err != nil {
		log.Fatal(err)
	}

	for _, url := range get_inputs() {
		if *verbose {
			fmt.Println("[-] Looking: " + url)
		}

		data := req(url,*timeout)

		for _, pattern := range lines {
			getLeak(url, data, pattern, &jsonArray)
		}
	}

	if *jsonOutput != "" {
		fo, err2 := os.Create(*jsonOutput)
		k, err1 := json.MarshalIndent(jsonArray, "", "\t")
		if _, err := fo.Write(k); err1 != nil || err2 != nil {
			panic(err)
		}
	}
}
