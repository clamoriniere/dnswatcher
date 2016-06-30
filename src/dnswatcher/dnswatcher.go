package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"config"
	"utils"

	units "github.com/docker/go-units"
	"github.com/golang/glog"
)

// Process structure to store Process information
type Process struct {
	config          *config.Config
	auth            smtp.Auth
	template        *template.Template
	pid             string
	processHostname string
	address         []*AddressWatcher
	userInfos       []UserInfo
}

// AddressWatcher store information about an address
type AddressWatcher struct {
	Address      string
	PreviousIP   *utils.TimerSet
	PreviousTime time.Time
}

// UserInfo use to store user information
type UserInfo struct {
	Email string
	Name  string
}

// NewAddressWatcher returns new AddressWatcher instance
func NewAddressWatcher(address string) *AddressWatcher {
	return &AddressWatcher{
		Address:      address,
		PreviousIP:   utils.NewTimerSet(time.Duration(5 * time.Minute)),
		PreviousTime: time.Now(),
	}
}

// NewProcess returns new process instance
func NewProcess(c config.Interface) utils.ProcessInterface {
	return &Process{config: c.(*config.Config), address: []*AddressWatcher{}}
}

// Init the process resources
func (p *Process) Init() error {
	glog.Info("Init Start")
	p.auth = smtp.PlainAuth("", p.config.User, p.config.Password, p.config.SMTPHostname)
	p.template = template.Must(template.New("email").Parse(emailTmpl))
	p.pid = strconv.Itoa(os.Getpid())
	p.processHostname, _ = os.Hostname()

	for _, addr := range strings.Split(p.config.Hostname, ",") {
		glog.Info("- Add addr:" + addr)
		p.address = append(p.address, NewAddressWatcher(addr))
	}

	for _, email := range strings.Split(p.config.Email, ",") {
		glog.Info("- Add email:" + email)
		p.userInfos = append(p.userInfos, UserInfo{Email: email, Name: email})
	}
	glog.Info("Init Stop")
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

func (p *Process) checkDNSAddress(addr *AddressWatcher) error {
	ips, err := net.LookupIP(addr.Address)
	if err != nil {
		return err
	}

	if len(ips) == 0 {
		return errors.New("IPs contains no or too-many ip")
	}

	if !addr.PreviousIP.FindIP(ips[0]) {
		if err := p.sendEmails(addr, ips[0]); err != nil {
			return err
		}
		addr.PreviousTime = time.Now()
	}
	addr.PreviousIP.AddIP(ips[0], time.Now())
	return nil
}

func (p *Process) checkDNS() error {

	for _, addr := range p.address {
		p.checkDNSAddress(addr)
	}

	return nil
}

func (p *Process) sendEmails(addr *AddressWatcher, newIP net.IP) error {
	for _, user := range p.userInfos {
		if err := p.sendEmail(addr, user, newIP); err != nil {
			return err
		}
	}

	return nil
}

func (p *Process) sendEmail(addr *AddressWatcher, user UserInfo, newIP net.IP) error {
	data := map[string]interface{}{
		"From":            "DNSWatcher",
		"Email":           user.Email,
		"User":            user.Name,
		"HostName":        addr.Address,
		"Duration":        units.HumanDuration(time.Since(addr.PreviousTime)),
		"PreviousIP":      addr.PreviousIP.Front().String(),
		"NewIp":           newIP,
		"Timestamp":       time.Now(),
		"Pid":             p.pid,
		"ProcessHostname": p.processHostname,
	}

	header := make(map[string]string)
	header["From"] = "DNSWatcher"
	header["To"] = user.Email
	header["Subject"] = string("DNSWatcher " + addr.Address + " has changed")
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

	err := smtp.SendMail(p.config.SMTPHostname+":587", p.auth, "DNSWatcher", []string{user.Email}, []byte(message))
	if err != nil {
		return err
	}
	return nil
}

const emailTmpl = `
Hi {{.User}}!

The Hostname: {{.HostName}} points to a new ip address

Previous IP: {{.PreviousIP}} since {{.Duration}}
New IP: {{.NewIp}}
Timestamp: {{.Timestamp}}
Process Hostname: {{.ProcessHostname}}
Pid: {{.Pid}}

Regards
DNSWatcher
`
