package main

import (
	"fmt"
	"math/rand"
	"github.com/pion/webrtc/v2"
	"github.com/imroc/req"
	gstreamer "github.com/notedit/gstreamer-go"
)



func benchPlay() {
	
	
	var streamID = "v7Vj09vH3DIaeS84S5ctDmJGgRsVDVOyzEyN"

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

			fmt.Printf("Connection State has changed %s \n", connectionState.String())
		})
	
	
	
		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			panic(err)
		}


		err = peerConnection.SetLocalDescription(offer)
		if err != nil {
			panic(err)
		}
	
	
		res, err := req.Post("http://127.0.0.1:6001/api/play", req.BodyJSON(map[string]string{
			"streamId": streamID,
			"sdp":      offer.SDP,
		}))
	
		if err != nil {
			panic(err)
		}
	
		var ret struct {
			Status int      `json:"s"`
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


func benchPublish() {

	var rtmpSource = "rtmp://alihb2clive.wangxiao.eaydu.com/live_ali/x_3_test_ali"


	pipelineStr := "gst-launch-1.0 -v rtmpsrc location=%s ! flvdemux name=demux  demux.video ! queue ! h264parse ! rtph264pay timestamp-offset=0 config-interval=-1 ! appsink name=sink"

	pipelineStr = fmt.Sprintf(pipelineStr, rtmpSource)

	pipeline, err := gstreamer.New(pipelineStr)
	if err != nil {
		panic(err)
	}
	
	appsink := pipeline.FindElement("sink")
	
	pipeline.Start()
	
	out := appsink.Poll()



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

	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}
	
	// Create a video track
	videoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", "video")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}


	peerConnection.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver){

	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {

		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})


	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}


	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}


	res, err := req.Post("http://127.0.0.1:6001/api/publish", req.BodyJSON(map[string]string{
		"streamId": "streamdddddddd",
		"sdp":      offer.SDP,
	}))

	if err != nil {
		panic(err)
	}

	var ret struct {
		Status int      `json:"s"`
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


	for {
		buffer := <-out

		videoTrack.Write(buffer)
		fmt.Println("push ", len(buffer))
	}
	
}


func main() {

	benchPublish()
}
