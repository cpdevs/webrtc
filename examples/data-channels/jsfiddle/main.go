//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"syscall/js"

	"github.com/cpdevs/webrtc/v3"
)

func main() {
	js.Global().Set("wasmInitPeerConnection", wasmInitPeerConnection())

	// Stay alive
	select {}
}

func initPeerConnection(url string) error {
	// Configure and create a new PeerConnection.
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:devturnaz.cloudpurge.io:443"},
			},
			{
				URLs:       []string{"turn:devturnaz.cloudpurge.io:443?transport=tcp"},
				Username:   "cagefox",
				Credential: "qZ5LwNU9ueiBbbkK",
			},
		},
	}

	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return err
	}

	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		// Register channel opening handling
		d.OnOpen(func() {
			fmt.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", d.Label(), d.ID())

			fmt.Println("Will send message through data channel")
			sendErr := d.SendText("Data channel opened in wasm")
			if sendErr != nil {
				fmt.Println("Error sending msg through data channel")
			}
		})

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
		})
	})

	// Add handlers for setting up the connection.
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Println("Connection state changed to ", fmt.Sprint(state))
	})

	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			fmt.Println("Candidate received: ", candidate)
		}
	})

	offerSDP, err := sendOffer(url, pc)
	if err != nil {
		return err
	}

	setRemoteDescription(offerSDP, pc)

	err = sendAnswer(url, pc)
	if err != nil {
		fmt.Println("Error while sending answer")
		return err
	}

	return nil
}

func sendOffer(url string, pc *webrtc.PeerConnection) (sdp string, err error) {
	requestBody, err := json.Marshal(map[string]string{
		"video": "true",
		"audio": "true",
		"type":  "offer",
	})
	if err != nil {
		return "", err
	}

	offerURL := url + "offer"
	resp, err := http.Post(offerURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)
	sdp = strings.ReplaceAll(string(responseBody), "\"", "")

	return
}

func setRemoteDescription(sdp string, pc *webrtc.PeerConnection) error {
	remoteSDP := webrtc.SessionDescription{}
	b, err := base64.StdEncoding.DecodeString(sdp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &remoteSDP)
	if err != nil {
		return err
	}

	err = pc.SetRemoteDescription(remoteSDP)
	if err != nil {
		return err
	}
	return nil
}

func sendAnswer(url string, pc *webrtc.PeerConnection) error {
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return err
	}

	err = pc.SetLocalDescription(answer)
	if err != nil {
		return err
	}

	a, err := json.Marshal(*pc.LocalDescription())
	if err != nil {
		return err
	}

	answerSDP := base64.StdEncoding.EncodeToString(a)
	answerBody, err := json.Marshal(map[string]string{
		"sdp": answerSDP,
	})
	answerURL := url + "answer"
	_, err = http.Post(answerURL, "application/json", bytes.NewBuffer(answerBody))
	if err != nil {
		return err
	}
	return nil
}

func wasmInitPeerConnection() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) (resp interface{}) {
		if len(args) != 1 {
			return "Invalid no of arguments"
		}
		url := args[0].String()
		go func() {
			err := initPeerConnection(url)
			if err != nil {
				fmt.Println("WASM: Unable to init peer connection ", err)
			}
		}()
		return resp
	})
	return jsonFunc
}
