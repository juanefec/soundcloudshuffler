package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	scp "github.com/zackradisic/soundcloud-api"
)

func main() {
	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		log.Fatal(err.Error())
	}

	user, err := sc.GetUser(scp.GetUserOptions{
		ProfileURL: "https://soundcloud.com/kr3a71ve",
	})

	if err != nil {
		log.Fatal(err.Error())
	}

	ls, err := sc.GetLikes(scp.GetLikesOptions{
		ID:    user.ID,
		Type:  "track",
		Limit: 20,
	})

	if err != nil {
		log.Fatal(err.Error())
	}

	likes, err := ls.GetLikes()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("%v likes to play on shuffle:\n", len(likes))
	for _, like := range likes {
		fmt.Println(like.Track.Title)
	}

	next := rand.Intn(len(likes) - 1)
	for {
		log.Printf("Playing: %v\n", likes[next].Track.Title)

		play(sc, likes[next].Track.Media.Transcodings[0])
		next = rand.Intn(len(likes) - 1)
	}

}

func play(sc *scp.API, t scp.Transcoding) error {
	buffer := &bytes.Buffer{}

	err := sc.DownloadTrack(t, buffer)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}

	streamer, format, err := mp3.Decode(ioutil.NopCloser(buffer))
	if err != nil {
		log.Fatal(err)
		return err
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
	return nil
}
