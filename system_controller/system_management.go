package main

import (
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const delim string = "_"

func main() {
	type OPTS struct {
		start, stop, reset, setup, getlogs bool
		seed                               int64
		readrate, writerate, filesize      float64
		status                             bool
	}

	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Name = "Controller for the atomicity setup"

	app.Commands = []cli.Command{
		{
			Name:   "setup",
			Usage:  "sets up reader, writers, servers and a controller",
			Action: setup,
		},
		{
			Name:   "status",
			Usage:  "shows the status of the system",
			Action: status,
		},
		{
			Name: "setread_dist",
			Usage: "the inter read wait time distribution (i) \"erlang k m\",\n" +
				"                   k is shape, and m is scale parameter (inverse of rate)\n" +
				"                   (ii) \"const k\" k is the inter read wait time in milliseconds\n",
			Action: setReadRateDistribution,
		},
		{
			Name: "setwrite_dist",
			Usage: "the inter write wait time distribution (i) \"erlang k m\",\n" +
				"                   k is shape, and m is scale parameter (inverse of rate)\n" +
				"                   (ii) \"const k\" k is the inter read wait time in milliseconds\n",
			Action: setWriteRateDistribution,
		},
		{
			Name:   "setwriteto",
			Usage:  "\"disk\" or \"mem\" either write to disk or memory",
			Action: setWriteTo,
		},
		{
			Name:   "setfile_size",
			Usage:  "const file size in KB",
			Action: setFileSize,
		},

		{
			Name:   "stop",
			Usage:  "pause reads and writes",
			Action: stop,
		},
		{
			Name:   "start",
			Usage:  "pause reads and writes",
			Action: start,
		},

		{
			Name:   "reset",
			Usage:  "reset the random generator",
			Action: reset,
		},
		{
			Name:   "setseed",
			Usage:  "set the random number seed provided",
			Action: setseed,
		},
		{
			Name:   "getparams",
			Usage:  "show the parameters set for the run",
			Action: getParameters,
		},
		{
			Name:   "setinitvaluesize",
			Usage:  "-----sets the intial value size stored in the data",
			Action: setInitialValueSize,
		},
		{
			Name:   "setrunid",
			Usage:  "sets the run id",
			Action: setRunId,
		},
		{
			Name:   "getlogs",
			Usage:  "get the logs from the clients and servers to the controller",
			Action: getlogs,
		},
		{
			Name:   "flushlogs",
			Usage:  "flush the logs from the clients and servers to the controller",
			Action: flushlogs,
		},
	}

	app.Run(os.Args)
}

// shows the status of the systems
func status(c *cli.Context) error {

	readers, writers, servers, controllers := getIPAddresses()

	fmt.Println("Number readers : ", len(readers))
	for _, value := range readers {
		fmt.Println("    ", value)
	}

	fmt.Println("Number writers : ", len(writers))

	for _, value := range writers {
		fmt.Println("    ", value)
	}

	fmt.Println("Number servers : ", len(servers))

	for _, value := range servers {
		fmt.Println("    ", value)
	}

	fmt.Println("Number controller : ", len(controllers))

	for _, value := range controllers {
		fmt.Println("    ", value)
	}
	return nil
}

// stop the reader, writers and servers
func stop(c *cli.Context) error {

	if !isSystemRunning() {
		return nil
	}
	//readers, writers, _, controllers := getIPAddresses()
	readers, writers, _, _ := getIPAddresses()

	for _, ipaddr := range readers {
		fmt.Println("reader", ipaddr, "StopProcess")
		ipaddrs := make([]string, 1)
		ipaddrs[0] = ipaddr
		sendCommandToControllers(ipaddrs, "StopProcess", "")

	}
	for _, ipaddr := range writers {
		fmt.Println("writer", ipaddr, "StopProcess")
		ipaddrs := make([]string, 1)
		ipaddrs[0] = ipaddr
		sendCommandToControllers(ipaddrs, "StopProcess", "")

	}

	//sendCommandToControllers(controllers, "StopReaders", "")
	//sendCommandToControllers(controllers, "StopWriters", "")
	//sendCommandToControllers(controllers, "StopServers", "")
	return nil
}

// start the reader, writers and servers
func start(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}

	_, _, _, controllers := getIPAddresses()
	sendCommandToControllers(controllers, "StartReaders", "")
	sendCommandToControllers(controllers, "StartWriters", "")
	//	sendCommandToControllers(z, "StartServers", "")
	return nil
}

