package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/donuts-are-good/colors"
)

func makeID() string {
	id, _ := machineid.ProtectedID("What's it profit a man to gain the world and lose his soul?")
	return id
}

func delay(seconds time.Duration) {
	time.Sleep(seconds * time.Second)
}

func externalIP() {
	Log(DEBUG, "Fetching external IP")
	resp, rerr := http.Get("http://icanhazip.com")
	if rerr != nil {
		Log(ERROR, "http.Get error: "+rerr.Error())
		return
	}
	defer resp.Body.Close()

	body, berr := io.ReadAll(resp.Body)
	if berr != nil {
		Log(ERROR, "io.ReadAll error: "+berr.Error())
		return
	}

	bodyStr := string(body)
	trimBody := strings.TrimSuffix(bodyStr, "\n")

	Log(DEBUG, "External IP Query Response: "+bodyStr)

	parsedIP := net.ParseIP(trimBody)
	Log(DEBUG, parsedIP.String())
	varint := ipToVarint(parsedIP)
	extAddr = varint

	portAsInt, errPortAsInt := strconv.Atoi(p2pPort)
	Log(DEBUG, strconv.Itoa(portAsInt))
	if errPortAsInt != nil {
		Log(ERROR, "There was an error converting the port type: "+errPortAsInt.Error())
	}
	Log(INFO, colors.BrightGreen+"External: "+colors.NC+trimBody)
	extAddrStr = trimBody
}

func displayNodeStatus() {

	numBootPeers := countPeerType("boot")
	numGreenPeers := countPeerType("green")
	numWhitePeers := countPeerType("white")
	numBlackPeers := countPeerType("black")

	fmt.Println("")

	fmt.Println(colors.BrightBlack + "🡺 " + colors.BrightPurple + "Version:\t" + colors.NC + "v" + semverInfo())

	fmt.Println(colors.BrightBlack + "🡺 " + colors.BrightPurple + "Peer ID:\t" + colors.NC + makeID())

	fmt.Println(colors.BrightBlack + "🡺 " + colors.BrightPurple + "Peer Addr:\t" + colors.NC + extAddrStr + ":" + p2pPort)

	fmt.Println(colors.BrightBlack + "🡺 " + colors.BrightMagenta + "Boot Peers:\t" + colors.NC + strconv.Itoa(numBootPeers))

	fmt.Println(colors.BrightBlack + "🡺 " + colors.BrightMagenta + "White Peers:\t" + colors.NC + strconv.Itoa(numWhitePeers))

	fmt.Println(colors.BrightBlack + "🡺 " + colors.BrightMagenta + "Green Peers:\t" + colors.NC + strconv.Itoa(numGreenPeers))

	fmt.Println(colors.BrightBlack + "🡺 " + colors.BrightMagenta + "Black Peers:\t" + colors.NC + strconv.Itoa(numBlackPeers))

	fmt.Println(colors.BrightBlack + "🡺 " + colors.BrightPurple + "DB Dir:\t" + colors.NC + databaseFilename)

	fmt.Println("")

}

func smallNumToVarint(msgType string) []byte {

	p2puint, _ := strconv.Atoi(msgType)

	p2puint16 := uint16(p2puint)

	numberToConvert := uint16ToBig(p2puint16)

	encodedNumber := encode(numberToConvert)

	return encodedNumber

}

func initID() {
	// Generate a sticky identifier
	identity := makeID()
	Log(DEBUG, "Generated Peer ID: "+identity[:16]+"...")
	Log(INFO, "Peer ID:  "+identity)
}

func roll(min, max int) int {
	if min < 0 {
		Log(DEBUG, "Min was less than 0")
		return 0
	}
	if max < 1 {
		Log(DEBUG, "Max was less than 1")
		return 1
	}
	return rand.Intn(max-min+1) + min
}

func serializePeerToByte(peer PeerObj) []byte {
	peerStatus := []byte(peer.PeerStatus)
	return bytes.Join([][]byte{peerStatus, peer.PeerIPAddrVarint, peer.PeerPortVarint, peer.PeerIDLenBytes, peer.PeerIDBytes}, []byte(""))
}
