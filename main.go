package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/sacloud/libsacloud/api"
	"github.com/taroooyan/confirm"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
	"time"
)

type Config struct {
	Token        string `toml:"token"`
	Secret       string `toml:"secret"`
	Zone         string `toml:"zone"`
	Name         string `toml:"name"`
	Description  string `toml:"description"`
	Tag          string `toml:"tag"`
	Cpu          int    `toml:"cpu"`
	Mem          int    `toml:"mem"`
	HostName     string `toml:"hostName"`
	Password     string `toml:"password"`
	SshPublicKey string `toml:"sshPublicKey"`
}

type Server struct {
	ServerId  int64
	Ipaddress string
	DiskId    int64
}

var config Config

func importConfig() {
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		fmt.Println("config import err")
		panic(err)
	}

	// set to Name or HostName is deirectory name
	pwd, _ := os.Getwd()
	tmp := strings.Split(pwd, "/")
	projectName := tmp[len(tmp)-1]
	if config.Name == "" {
		config.Name = projectName
	}
	if config.HostName == "" {
		config.HostName = projectName
	}
}

func findResource() (server Server) {
	// authorize
	client := api.NewClient(config.Token, config.Secret, config.Zone)

	// サーバーの検索
	res, err := client.Server.
		WithNameLike(config.HostName). // サーバー名に"server name"が含まれる
		Offset(0).                     // 検索結果の位置0(先頭)から取得
		Limit(5).                      // 5件取得
		Include("Name").               // 結果にName列を含める
		Include("Description").        // 結果にDescription列を含める
		Include("Interfaces.IPAddress").
		Include("Disks").
		Find() // 検索実施

	if err != nil {
		panic(err)
	}
	// No matching
	if res.Total == 0 {
		return
	}
	server.ServerId = res.Servers[0].Resource.ID
	server.Ipaddress = res.SakuraCloudResourceList.Servers[0].Interfaces[0].IPAddress
	server.DiskId = res.SakuraCloudResourceList.Servers[0].Disks[0].Resource.ID
	return
}

func createServer() {
	// authorize
	client := api.NewClient(config.Token, config.Secret, config.Zone)

	// search archives
	fmt.Println("searching archives")
	archive, _ := client.Archive.FindLatestStableCentOS()

	// search scripts
	fmt.Println("searching scripts")
	res, _ := client.Note.
		WithNameLike("WordPress").
		WithSharedScope().
		Limit(1).
		Find()
	script := res.Notes[0]

	// create a disk
	fmt.Println("creating a disk")
	disk := client.Disk.New()
	disk.Name = config.HostName
	disk.Description = config.Description
	disk.Tags = []string{config.Tag}
	disk.SetDiskPlanToSSD()
	disk.SetSourceArchive(archive.ID)

	disk, _ = client.Disk.Create(disk)

	// create a server
	fmt.Println("creating a server")
	server := client.Server.New()
	server.Name = config.HostName
	server.Description = config.Description
	server.Tags = []string{config.Tag}

	// set ServerPlan
	plan, _ := client.Product.Server.GetBySpec(config.Cpu, config.Mem)
	server.SetServerPlanByID(plan.GetStrID())

	server, _ = client.Server.Create(server)

	// connect to shared segment
	fmt.Println("connecting the server to shared segment")
	iface, _ := client.Interface.CreateAndConnectToServer(server.ID)
	client.Interface.ConnectToSharedSegment(iface.ID)

	// wait disk copy
	err := client.Disk.SleepWhileCopying(disk.ID, 120*time.Second)
	if err != nil {
		fmt.Println("failed")
		os.Exit(1)
	}

	// config the disk
	diskConf := client.Disk.NewCondig()
	diskConf.SetHostName(config.HostName)
	diskConf.SetPassword(config.Password)
	diskConf.AddSSHKeyByString(config.SshPublicKey)
	diskConf.AddNote(script.GetStrID())
	client.Disk.Config(disk.ID, diskConf)

	// connect to server
	client.Disk.ConnectToServer(disk.ID, server.ID)
}

func bootServer(serverId int64) {
	// authorize
	client := api.NewClient(config.Token, config.Secret, config.Zone)

	// boot
	fmt.Println("booting the server")
	client.Server.Boot(serverId)
}

func stopServer(serverId int64) {
	// authorize
	client := api.NewClient(config.Token, config.Secret, config.Zone)

	// stop
	time.Sleep(3 * time.Second)
	fmt.Println("stopping the server")
	client.Server.Stop(serverId)

	// wait for server to down
	err := client.Server.SleepUntilDown(serverId, 120*time.Second)
	if err != nil {
		fmt.Println("failed")
		os.Exit(1)
	}
}

func delServer(serverId int64, diskId int64) {
	// authorize
	client := api.NewClient(config.Token, config.Secret, config.Zone)

	// disconnect the disk from the server
	fmt.Println("disconnecting the disk")
	client.Disk.DisconnectFromServer(diskId)

	// delete the server
	fmt.Println("deleting the server")
	client.Server.Delete(serverId)

	// delete the disk
	fmt.Println("deleting the disk")
	client.Disk.Delete(diskId)
}

func main() {
	importConfig()
	server := findResource()

	var boot = flag.Bool("boot", false, "boot server")
	var stop = flag.Bool("stop", false, "stop server")
	var del = flag.Bool("delete", false, "delete server")
	var create = flag.Bool("create", false, "create new server")
	var show = flag.Bool("show", false, "show server")
	var ssh = flag.Bool("ssh", false, "ssh connect server")
	flag.Parse()

	if *create == true {
		if server.ServerId == 0 {
			createServer()
			server = findResource()
		}
	}

	if *boot == true {
		bootServer(server.ServerId)
		fmt.Println("serverID(", server.ServerId, ") is UP")
	}

	if *stop == true {
		stopServer(server.ServerId)
		fmt.Println("serverID(", server.ServerId, ") is DOWN")
	}

	if *del == true {
		message := "Is is okay to delete this server?[y/n]"
		if confirm.AskConfirm(message) {
			delServer(server.ServerId, server.DiskId)
			fmt.Println("serverID(", server.ServerId, ") is DELETED")
		}
	}

	if *show == true {
	}

	if *ssh == true {
		access("root", server.Ipaddress+":22", config.Password)
	}

	fmt.Println(server.Ipaddress, server.DiskId)
}

func access(user, host, password string) {
	client, session, err := connectToHost(user, host, password)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	modes := ssh.TerminalModes{
		// ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		return
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin
	session.Run("bash")
	defer session.Close()
}

func connectToHost(user, host, password string) (*ssh.Client, *ssh.Session, error) {
	// fmt.Print("Password: ")
	// pass, err := terminal.ReadPassword(int(syscall.Stdin))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}