//set the read rate
func setReadRateDistribution(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}

	_, _, _, controllers := getIPAddresses()
	var rateParametersString string
	if len(c.Args()) == 0 {
		fmt.Println("No distribution provided")
		return nil
	}
	rateParametersString = c.Args().First()
	for i := 1; i < len(c.Args()); i++ {
		rateParametersString += "_" + c.Args()[i]
	}
	sendCommandToControllers(controllers, "SetReadRateDistribution", rateParametersString)

	return nil
}

//set the write rate distribution
func setWriteRateDistribution(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}
	_, _, _, controllers := getIPAddresses()

	var rateParametersString string

	if len(c.Args()) == 0 {
		fmt.Println("No distribution provided")
		return nil
	}

	rateParametersString = c.Args().First()
	for i := 1; i < len(c.Args()); i++ {
		rateParametersString += "_" + c.Args()[i]
	}
	sendCommandToControllers(controllers, "SetWriteRateDistribution", rateParametersString)
	return nil
}

//set the write option either to disk or mem (virtual memory)
func setWriteTo(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}
	_, _, _, controllers := getIPAddresses()

	writeToOption := c.Args().First()
	sendCommandToControllers(controllers, "SetWriteTo", writeToOption)
	return nil
}

//set the write rate
/*
func setWriteRate(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}
	rate := c.Args().First()
	_, _, _, controllers := getIPAddresses()
	rateString := rate
	sendCommandToControllers(controllers, "SetWriteRate", rateString)
	return nil
}
*/

//set the file size
func setFileSize(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}
	size := c.Args().First()
	_, _, _, controllers := getIPAddresses()
	sizeString := size

	sendCommandToControllers(controllers, "SetFileSize", sizeString)
	return nil
}

//set the seed
func setseed(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}
	seed := c.Args().First()
	_, _, _, controllers := getIPAddresses()
	seedString := fmt.Sprintf("%d", seed)
	sendCommandToControllers(controllers, "SetSeed", seedString)

	sendCommandToControllers(controllers, "GetSeed", "")
	return nil

}

//set the seed
func getlogs(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}

	folder := c.Args().First()
	_, err := os.Stat(folder)
	if os.IsNotExist(err) {
		err = os.Mkdir(folder, 0700)
		if err != nil {
			log.Fatal(err)
		}
	}

	readers, writers, servers, _ := getIPAddresses()

	// pulll logs from the readers
	for _, e := range readers {
		_name := getName(e)
		name := strings.TrimSpace(_name)
		logstr := getLogFile(e)
		f, err := os.Create(folder + "/" + name + ".log")
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.WriteString(logstr)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()
	}

	// pulll logs from the writers
	for _, e := range writers {
		_name := getName(e)
		name := strings.TrimSpace(_name)
		logstr := getLogFile(e)
		f, err := os.Create(folder + "/" + name + ".log")
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.WriteString(logstr)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()
	}

	// pulll logs from the servers
	for _, e := range servers {
		_name := getName(e)
		name := strings.TrimSpace(_name)
		logstr := getLogFile(e)
		f, err := os.Create(folder + "/" + name + ".log")
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.WriteString(logstr)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()
	}

	return nil
}

//set the seed
func flushlogs(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}
	_, _, _, controllers := getIPAddresses()
	sendCommandToControllers(controllers, "FlushLogs", "")
	return nil
}

//reset the seed
func reset(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}
	_, _, _, controllers := getIPAddresses()
	sendCommandToControllers(controllers, "Reset", "")
	return nil
}

func isSystemRunning() bool {
	_, _, _, controllers := getIPAddresses()

	if len(controllers) == 0 {
		fmt.Println("System does not seem to be up!")
		return false
	}
	return true
}

//get the parameters from controller
func getParameters(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}
	_, _, _, controllers := getIPAddresses()

	params := sendCommandToControllers(controllers, "GetParams", "")
	fmt.Println(params)

	return nil
}

