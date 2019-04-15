package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"io"
	"net"
	"net/http"
	"os"
	"log"
	"time"
	//"strconv"
	//"strings"
)

const (
	sshPortEnv  = "SSH_PORT"
	httpPortEnv = "PORT"

	defaultSshPort  = "2022"
	defaultHttpPort = "3000"
)

var (
	currentGame = 0 //currently user number
	holdPoint = false
)
func ConnHandler(conn net.Conn, gm *GameManager){

		/////migration setting
		recvBuf := make([]byte, 4096)
	   for {
	      n, err := conn.Read(recvBuf)
	      if nil != err {
	         if io.EOF == err {
	            log.Println(err);
	            return
	         }
	         log.Println(err);
	         return
	      }
	      if 0 < n {
	         data := recvBuf[:n] // migration msg here
	         log.Println(string(data))
					 if(string(data)=="migrate"){
						 holdPoint = true
						 conn.Close()
						 fmt.Println("msg : ",string(data))
					 }
					 if(string(data)=="finish"){
						 holdPoint = false
						 conn.Close()
						 fmt.Println("msg : ",string(data))
					 }

	         if err != nil {
	            log.Println(err)
	            return
	         }
	      }
	   }



}
func handler(conn net.Conn, gm *GameManager, config *ssh.ServerConfig) {
	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		fmt.Println("Failed to handshake with new client")
		return
	}
	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			fmt.Println("could not accept channel.")
			return
		}

		// TODO: Remove this -- only temporary while we launch on HN
		//
		// To see how many concurrent users are online
		fmt.Printf("Player joined. Current stats: %d users, %d games\n",
			gm.SessionCount(), gm.GameCount())

		currentGame++

		// Reject all out of band requests accept for the unix defaults, pty-req and
		// shell.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				switch req.Type {
				case "pty-req":
					req.Reply(true, nil)
					continue
				case "shell":
					req.Reply(true, nil)
					continue
				}
				req.Reply(false, nil)
			}
		}(requests)
		//checkpoint :=0
		fmt.Printf(" sshConn.User : %s\n",sshConn.User())
		////user is jcdad and kyj
		/*if(gm.Games["a"]!=nil){
			for session := range gm.Games["a"].hub.Sessions{
				player := session.Player
				if(player.Name == sshConn.User()){
					gm.HandleExistChannel(channel,sshConn.User())
					checkpoint =1
				}
			}
		}
		if(checkpoint ==0){*/
		gm.HandleNewChannel(channel, sshConn.User())
		//}
	}
}

func port(env, def string) string {
	port := os.Getenv(env)
	if port == "" {
		port = def
	}

	return fmt.Sprintf(":%s", port)
}

/*func put_player_struct(player * Player) string {

}

func get_player_struct(buf string) * Player {

}*/

func main() {

//fmt.Println("In main.go main function")
	sshPort := port(sshPortEnv, defaultSshPort)
	httpPort := port(httpPortEnv, defaultHttpPort)

	// Everyone can login!
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		panic("Failed to load private key")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
	}

	config.AddHostKey(private)

	// Create the GameManager
	gm := NewGameManager()

	go func(){
		l, err := net.Listen("tcp", "0.0.0.0:8000")
		if nil != err {
		 	log.Println(err);
		}
		defer l.Close()

		for {
		 conn, err := l.Accept()
		 if nil != err {
					log.Println(err);
					continue
		 		}
		 	defer conn.Close()
			//fmt.Println("data arrive")
		 	go ConnHandler(conn,gm)
		}
	}()

	go func(){
		for i:=0; i<10000;i++{
			fmt.Println(i)
			time.Sleep(1*time.Second)
		}
	}()
	/*go func(){

		for{

			if(gm.Games["a"]!=nil && playerchecker){
				//for session := range gm.Games["a"].hub.Sessions{
					//player := session.Player
					fmt.Printf("Player : %s\n",tmpPlayer[0].Name)
					time.Sleep(1*time.Second)
				//}
			}
		}
	}()*/

	/*///////////////////send message
	go func() {
		conn,err:=net.Dial("tcp","143.248.57.99:8000")
		if nil!=err{
			log.Println(err)
		}
		for{
			if(gm.Games["a"]!=nil){
				for session := range gm.Games["a"].hub.Sessions{
					player := session.Player
					//fmt.Printf("Player Direction : %s \n",player.Direction)
					//input sequence ID=color, Direction, MArker, pos,score
					var tmpSend []string
					//var tmpTotalUser string

					//fmt.Printf("%s\n",player.Trail)

					tmpSend = append(tmpSend, player.Name,
						strconv.Itoa(int(player.Color)),
						strconv.Itoa(int(player.Direction)),
						strconv.Itoa(int(player.Marker)),
						strconv.Itoa(int(player.Pos.X)),
						strconv.Itoa(int(player.Pos.Y)),
						strconv.Itoa(int(player.score)))

					//fmt.Println("trail : %s", player.Trail[0].Marker)
					for i := range player.Trail {
						seg:=player.Trail[i]
						//fmt.Println("num : %d, trail : %s",i, seg)
						tmpSend = append(tmpSend,
						strconv.Itoa(int(seg.Marker)),
						strconv.Itoa(int(seg.Pos.X)),
						strconv.Itoa(int(seg.Pos.Y)),
						strconv.Itoa(int(seg.Color)))
					}
					var result string = strings.Join(tmpSend[:], ",")

					conn.Write([]byte(result))
				}
			}
			time.Sleep(time.Second/60)
		}
	}()

	///////////////////////*/
	fmt.Printf(
		"Listening on port %s for SSH and port %s for HTTP...\n",
		sshPort,
		httpPort,
	)

	go func() {
		panic(http.ListenAndServe(httpPort, http.FileServer(http.Dir("./static/"))))
	}()

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0%s", sshPort))
	if err != nil {
		panic("failed to listen for connection")
	}

	for {
		nConn, err := listener.Accept()
		if err != nil {
			panic("failed to accept incoming connection")
		}

		go handler(nConn, gm, config)
	}
}
