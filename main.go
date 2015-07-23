package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/paulstuart/sshclient"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	username = os.Getenv("SSH_USERNAME")
	password = os.Getenv("SSH_PASSWORD")
	works    = make(map[string]string)
	hosts    = make([]string, 0, 1024)
	timeout  = 5
	ipFile   string
	wg       sync.WaitGroup
	cmd      string
)

func init() {
	flag.StringVar(&ipFile, "f", "", "file containing list of IPs to run against")
	flag.IntVar(&timeout, "t", timeout, "timeout period")
}

func main() {
	flag.Parse()
	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(1)
	}
	cmds := make([]string, flag.NArg())

	for i := 0; i < flag.NArg(); i++ {
		cmds[i] = flag.Arg(i)
	}
	cmd = strings.Join(cmds, " ")
	if _, err := os.Stat(ipFile); err != nil {
		flag.Usage()
		log.Fatal("File not found:", err)
	}
	file, err := os.Open(ipFile)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		host := strings.TrimSpace(scanner.Text())
		if len(host) > 0 {
			hosts = append(hosts, host)
		}
	}

	if len(hosts) == 0 {
		flag.Usage()
		log.Fatal("No host specified")
	}

	if len(username) == 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Username: ")
		var err error
		username, err = reader.ReadString('\n')
		if err != nil {
			log.Println("read error:", err)
		}
		username = strings.TrimSpace(username)
	}

	if len(password) == 0 {
		fmt.Print("Password: ")
		if pw, err := terminal.ReadPassword(int(os.Stdin.Fd())); err == nil {
			password = strings.TrimSpace(string(pw))
		}
		fmt.Println()
	}

	for _, host := range hosts {
		wg.Add(1)
		go tryHost(host)
	}
	wg.Wait()
}

func tryHost(host string) {
	defer wg.Done()
	_, stdout, _, err := sshclient.Exec(host+":22", username, password, cmd, timeout)
	if err == nil {
		stdout = strings.TrimSpace(stdout)
	} else {
		stdout = err.Error()
	}
	for _, line := range strings.Split(stdout, "\n") {
		fmt.Println(host, line)
	}
}
