package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/sacloud/libsacloud/api"
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
	fmt.Println(config)
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

	// config the disk
	diskConf := client.Disk.NewCondig()
	diskConf.SetHostName(config.HostName)
	diskConf.SetPassword(config.Password)
	diskConf.AddSSHKeyByString(config.SshPublicKey)
	diskConf.AddNote(script.GetStrID())
	client.Disk.Config(disk.ID, diskConf)

	// connect to server
	client.Disk.ConnectToServer(disk.ID, server.ID)

	// boot
	bootServer(server.ID)
}

func findResource() (id int64, ipaddress string) {
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
		Find() // 検索実施

	if err != nil {
		panic(err)
	}
	// No matching
	if res.Total == 0 {
		return 0, "error"
	}

	id = res.Servers[0].Resource.ID
	ipaddress = res.SakuraCloudResourceList.Servers[0].Interfaces[0].IPAddress
	return
}

func bootServer(serverID int64) {
	// authorize
	client := api.NewClient(config.Token, config.Secret, config.Zone)

	// boot
	fmt.Println("booting the server")
	client.Server.Boot(serverID)
}

func stopServer(serverID int64) {
	// authorize
	client := api.NewClient(config.Token, config.Secret, config.Zone)

	// stop
	time.Sleep(3 * time.Second)
	fmt.Println("stopping the server")
	client.Server.Stop(serverID)

	// wait for server to down
	err := client.Server.SleepUntilDown(serverID, 120*time.Second)
	if err != nil {
		fmt.Println("failed")
		os.Exit(1)
	}
}

func main() {
	importConfig()

	serverID, ipaddress := findResource()

	if serverID != 0 {
		bootServer(serverID)
	} else {
		createServer()
		serverID, ipaddress = findResource()
	}
	fmt.Println("serverID is", serverID)
	fmt.Println(ipaddress, "is UP")
	// stopServer(findResource(projectName))
}
