///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
// LICENSE: GNU General Public License, version 2 (GPLv2)
// Copyright 2016, Charlie J. Smotherman
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License v2
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////

package ampgosetup2

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
	// "context"
	"path/filepath"
	"strconv"
)

var OFFSET string = os.Getenv("AMPGO_OFFSET")
var OffSet int = convertSTR(OFFSET)

func convertSTR(astring string) int {
	Ofset, err := strconv.Atoi(astring)
	CheckError(err, "strconv has failed")
	return Ofset
}

//CheckError exported
func CheckError(err error, msg string) {
	if err != nil {
		fmt.Println(msg)
		log.Println(msg)
		log.Println(err)
		panic(err)
	}
}

var titlepage int = 0
var ii int = 0

func visit(pAth string, f os.FileInfo, err error) error {
	log.Println(pAth)

	ext := path.Ext(pAth)
	if ext == ".mp3info" {
		InsertDurationInfo(pAth)
	} else if ext == ".mp3" {
		if ii < OffSet {
			ii++
			titlepage = 1
		} else if ii%OffSet == 0 {
			ii++
			titlepage++
		} else {
			ii++
			titlepage = titlepage + 0
		}
		TaGmap(pAth, titlepage, ii)
	} else {
		fmt.Println("WTF are you? You must be a Dir")
		fmt.Println(pAth)
	}
	log.Println(pAth)
	return nil
}

func StartSetupLogging() string {
	logtxtfile := os.Getenv("AMPGO_SETUP_LOG_PATH")
	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile(logtxtfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Println("Logging started")
	return "server logging started"
}

func SetUpCheck() {
	// StartLibLogging()
	// StartSetupLogging()
	Setup()

	// fileinfo, err := os.Stat("setup.txt")
	// if os.IsNotExist(err) {
	// 	Setup()
	// }
	// log.Println(fileinfo)
}

//SetUp is exported to main
func Setup() {
	ti := time.Now()
	fmt.Println(ti)
	log.Println(ti)
	runtime.GOMAXPROCS(runtime.NumCPU())

	// log.Println("starting duration walk")
	// filepath.Walk(os.Getenv("AMPGO_MEDIA_PATH"), durationVisit)
	// log.Println("duration walk is complete")
	// log.Println("starting imgInfoDuration walk")
	// filepath.Walk(os.Getenv("AMPGO_MEDIA_PATH"), imgInfoVisit)

	log.Println("starting walk")
	filepath.Walk(os.Getenv("AMPGO_MEDIA_PATH"), visit)
	log.Println("walk is complete")

	log.Println("starting GetDistAlbumMeta1")
	dalb := AmpgoDistinct("tempdb1", "meta1", "album")
	fmt.Println(dalb)
	log.Println(dalb)
	log.Println("GetDistAlbumMeta1 is complete ")

	log.Println("starting InsAlbumID")
	var wg1 sync.WaitGroup
	for _, alb := range dalb {
		wg1.Add(1)
		go func(alb string) {
			InsAlbumID(alb)
			wg1.Done()
		}(alb)
		wg1.Wait()
	}
	log.Println("InsAlbumID is complete ")

	log.Println("starting GDistArtist")
	dart := AmpgoDistinct("tempdb1", "meta1", "artist")
	log.Println("GDistArtist is complete ")

	log.Println("starting InsArtistID")
	var wg2 sync.WaitGroup
	for _, art := range dart {
		wg2.Add(1)
		go func(art string) {
			InsArtistID(art)
			wg2.Done()
		}(art)
		wg2.Wait()
	}
	log.Println("InsArtistID is complete ")

	log.Println("starting GetTitleOffSetAll")
	AllObj := GetTitleOffsetAll()
	log.Println("GetTitleOffSetAll is complete ")

	log.Println("starting UpdateMainDB")
	var wg3 sync.WaitGroup
	for _, blob := range AllObj {
		log.Println(blob)
		wg3.Add(1)
		go func(blob map[string]string) {
			UpdateMainDB(blob)
			wg3.Done()
		}(blob)
		wg3.Wait()
	}
	log.Println("UpdateMainDB is complete ")

	// log.Println("starting GetPicForAlbum ")
	// var wg133 sync.WaitGroup
	// for _, alb := range dalb {
	// 	wg133.Add(1)
	// 	go func(alb string) {
	// 		zoo := GetPicForAlbum(alb)
	// 		fmt.Println(zoo)
	// 		wg133.Done()
	// 	}(alb)
	// 	wg133.Wait()
	// }
	// log.Println("GetPicForAlbum is complete")

	// //AggArtist
	// log.Println("starting UpdateMainDB")
	// DistArtist := GDistArtist2()
	// log.Println("GDistArtist2 is complete ")

	// log.Println("starting GArtInfo2")
	// var wg5 sync.WaitGroup
	// // var wg15 sync.WaitGroup
	// var artpage int = 0
	// for artIdx, DArtt := range DistArtist {
	// 	if artIdx < OffSet {
	// 		artpage = 1
	// 	} else if artIdx%OffSet == 0 {
	// 		artpage++
	// 	} else {
	// 		artpage = artpage + 0
	// 	}

	// 	APL := ArtPipline(DArtt, artpage, artIdx)

	// 	wg5.Add(1)
	// 	go func(APL ArtVieW2) {
	// 		InsArtPipeline(APL)
	// 		wg5.Done()
	// 	}(APL)
	// 	wg5.Wait()

	// }
	// fmt.Println("AggArtists is complete")
	// log.Println("AggArtists is complete")
	// // ArtistOffSet()w11
	// // fmt.Println("ArtistOffSet is complete")

	// //AggAlbum
	// fmt.Println("AggAlbum has started")

	// log.Println("Starting GDistAlbum3")
	// DistAlbum := GDistAlbum()

	// var wg6 sync.WaitGroup
	// var albpage int = 0
	// for albIdx, DAlb := range DistAlbum {
	// 	wg6.Add(1)
	// 	if albIdx < OffSet {
	// 		albpage = 1
	// 	} else if albIdx%OffSet == 0 {
	// 		albpage++
	// 	} else {
	// 		albpage = albpage + 0
	// 	}
	// 	APLX := AlbPipeline(DAlb, albpage, albIdx)
	// 	go func(APLX AlbVieW2) {
	// 		InsAlbViewID(APLX)
	// 		wg6.Done()
	// 	}(APLX)
	// 	wg6.Wait()
	// }
	// CreateRandomPicsDB()

	// CreateRandomPlaylistDB()

	// CreateCurrentPlayListNameDB()

	// var lines = []string{
	// 	"Go",
	// 	"is",
	// 	"the",
	// 	"best",
	// 	"programming",
	// 	"language",
	// 	"in",
	// 	"the",
	// 	"world",
	// }

	// f, err := os.Create("setup.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // remember to close the file
	// defer f.Close()

	// for _, line := range lines {
	// 	_, err := f.WriteString(line + "\n")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// fmt.Println("AlbumOffSet is complete")
	t2 := time.Now().Sub(ti)
	fmt.Println(t2)
	fmt.Println("THE END")

}
