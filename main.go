package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	//	"github.com/bitly/go-simplejson"
)

const (
	VERSION = "0.1.2"
)

type SSHHost struct {
	Host     string
	Port     int
	Username string
	Password string
	CmdFile  string
	Cmd      []string
	Result   string
}
type HostJson struct {
	SshHosts []SSHHost
}

func main() {
	version := flag.Bool("v", false, "show version")
	hosts := flag.String("hosts", "", "host address list")
	cmd := flag.String("cmd", "", "cmds")
	username := flag.String("u", "", "username")
	password := flag.String("p", "", "password")
	port := flag.Int("port", 22, "ssh port")
	cmdFile := flag.String("cmdfile", "", "cmdfile path")
	hostFile := flag.String("hostfile", "", "hostfile path")
	ipFile := flag.String("ipfile", "", "hostfile path")
	//gu
	jsonFile := flag.String("j", "", "Json File Path")
	outTxt := flag.Bool("outTxt", false, "write result into txt")
	timeLimit := flag.Int("t", 30, "max timeout")
	numLimit := flag.Int("n", 20, "max execute number")

	flag.Parse()
	var cmdList []string
	var hostList []string
	var err error

	sshHosts := []SSHHost{}
	var host_Struct SSHHost
	timeout := time.Duration(*timeLimit) * time.Second

	if *version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if *ipFile != "" {
		hostList, err = GetIpList(*ipFile)
		if err != nil {
			log.Println("load hostlist error: ", err)
			return
		}
	}

	if *hostFile != "" {
		hostList, err = Getfile(*hostFile)
		if err != nil {
			log.Println("load hostfile error: ", err)
			return
		}
	}
	if *hosts != "" {
		hostList = strings.Split(*hosts, ";")
	}

	if *cmdFile != "" {
		cmdList, err = Getfile(*cmdFile)
		if err != nil {
			log.Println("load cmdfile error: ", err)
			return
		}
	}
	if *cmd != "" {
		cmdList = strings.Split(*cmd, ";")
	}
	if *jsonFile == "" {
		for _, host := range hostList {
			host_Struct.Host = host
			host_Struct.Username = *username
			host_Struct.Password = *password
			host_Struct.Port = *port
			host_Struct.Cmd = cmdList
			sshHosts = append(sshHosts, host_Struct)
		}
	}
	//gu
	if *jsonFile != "" {
		sshHosts, err = GetJsonFile(*jsonFile)
		if err != nil {
			log.Println("load jsonFile error: ", err)
			return
		}
		for i := 0; i < len(sshHosts); i++ {
			cmdList, err = Getfile(sshHosts[i].CmdFile)
			if err != nil {
				log.Println("load cmdFile error: ", err)
				return
			}
			sshHosts[i].Cmd = cmdList
		}
	}

	chLimit := make(chan bool, *numLimit) //控制并发访问量
	chs := make([]chan string, len(sshHosts))
	limitFunc := func(chLimit chan bool, ch chan string, host SSHHost) {
		dossh(host.Username, host.Password, host.Host, host.Cmd, host.Port, ch)
		<-chLimit
	}
	for i, host := range sshHosts {
		chs[i] = make(chan string, 1)
		chLimit <- true
		go limitFunc(chLimit, chs[i], host)
	}
	for i, ch := range chs {
		fmt.Println(sshHosts[i].Host, " ssh start")
		select {
		case res := <-ch:
			if res != "" {
				fmt.Println(res)
				sshHosts[i].Result += res
			}
		case <-time.After(timeout):
			log.Println("SSH run timeout")
			sshHosts[i].Result += ("SSH run timeout：" + strconv.Itoa(int(*timeLimit)) + "second.")
		}

		fmt.Println(sshHosts[i].Host, " ssh end")
	}

	//gu
	if *outTxt {
		for i := 0; i < len(sshHosts); i++ {
			err = WriteIntoTxt(sshHosts[i])
			if err != nil {
				log.Println("write into txt error: ", err)
				return
			}
		}
	}

}