//set the intial value size
func setInitialValueSize(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}
	fmt.Println("yet to implement")
	_, _, _, controllers := getIPAddresses()

	sendCommandToControllers(controllers, "SetInitialValueSize", "")
	return nil
}

//set the run Id
func setRunId(c *cli.Context) error {
	if !isSystemRunning() {
		return nil
	}
	runid := c.Args().First()
	_, _, _, controllers := getIPAddresses()

	sendCommandToControllers(controllers, "SetRunId", runid)
	return nil
}
func setup(c *cli.Context) error {
	fmt.Println("Discovering readers, writers, servers and controllers locally\n")
	readers, writers, servers, controllers := getIPAddresses()
	fmt.Println("Readers      :", readers)
	fmt.Println("Writers      :", writers)
	fmt.Println("Servers      :", servers)
	fmt.Println("Controller/s :", controllers)

	// setup servers to controllers
	fmt.Println("Setting up the Servers\n")

	// send servers to controllers
	serverIPsStack := joinIPs(servers)
	sendCommandToControllers(controllers, "SetServers", serverIPsStack)

	// send reader ids to controllers; and then controllers sends servers ids  to readers
	fmt.Println("Setting up the Readers\n")
	readerIPsStack := joinIPs(readers)
	sendCommandToControllers(controllers, "SetReaders", readerIPsStack)

	// send writer ids to controllers; and then controllers sends servers ids  to writers
	fmt.Println("Setting up the writers\n")
	writerIPsStack := joinIPs(writers)
	sendCommandToControllers(controllers, "SetWriters", writerIPsStack)

	if len(controllers) == 0 {
		return errors.New("Invalid number of controllers")
	}
	fmt.Println("Setting up the controller/s\n")
	return nil
}

func getLogFile(ip string) (logs string) {
	url := "http://" + ip + ":8080" + "/GetLog"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	logs = string(contents)

	return logs
}

func getName(ip string) (name string) {
	url := "http://" + ip + ":8080" + "/GetName"
	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}
	name = string(contents)
	return
}

