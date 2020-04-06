package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/imroc/req"
	"github.com/pion/webrtc/v2"
)

func benchPlay(streamURL string, apiURL string, count int) {


	config := webrtc.Configuration{
		ICEServers:   []webrtc.ICEServer{},
		BundlePolicy: webrtc.BundlePolicyMaxBundle,
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
	}

	m := webrtc.MediaEngine{}
	m.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	m.RegisterCodec(webrtc.NewRTPH264Codec(webrtc.DefaultPayloadTypeH264, 90000))

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	for i := 0; i < count; i++ {

		peerConnection, err := api.NewPeerConnection(config)
		if err != nil {
			panic(err)
		}

		if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RtpTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
			panic(err)
		} else if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RtpTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
			panic(err)
		}

		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			fmt.Printf("Connection State has changed %s  connection id %d \n", connectionState.String(), i)
		})

		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			panic(err)
		}

		err = peerConnection.SetLocalDescription(offer)
		if err != nil {
			panic(err)
		}

		res, err := req.Post(apiURL, req.BodyJSON(map[string]string{
			"streamurl": streamURL,
			"sdp":      offer.SDP,
		}))

		if err != nil {
			panic(err)
		}

		var ret struct {
			Code int
			Sdp  string
		}

		err = res.ToJSON(&ret)
		if err != nil {
			panic(err)
		}

		answerStr := ret.Sdp

		answer := webrtc.SessionDescription{
			SDP:  answerStr,
			Type: webrtc.SDPTypeAnswer,
		}

		err = peerConnection.SetRemoteDescription(answer)

		if err != nil {
			panic(err)
		}

	}

	select {}

}

func main() {

	var args struct {
		Count  int    `arg:"required" "-c" help:"stream count to play"`
		Stream string `arg:"required" "-s" help:"stream url"`
		Url    string `arg:"required" "-u" help:"http url to play"`
	}

	arg.MustParse(&args)

	benchPlay(args.Stream, args.Url,args.Count)
}
