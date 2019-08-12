package main

import (
	"fmt"
	//"encoding/json"
	"github.com/pion/webrtc/v2"
	"github.com/imroc/req"
)


var streamID = "dn3V99g1U0S8m6Z7pjnbxAHo3jpSBD4ne0AR"


func main() {

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		BundlePolicy: webrtc.BundlePolicyMaxBundle,
		SDPSemantics: webrtc.SDPSemanticsPlanB,
	}


	m := webrtc.MediaEngine{}
	m.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	m.RegisterCodec(webrtc.NewRTPH264Codec(webrtc.DefaultPayloadTypeH264, 90000))
	m.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
	
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))


	for i := 0; i < 10; i++ {

		peerConnection, err := api.NewPeerConnection(config)
		if err != nil {
			panic(err)
		}
	
	
		// add audio and video tranceiver 
		if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeAudio, webrtc.RtpTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
			panic(err)
		} else if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeVideo,webrtc.RtpTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
			panic(err)
		}

		peerConnection.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver){

			codec := track.Codec()
			fmt.Println("Track has started", i,  codec.Name)
		})
	
		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {

			//fmt.Printf("Connection State has changed %s \n", connectionState.String())
		})
	
	
	
		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			panic(err)
		}


		err = peerConnection.SetLocalDescription(offer)
		if err != nil {
			panic(err)
		}
	
	
		res, err := req.Post("http://127.0.0.1:6000/api/play", req.BodyJSON(map[string]string{
			"streamId": streamID,
			"sdp":      offer.SDP,
		}))
	
		if err != nil {
			panic(err)
		}
	
		var ret struct {
			Status int               `json:"s"`
			Data   struct {
				Sdp string  `json:"sdp"`
			} `json:"d"`
		}
		
		err = res.ToJSON(&ret)
		if err != nil {
			panic(err)
		}
	
		answerStr := ret.Data.Sdp
	
		answer := webrtc.SessionDescription{
			SDP: answerStr,
			Type: webrtc.SDPTypeAnswer,
		}
	
	
		err = peerConnection.SetRemoteDescription(answer)
	
		if err != nil {
			panic(err)
		}

	}


	select {}
}