func sendCommandToControllers(controllers []string, route string, queryStr string) (ack string) {
	var url string

	if len(controllers) < 1 {
		log.Fatal("No available controllers to send commands to. Exiting...")
	}

	if len(queryStr) > 0 {
		url = "http://" + controllers[0] + ":8080" + "/" + route + "/" + queryStr
	} else {
		url = "http://" + controllers[0] + ":8080" + "/" + route
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err != nil {
		log.Fatal(err)
	}
	ack = string(contents)
	return ack
}

func joinIPs(ips []string) string {

	ipstr := ""
	for i, ip := range ips {
		if i > 0 {
			ipstr = ipstr + delim + ip
		} else {
			ipstr = ipstr + ip
		}
	}
	return ipstr
}

func getIPAddresses() ([]string, []string, []string, []string) {
	re := regexp.MustCompile("\"IPAddress\": \"[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+\"")
	ipaddrpat := regexp.MustCompile("[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+")

	_out, err := exec.Command("/usr/bin/docker", "ps", "-q").Output()

	out := string(_out)

	var readers []string
	var writers []string
	var servers []string
	var controllers []string

	if err != nil {
		log.Fatal(err)
	}
	ids := strings.Split(string(out), "\n")

	for _, id := range ids {
		if len(string(id)) == 0 {
			continue
		}

		_out, err := exec.Command("/usr/bin/docker", "inspect", string(id)).Output()
		out = string(_out)

		var ipaddr string
		ipaddrs := re.FindAllString(string(out), 1)
		if len(ipaddrs) > 0 {
			_ipaddr := ipaddrpat.FindAllString(string(ipaddrs[0]), 1)
			ipaddr = string(_ipaddr[0])
		}

		_matched0, err := regexp.MatchString("reader", string(out))
		matched0 := bool(_matched0)
		if err != nil {
			log.Fatal(err)
		}

		matched1, err := regexp.MatchString("writer", string(out))
		if err != nil {
			log.Fatal(err)
		}

		matched2, err := regexp.MatchString("server", string(out))
		if err != nil {
			log.Fatal(err)
		}

		matched3, err := regexp.MatchString("_controller_", string(out))
		if err != nil {
			log.Fatal(err)
		}

		if matched0 {
			readers = append(readers, ipaddr)
		} else if matched1 {
			writers = append(writers, ipaddr)
		} else if matched2 {
			servers = append(servers, ipaddr)
		} else if matched3 {
			controllers = append(controllers, ipaddr)
		}
	}
	return readers, writers, servers, controllers
}
/*
./system_management start

writer error

Running http server
INFO	Set Servers
 servers   1
INFO	SetName
writer_0
Expected 1 parameters, found 2
writer_0
INFO	StartProcess called
fatal error: unexpected signal during runtime execution
[signal 0xb code=0x1 addr=0x0 pc=0x7f41388130cc]

runtime stack:
runtime.throw(0x854ba0, 0x2a)
	/home/docker/go/src/runtime/panic.go:547 +0x90
runtime.sigpanic()
	/home/docker/go/src/runtime/sigpanic_unix.go:12 +0x5a

goroutine 1 [syscall, locked to thread]:
runtime.cgocall(0x6b9300, 0xc82004baa0, 0x0)
	/home/docker/go/src/runtime/cgocall.go:123 +0x11b fp=0xc82004ba50 sp=0xc82004ba20
_/home/docker/COLAS/src/daemons._Cfunc_SODAW_write(0x7f41300008c0, 0x7f4130000ac0, 0x7f4100000000, 0x7f4130000a20, 0x88, 0x7f4130000ae0, 0x7f4130000b00, 0x0)
	??:0 +0x41 fp=0xc82004baa0 sp=0xc82004ba50
_/home/docker/COLAS/src/daemons.writer_deamon()
	/home/docker/COLAS/src/daemons/writer.go:72 +0x72a fp=0xc82004bd50 sp=0xc82004baa0
_/home/docker/COLAS/src/daemons.Writer_process(0xc820072f00)
	/home/docker/COLAS/src/daemons/writer.go:102 +0x1ea fp=0xc82004bdf8 sp=0xc82004bd50
main.main()
	/home/docker/COLAS/src/abdprocess.go:87 +0x29e fp=0xc82004bf50 sp=0xc82004bdf8
runtime.main()
	/home/docker/go/src/runtime/proc.go:188 +0x2b0 fp=0xc82004bfa0 sp=0xc82004bf50
runtime.goexit()
	/home/docker/go/src/runtime/asm_amd64.s:1998 +0x1 fp=0xc82004bfa8 sp=0xc82004bfa0

goroutine 17 [syscall, locked to thread]:
runtime.goexit()
	/home/docker/go/src/runtime/asm_amd64.s:1998 +0x1

goroutine 20 [IO wait]:
net.runtime_pollWait(0x7f4138ba20e8, 0x72, 0x0)
	/home/docker/go/src/runtime/netpoll.go:160 +0x60
net.(*pollDesc).Wait(0xc820148b50, 0x72, 0x0, 0x0)
	/home/docker/go/src/net/fd_poll_runtime.go:73 +0x3a
net.(*pollDesc).WaitRead(0xc820148b50, 0x0, 0x0)
	/home/docker/go/src/net/fd_poll_runtime.go:78 +0x36
net.(*netFD).accept(0xc820148af0, 0x0, 0x7f4138ba21e0, 0xc820132e00)
	/home/docker/go/src/net/fd_unix.go:426 +0x27c
net.(*TCPListener).AcceptTCP(0xc820028228, 0x452c60, 0x0, 0x0)
	/home/docker/go/src/net/tcpsock_posix.go:254 +0x4d
net/http.tcpKeepAliveListener.Accept(0xc820028228, 0x0, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:2427 +0x41
net/http.(*Server).Serve(0xc8200f3980, 0x7f4138ba21a8, 0xc820028228, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:2117 +0x129
net/http.(*Server).ListenAndServe(0xc8200f3980, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:2098 +0x136
net/http.ListenAndServe(0x7dda70, 0x5, 0x7f4138ba1110, 0xc82000e050, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:2195 +0x98
_/home/docker/COLAS/src/daemons.HTTP_Server()
	/home/docker/COLAS/src/daemons/httpServer.go:62 +0x801
created by _/home/docker/COLAS/src/daemons.Writer_process
	/home/docker/COLAS/src/daemons/writer.go:100 +0x1e5

goroutine 3 [IO wait]:
net.runtime_pollWait(0x7f4138ba2028, 0x72, 0xc82016e000)
	/home/docker/go/src/runtime/netpoll.go:160 +0x60
net.(*pollDesc).Wait(0xc820148bc0, 0x72, 0x0, 0x0)
	/home/docker/go/src/net/fd_poll_runtime.go:73 +0x3a
net.(*pollDesc).WaitRead(0xc820148bc0, 0x0, 0x0)
	/home/docker/go/src/net/fd_poll_runtime.go:78 +0x36
net.(*netFD).Read(0xc820148b60, 0xc82016e000, 0x1000, 0x1000, 0x0, 0x7f4138b61000, 0xc820070000)
	/home/docker/go/src/net/fd_unix.go:250 +0x23a
net.(*conn).Read(0xc820028230, 0xc82016e000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/net.go:172 +0xe4
net/http.(*connReader).Read(0xc820132e20, 0xc82016e000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:526 +0x196
bufio.(*Reader).fill(0xc82013a660)
	/home/docker/go/src/bufio/bufio.go:97 +0x1e9
bufio.(*Reader).ReadSlice(0xc82013a660, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0)
	/home/docker/go/src/bufio/bufio.go:328 +0x21a
bufio.(*Reader).ReadLine(0xc82013a660, 0x0, 0x0, 0x0, 0x7c4200, 0x0, 0x0)
	/home/docker/go/src/bufio/bufio.go:357 +0x53
net/textproto.(*Reader).readLineSlice(0xc820129350, 0x0, 0x0, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/textproto/reader.go:55 +0x81
net/textproto.(*Reader).ReadLine(0xc820129350, 0x0, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/textproto/reader.go:36 +0x40
net/http.readRequest(0xc82013a660, 0x0, 0xc8201a0000, 0x0, 0x0)
	/home/docker/go/src/net/http/request.go:721 +0xb6
net/http.(*conn).readRequest(0xc8200f3a00, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:705 +0x359
net/http.(*conn).serve(0xc8200f3a00)
	/home/docker/go/src/net/http/server.go:1425 +0x947
created by net/http.(*Server).Serve
	/home/docker/go/src/net/http/server.go:2137 +0x44e

*/

/*


INFO	Starting reader

Running http server
INFO	Set Servers
 servers   0
INFO	SetName
reader_0
Expected 1 parameters, found 2
reader_0
INFO	StartProcess called
fatal error: unexpected signal during runtime execution
[signal 0xb code=0x1 addr=0x0 pc=0x7fe3531350cc]

runtime stack:
runtime.throw(0x854ba0, 0x2a)
	/home/docker/go/src/runtime/panic.go:547 +0x90
runtime.sigpanic()
	/home/docker/go/src/runtime/sigpanic_unix.go:12 +0x5a

goroutine 1 [syscall, locked to thread]:
runtime.cgocall(0x6b9230, 0xc82004baf8, 0x0)
	/home/docker/go/src/runtime/cgocall.go:123 +0x11b fp=0xc82004baa8 sp=0xc82004ba78
_/home/docker/COLAS/src/daemons._Cfunc_SODAW_read(0x7fe3480008c0, 0x7fe348000a20, 0x7fe300000000, 0x7fe348000a40, 0x7fe348000a60, 0x0)
	??:0 +0x42 fp=0xc82004baf8 sp=0xc82004baa8
_/home/docker/COLAS/src/daemons.reader_daemon()
	/home/docker/COLAS/src/daemons/reader.go:69 +0x495 fp=0xc82004bd58 sp=0xc82004baf8
_/home/docker/COLAS/src/daemons.Reader_process(0xc820072f00)
	/home/docker/COLAS/src/daemons/reader.go:101 +0x2b3 fp=0xc82004bdf8 sp=0xc82004bd58
main.main()
	/home/docker/COLAS/src/abdprocess.go:85 +0x278 fp=0xc82004bf50 sp=0xc82004bdf8
runtime.main()
	/home/docker/go/src/runtime/proc.go:188 +0x2b0 fp=0xc82004bfa0 sp=0xc82004bf50
runtime.goexit()
	/home/docker/go/src/runtime/asm_amd64.s:1998 +0x1 fp=0xc82004bfa8 sp=0xc82004bfa0

goroutine 17 [syscall, locked to thread]:
runtime.goexit()
	/home/docker/go/src/runtime/asm_amd64.s:1998 +0x1

goroutine 20 [IO wait]:
net.runtime_pollWait(0x7fe3534c40e8, 0x72, 0x0)
	/home/docker/go/src/runtime/netpoll.go:160 +0x60
net.(*pollDesc).Wait(0xc820164b50, 0x72, 0x0, 0x0)
	/home/docker/go/src/net/fd_poll_runtime.go:73 +0x3a
net.(*pollDesc).WaitRead(0xc820164b50, 0x0, 0x0)
	/home/docker/go/src/net/fd_poll_runtime.go:78 +0x36
net.(*netFD).accept(0xc820164af0, 0x0, 0x7fe3534c81e0, 0xc820194000)
	/home/docker/go/src/net/fd_unix.go:426 +0x27c
net.(*TCPListener).AcceptTCP(0xc820028218, 0x452c60, 0x0, 0x0)
	/home/docker/go/src/net/tcpsock_posix.go:254 +0x4d
net/http.tcpKeepAliveListener.Accept(0xc820028218, 0x0, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:2427 +0x41
net/http.(*Server).Serve(0xc8200ff980, 0x7fe3534c41a8, 0xc820028218, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:2117 +0x129
net/http.(*Server).ListenAndServe(0xc8200ff980, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:2098 +0x136
net/http.ListenAndServe(0x7dda70, 0x5, 0x7fe3534c3110, 0xc82000e050, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:2195 +0x98
_/home/docker/COLAS/src/daemons.HTTP_Server()
	/home/docker/COLAS/src/daemons/httpServer.go:62 +0x801
created by _/home/docker/COLAS/src/daemons.Reader_process
	/home/docker/COLAS/src/daemons/reader.go:96 +0x116

goroutine 50 [IO wait]:
net.runtime_pollWait(0x7fe3534c4028, 0x72, 0xc82019e000)
	/home/docker/go/src/runtime/netpoll.go:160 +0x60
net.(*pollDesc).Wait(0xc82018a060, 0x72, 0x0, 0x0)
	/home/docker/go/src/net/fd_poll_runtime.go:73 +0x3a
net.(*pollDesc).WaitRead(0xc82018a060, 0x0, 0x0)
	/home/docker/go/src/net/fd_poll_runtime.go:78 +0x36
net.(*netFD).Read(0xc82018a000, 0xc82019e000, 0x1000, 0x1000, 0x0, 0x7fe353483000, 0xc820070000)
	/home/docker/go/src/net/fd_unix.go:250 +0x23a
net.(*conn).Read(0xc82018e000, 0xc82019e000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/net.go:172 +0xe4
net/http.(*connReader).Read(0xc820194020, 0xc82019e000, 0x1000, 0x1000, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:526 +0x196
bufio.(*Reader).fill(0xc82019c000)
	/home/docker/go/src/bufio/bufio.go:97 +0x1e9
bufio.(*Reader).ReadSlice(0xc82019c000, 0xa, 0x0, 0x0, 0x0, 0x0, 0x0)
	/home/docker/go/src/bufio/bufio.go:328 +0x21a
bufio.(*Reader).ReadLine(0xc82019c000, 0x0, 0x0, 0x0, 0x7c4200, 0x0, 0x0)
	/home/docker/go/src/bufio/bufio.go:357 +0x53
net/textproto.(*Reader).readLineSlice(0xc82018c060, 0x0, 0x0, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/textproto/reader.go:55 +0x81
net/textproto.(*Reader).ReadLine(0xc82018c060, 0x0, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/textproto/reader.go:36 +0x40
net/http.readRequest(0xc82019c000, 0x0, 0xc82010d6c0, 0x0, 0x0)
	/home/docker/go/src/net/http/request.go:721 +0xb6
net/http.(*conn).readRequest(0xc820190000, 0x0, 0x0, 0x0)
	/home/docker/go/src/net/http/server.go:705 +0x359
net/http.(*conn).serve(0xc820190000)
	/home/docker/go/src/net/http/server.go:1425 +0x947
created by net/http.(*Server).Serve
	/home/docker/go/src/net/http/server.go:2137 +0x44e

*/