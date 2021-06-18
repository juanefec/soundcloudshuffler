package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"strconv"
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

	likes := getAllLikes(sc, user, 0)

	log.Printf("%v likes to play on shuffle:\n", len(likes))
	for _, like := range likes {
		fmt.Println(like.Track.Title)
	}

	for {
		next := rand.Intn(len(likes) - 1)

		log.Printf("Playing: %v\n", likes[next].Track.Title)

		play(sc, likes[next].Track.Media.Transcodings[0])
	}
}

func getAllLikes(sc *scp.API, user scp.User, offset int) []scp.Like {
	ls, err := sc.GetLikes(scp.GetLikesOptions{
		ID:     user.ID,
		Type:   "track",
		Limit:  200,
		Offset: offset,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	l, err := ls.GetLikes()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println(ls.NextHref)
	if ls.NextHref != "" {
		url, err := url.Parse(ls.NextHref)
		if err != nil {
			log.Fatal(err.Error())
		}
		off, err := strconv.Atoi(url.Query()["offset"][0])
		if err != nil {
			log.Fatal(err.Error())
		}
		l = append(l, getAllLikes(sc, user, off)...)
	}
	return l
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
