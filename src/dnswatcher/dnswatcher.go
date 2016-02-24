package main

import (
	"bytes"
	"config"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"text/template"
	"time"
	"utils"

	"github.com/golang/glog"
)

// Process structure to store Process information
type Process struct {
	config          *config.Config
	auth            smtp.Auth
	template        *template.Template
	previousIP      net.IP
	pid             string
	processHostname string
}

// NewProcess returns new process instance
func NewProcess(c config.Interface) utils.ProcessInterface {
	return &Process{config: c.(*config.Config)}
}

// Init the process resources
func (p *Process) Init() error {
	p.auth = smtp.PlainAuth("", p.config.User, p.config.Password, p.config.SMTPHostname)
	p.template = template.Must(template.New("email").Parse(emailTmpl))
	p.pid = strconv.Itoa(os.Getpid())
	p.processHostname = os.Getenv("HOSTNAME")
	return nil
}

// Clear clear the Process resources
func (p *Process) Clear() {

}

// Start process
func (p *Process) Start(stop chan utils.Stop) error {
	ticker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			glog.Info("checkDNS")
			err := p.checkDNS()
			if err != nil {
				glog.Error("checkDNS err:", err)
			}
		case <-stop:
			glog.Info("Stop")
			return nil
		}
	}
}

func (p *Process) checkDNS() error {
	ips, err := net.LookupIP(p.config.Hostname)
	if err != nil {
		return err
	}

	if len(ips) == 0 {
		return errors.New("IPs contains no or too-many ip")
	}

	if !ips[0].Equal(p.previousIP) {
		if err := p.sendEmail(ips[0]); err != nil {
			return err
		}
		p.previousIP = ips[0]
	}

	return nil
}

func (p *Process) sendEmail(newIP net.IP) error {
	data := map[string]interface{}{
		"From":            "DNSWatcher",
		"Email":           p.config.Email,
		"User":            p.config.User,
		"HostName":        p.config.Hostname,
		"PreviousIP":      p.previousIP,
		"NewIp":           newIP,
		"Timestamp":       time.Now(),
		"Pid":             p.pid,
		"ProcessHostname": p.processHostname,
	}

	header := make(map[string]string)
	header["From"] = "DNSWatcher"
	header["To"] = p.config.Email
	header["Subject"] = string("DNSWatcher " + p.config.Hostname + " has changed")
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	buf := &bytes.Buffer{}
	if err := p.template.Execute(buf, data); err != nil {
		return err
	}

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	message += "\r\n" + base64.StdEncoding.EncodeToString(buf.Bytes())
	glog.Info("Email:", message)

	err := smtp.SendMail(p.config.SMTPHostname+":587", p.auth, "DNSWatcher", []string{p.config.Email}, []byte(message))
	if err != nil {
		return err
	}
	return nil
}

const emailTmpl = `
Hi {{.User}}!

The Hostname: {{.HostName}} points to a new ip address

Previous IP: {{.PreviousIP}}
New IP: {{.NewIp}}
Timestamp: {{.Timestamp}}
Process Hostname: {{.ProcessHostname}}
Pid: {{.Pid}}

Regards
DNSWatcher
`
