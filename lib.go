package ampgosetup2

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/bogem/id3v2"
	"github.com/disintegration/imaging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Tagmap exported
type Tagmap struct {
	Dirpath     string `bson:"dirpath"`
	Filename    string `bson:"filename"`
	Extension   string `bson:"extension"`
	FileID      string `bson:"fileID"`
	Filesize    string `bson:"filesize"`
	Artist      string `bson:"artist"`
	ArtistID    string `bson:"artistID"`
	Album       string `bson:"album"`
	AlbumID     string `bson:"albumID"`
	Title       string `bson:"title"`
	Genre       string `bson:"genre"`
	TitlePage   string `bson:"titlepage"`
	PicID       string `bson:"picID"`
	PicDB       string `bson:"picDB"`
	PicPath     string `bson:"picPath"`
	PicHttpAddr string `bson:"picHttpAddr"`
	Idx         string `bson:"idx"`
	HttpAddr    string `bson:"httpaddr"`
	Duration    string `bson:"duration"`

	// ArtStart string `bson:"artstart"`
	// AlbStart string `bson:"albstart"`
	// TitStart string `bson:"titstart"`
	// Howl     string `bson:"howl"`
}

type ArtVieW2 struct {
	Artist   string              `bson:"artist"`
	ArtistID string              `bson:"artistID"`
	Albums   []map[string]string `bson:"albums"`
	AlbCount string              `bson:"albcount"`
	Page     string              `bson:"page"`
	Index    string              `bson:"idx"`
}

// type ArtVieW3 struct {
// 	Artist   string   `bson:"artist"`
// 	ArtistID string   `bson:"artistID"`
// 	Albums   []string `bson:"albums"`
// 	AlbCount string   `bson:"albcount"`
// 	Page     string   `bson:"page"`
// 	Index    string   `bson:"idx"`
// }

type AlbVieW2 struct {
	Artist      string              `bson:"artist"`
	ArtistID    string              `bson:"artistID"`
	Album       string              `bson:"album"`
	AlbumID     string              `bson:"albumID"`
	Songs       []map[string]string `bson:"songs"`
	AlbumPage   string              `bson:"albumpage"`
	NumSongs    string              `bson:"numsongs"`
	PicPath     string              `bson:"picPath"`
	Idx         string              `bson:"idx"`
	PicHttpAddr string              `bson:"picHttpAddr"`
}

type Imageinfomap struct {
	Dirpath       string `bson:"dirpath"`
	Filename      string `bson:"filename"`
	Imagesize     string `bson:"imagesize"`
	ImageHttpAddr string `bson:"imageHttpAddr"`
	Index         string `bson:"index"`
	IType         string `bson:"itype"`
	Page          string `bson:"page"`
}

func StartLibLogging() string {
	logtxtfile := os.Getenv("AMPGO_LIB_LOG_PATH")
	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile(logtxtfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Println("Logging started")
	return "Logging started"
}

func Close(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func Connect(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

func InsertOne(client *mongo.Client, ctx context.Context, dataBase, col string, doc interface{}) (*mongo.InsertOneResult, error) {
	collection := client.Database(dataBase).Collection(col)
	result, err := collection.InsertOne(ctx, doc)
	return result, err
}

func Query(client *mongo.Client, ctx context.Context, dataBase, col string, query, field interface{}) (result *mongo.Cursor, err error) {
	collection := client.Database(dataBase).Collection(col)
	result, err = collection.Find(ctx, query, options.Find().SetProjection(field))
	return
}

func AmpgoDistinct(db string, coll string, fieldd string) []string {
	filter := bson.D{}
	opts := options.Distinct().SetMaxTime(2 * time.Second)
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "AmpgoDistinct: MongoDB connection has failed")
	collection := client.Database(db).Collection(coll)
	DD1, err2 := collection.Distinct(context.TODO(), fieldd, filter, opts)
	CheckError(err2, "AmpgoDistinct: MongoDB distinct album has failed")
	var DAlbum1 []string
	for _, DD := range DD1 {
		zoo := fmt.Sprintf("%s", DD)
		DAlbum1 = append(DAlbum1, zoo)
	}
	return DAlbum1
}

func AmpgoFindOne(db string, coll string, filtertype string, filterstring string) map[string]string {
	filter := bson.M{filtertype: filterstring}
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "AmpgoFindOne: MongoDB connection has failed")
	collection := client.Database(db).Collection(coll)
	var results map[string]string = make(map[string]string)
	err = collection.FindOne(context.Background(), filter).Decode(&results)
	if err != nil {
		log.Println("AmpgoFindOne: find one has fucked up")
		log.Fatal(err)
	}
	return results
}

func AmpgoFind(dbb string, collb string, filtertype string, filterstring string) []map[string]string {
	filter := bson.M{}
	if filtertype != "None" && filterstring != "None" {
		filter = bson.M{filtertype: filterstring}
	}
	// filter := bson.D{{filtertype, filterstring}}
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "AmpgoFind: MongoDB connection has failed")
	coll := client.Database(dbb).Collection(collb)
	cur, err := coll.Find(context.TODO(), filter)
	CheckError(err, "AmpgoFind: ArtPipeline find has failed")
	var results []map[string]string //all albums for artist to include double entries
	if err = cur.All(context.TODO(), &results); err != nil {
		log.Println("AmpgoFind: cur.All has fucked up")
		log.Fatal(err)
	}
	return results
}

func AmpgoInsertOne(db string, coll string, ablob map[string]string) {
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	CheckError(err, "AmpgoInsertOne: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, db, coll, ablob)
	CheckError(err2, "AmpgoInsertOne has failed")
}

//////////////////////////////////////////////////////////////////////////

func getFileInfo(apath string) (filename string, size string) {
	ltn, err := os.Open(apath)
	CheckError(err, "getFileInfo: file open has fucked up")
	defer ltn.Close()
	ltnInfo, _ := ltn.Stat()
	filename = ltnInfo.Name()
	size = strconv.FormatInt(ltnInfo.Size(), 10)
	return
}

func UUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	uuid[8] = 0x80
	uuid[4] = 0x40
	boo := hex.EncodeToString(uuid)
	return boo, nil
}

func resizeImage(infile string, outfile string) string {
	pic, err := imaging.Open(infile)
	if err != nil {
		return os.Getenv("AMPGO_NO_ART_PIC_PATH")
	}
	sjImage := imaging.Resize(pic, 200, 0, imaging.Lanczos)
	err = imaging.Save(sjImage, outfile)
	CheckError(err, "resizeImage: image save has fucked up")
	return outfile
}

type Fjpg struct {
	exists bool
	path   string
}

func folderjpg_check(apath string) Fjpg {
	fjpg := "folder.jpg"
	dir, _ := filepath.Split(apath)
	testfile := dir + fjpg
	_, error := os.Stat(testfile)
	if os.IsNotExist(error) {
		var pic Fjpg
		pic.exists = false
		pic.path = testfile
		return pic
	} else {
		var pic Fjpg
		pic.exists = true
		pic.path = testfile
		return pic
	}
}

func DumpArtToFile(apath string) (string, string, string, string, string) {
	tag, err := id3v2.Open(apath, id3v2.Options{Parse: true})
	if err != nil {
		log.Println(err)
		log.Println(apath)
		return "None", "None", "None", "None", "None"
	}
	defer tag.Close()
	artist := tag.Artist()
	album := tag.Album()
	title := tag.Title()
	genre := tag.Genre()
	folderjpgcheck := folderjpg_check(apath)
	if folderjpgcheck.exists {
		CreateFolderJpgImageInfoMap(folderjpgcheck.path)
		return artist, album, title, genre, folderjpgcheck.path
	} else {
		dumpOutFile2 := os.Getenv("AMPGO_THUMB_PATH") + tag.Artist() + "_-_" + tag.Album() + ".jpg"
		newdumpOutFile2 := strings.Replace(dumpOutFile2, " ", "_", -1)
		dumpOutFileThumb := os.Getenv("AMPGO_THUMB_PATH") + tag.Artist() + "_-_" + tag.Album() + "_thumb.jpg"
		newdumpOutFileThumb := strings.Replace(dumpOutFileThumb, " ", "_", -1)
		pictures := tag.GetFrames(tag.CommonID("Attached picture"))

		dir, _ := filepath.Split(apath)
		newfolderjpg_path := dir + "/folder.jpg"
		for _, f := range pictures {
			pic, ok := f.(id3v2.PictureFrame)
			if !ok {
				log.Fatal("DumpArtToFile: Couldn't assert picture frame")
				CreateFolderJpgImageInfoMap(os.Getenv("AMPGO_NO_ART_PIC_PATH"))
				return artist, album, title, genre, os.Getenv("AMPGO_NO_ART_PIC_PATH")
			}
			g, err := os.Create(newdumpOutFile2)
			CheckError(err, "DumpArtToFile: Unable to create newdumpOutFile2")
			h, err := os.Create(newfolderjpg_path)
			CheckError(err, "DumpArtToFile: Unable to create newdumpOutFile2")
			n3, err := g.Write(pic.Picture)
			CheckError(err, "DumpArtToFile: newdumpOutfile2 Write has fucked up")
			h3, err := h.Write(pic.Picture)
			g.Close()
			h.Close()
			fmt.Println(n3, "DumpArtToFile: bytes written successfully")
		}
		outfile22 := resizeImage(newdumpOutFile2, newdumpOutFileThumb)
		CreateFolderJpgImageInfoMap(outfile22)
		return artist, album, title, genre, outfile22
	}
}

func TaGmap(apath string, apage int, idx int) (TaGmaP Tagmap) {
	artist, album, title, genre, picpath := DumpArtToFile(apath)
	if artist != "None" && album != "None" && title != "None" {
		log.Println(apath)
		page := strconv.Itoa(apage)
		index := strconv.Itoa(idx)
		uuid, _ := UUID()
		pichttpaddr := os.Getenv("AMPGO_SERVER_ADDRESS") + ":" + os.Getenv("AMPGO_SERVER_PORT") + picpath[5:]
		fname, size := getFileInfo(apath)
		httpaddr := os.Getenv("AMPGO_SERVER_ADDRESS") + ":" + os.Getenv("AMPGO_SERVER_PORT") + apath[5:]
		TaGmaP.Dirpath = filepath.Dir(apath)
		TaGmaP.Filename = fname
		TaGmaP.Extension = filepath.Ext(apath)
		TaGmaP.FileID = uuid
		TaGmaP.Filesize = size
		TaGmaP.Artist = artist
		TaGmaP.ArtistID = "None"
		TaGmaP.Album = album
		TaGmaP.AlbumID = "None"
		TaGmaP.Title = title
		TaGmaP.Genre = genre
		TaGmaP.TitlePage = page
		TaGmaP.PicID = uuid
		TaGmaP.PicDB = "None"
		TaGmaP.PicPath = picpath
		TaGmaP.PicHttpAddr = pichttpaddr
		TaGmaP.Idx = index
		TaGmaP.HttpAddr = httpaddr
		TaGmaP.Duration = "None"
		// TaGmaP.ArtStart = "None"
		// TaGmaP.AlbStart = "None"
		// TaGmaP.TitStart = "None"
		// TaGmaP.Howl = ""
		client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
		CheckError(err, "TaGmap: Connections has failed")
		defer Close(client, ctx, cancel)
		_, err2 := InsertOne(client, ctx, "tempdb1", "meta1", &TaGmaP)
		CheckError(err2, "TaGmap: Tempdb1 insertion has failed")
		return
	} else {
		os.Remove(apath)
	}
	return
}

/////////////////////////////////////////////////////////////////////////////////////////////

func InsAlbumID(alb string) {
	uuid, _ := UUID()
	Albid := map[string]string{"album": alb, "albumID": uuid}
	AmpgoInsertOne("tempdb2", "albumid", Albid)
}

// func startLibLogging() string {
// 	var logtxtfile string = os.Getenv("AMPGO_LIB_LOG_PATH")
// 	file, err := os.OpenFile(logtxtfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.SetOutput(file)
// 	return "Logging started"
// }

func GetPicForAlbum(alb string) map[string]string {
	// startLibLogging()
	log.Printf("GetPicForAlbum: %s this is alb", alb)
	filter := bson.M{"album": alb}
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "GetPicForAlbum: MongoDB connection has failed")
	collection := client.Database("maindb").Collection("maindb")
	var albuminfo Tagmap
	err = collection.FindOne(context.Background(), filter).Decode(&albuminfo)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("GetPicForAlbum: %s this is album", alb)
	log.Printf("GetPicForAlbum: %s this is AlbumID", albuminfo.AlbumID)
	log.Printf("GetPicForAlbum: %s this is PicHttpAddr", albuminfo.PicHttpAddr)

	var albinfo map[string]string = make(map[string]string)
	albinfo["Album"] = alb
	albinfo["AlbumID"] = albuminfo.AlbumID
	albinfo["PicPath"] = albuminfo.PicHttpAddr
	AmpgoInsertOne("tempdb2", "artidpic", albinfo)
	fmt.Println(albinfo)
	log.Println(albinfo)
	return albinfo
}

func InsArtistID(art string) {
	uuid, _ := UUID()
	Artid := map[string]string{"artist": art, "artistID": uuid}
	AmpgoInsertOne("tempdb2", "artistid", Artid)
}

func GetTitleOffsetAll() (Main2SL []map[string]string) {
	filter := bson.D{}
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "GetTitleOffsetAll: MongoDB connection has failed")
	collection := client.Database("tempdb1").Collection("meta1")
	cur, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	if err = cur.All(context.Background(), &Main2SL); err != nil {
		log.Println("GetTitleOffsetAll: cur.All has failed")
		log.Fatal(err)
	}
	return
}

func gArtistInfo(Art string) map[string]string {
	filter := bson.M{"artist": Art}
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "gArtistInfo: MongoDB connection has failed")
	collection := client.Database("tempdb2").Collection("artistid")
	var ArtInfo map[string]string = make(map[string]string)
	err = collection.FindOne(context.Background(), filter).Decode(&ArtInfo)
	if err != nil {
		log.Println("gArtistInfo: has failed")
		log.Fatal(err)
	}
	return ArtInfo
}

func gAlbumInfo(Alb string) map[string]string {
	filter := bson.M{"album": Alb}
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "gAlbumInfo: MongoDB connection has failed")
	collection := client.Database("tempdb2").Collection("albumid")
	var AlbInfo map[string]string = make(map[string]string)
	err = collection.FindOne(context.Background(), filter).Decode(&AlbInfo)
	if err != nil {
		log.Println("gAlbumInfo: has failed")
		log.Fatal(err)
	}
	return AlbInfo
}

func gDurationInfo(filename string) map[string]string {
	log.Println(filename)
	filter := bson.M{"filename": filename}
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "MongoDB connection has failed")
	collection := client.Database("durdb").Collection("durdb")
	var durinfo map[string]string = make(map[string]string)
	err = collection.FindOne(context.Background(), filter).Decode(&durinfo)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(durinfo)
	return durinfo
}

// func startsWith(astring string) string {
// 	if len(astring) > 3 {
// 		if astring[3:] == "The" || astring[3:] == "the" {
// 			return strings.ToUpper(astring[4:5])
// 		} else {
// 			return strings.ToUpper(astring[:1])
// 		}
// 	} else {
// 		return strings.ToUpper(astring[:1])
// 	}
// }

func UpdateMainDB(m2 map[string]string) (Doko Tagmap) {
	log.Println(m2["filename"])
	artID := gArtistInfo(m2["artist"])
	log.Println(artID)
	albID := gAlbumInfo(m2["album"])
	log.Println(albID)
	fullpath := m2["dirpath"] + "/" + m2["filename"]
	log.Println(fullpath)
	duration := gDurationInfo(fullpath)
	log.Println(duration)
	Doko.Dirpath = m2["dirpath"]
	Doko.Filename = m2["filename"]
	Doko.Extension = m2["extension"]
	Doko.FileID = m2["fileID"]
	Doko.Filesize = m2["filesize"]
	Doko.Artist = m2["artist"]
	Doko.ArtistID = artID["artistID"]
	Doko.Album = m2["album"]
	Doko.AlbumID = albID["albumID"]
	Doko.Title = m2["title"]
	Doko.Genre = m2["genre"]
	Doko.PicID = m2["picID"]
	Doko.PicDB = "thumbnails"
	Doko.TitlePage = m2["titlepage"]
	Doko.Idx = m2["idx"]
	Doko.PicPath = m2["picPath"]
	Doko.PicHttpAddr = m2["picHttpAddr"]
	Doko.HttpAddr = m2["httpaddr"]
	Doko.Duration = duration["duration"]
	// Doko.ArtStart = startsWith(m2["artist"])
	// Doko.AlbStart = strings.ToUpper(m2["album"][:1])
	// Doko.TitStart = strings.ToUpper(m2["title"][:1])
	// Doko.Howl = m2["howl"]
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	CheckError(err, "UpdateMainDB: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, "maindb", "maindb", &Doko)
	CheckError(err2, "UpdateMainDB: maindb insertion has failed")
	return
}

func GDistArtist2() (dArtAll []map[string]string) {
	dArtist := AmpgoDistinct("maindb", "maindb", "artist")
	for _, art := range dArtist {
		dArt := AmpgoFindOne("maindb", "maindb", "artist", art)
		dArtAll = append(dArtAll, dArt)
	}
	return dArtAll
}

func Unique(arr []string) []string {
	occured := map[string]bool{}
	result := []string{}
	for e := range arr {
		if !occured[arr[e]] {
			occured[arr[e]] = true
			result = append(result, arr[e])
		}
	}
	return result
}

func create_just_albumID_list(alist []map[string]string) (just_albumID_list []string) {
	for _, albID := range alist {
		just_albumID_list = append(just_albumID_list, albID["albumID"])
	}
	return
}

func get_albums_for_artist(fullalblist []map[string]string) (final_alblist []map[string]string) {
	just_albumID_list := create_just_albumID_list(fullalblist)
	//remove double albumid entries
	unique_items := Unique(just_albumID_list)
	for _, uitem := range unique_items {
		albINFO := AmpgoFindOne("maindb", "maindb", "albumID", uitem)
		final_alblist = append(final_alblist, albINFO)
	}
	return
}

func ArtPipline(artmap map[string]string, page int, idx int) (MyArView ArtVieW2) {
	dirtyalblist := AmpgoFind("maindb", "maindb", "artistID", artmap["artistID"]) //[]map[string]string
	results2 := get_albums_for_artist(dirtyalblist)
	albc := len(results2)
	albcount := strconv.Itoa(albc)
	MyArView.Artist = artmap["artist"]
	MyArView.ArtistID = artmap["artistID"]
	MyArView.Albums = results2
	MyArView.AlbCount = albcount
	MyArView.Page = strconv.Itoa(page)
	MyArView.Index = strconv.Itoa(idx)
	return
}

func InsArtPipeline(AV1 ArtVieW2) {
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	CheckError(err, "InsArtPipeline: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, "artistview", "artistview", &AV1)
	CheckError(err2, "InsArtPipeline: artistview insertion has failed")
}

// func get_album_art_for_artist(fullalblist []map[string]string) (final_alblist []string) {
// 	just_albumID_list := create_just_albumID_list(fullalblist)
// 	//remove double albumid entries
// 	unique_items := unique(just_albumID_list)
// 	for _, uitem := range unique_items {
// 		albINFO := AmpgoFindOne("maindb", "maindb", "albumID", uitem)
// 		ai := albINFO["picHttpAddr"]
// 		final_alblist = append(final_alblist, ai)
// 	}
// 	return
// }

// func ArtPipline2(artmap map[string]string, page int, idx int) ArtVieW3 {
// 	dirtyalblist := AmpgoFind("maindb", "maindb", "artistID", artmap["artistID"]) //[]map[string]string
// 	results2 := get_album_art_for_artist(dirtyalblist)
// 	// albc := len(results2)
// 	var MyArView ArtVieW3
// 	MyArView.Artist = artmap["artist"]
// 	MyArView.ArtistID = artmap["artistID"]
// 	MyArView.Albums = results2
// 	MyArView.Page = strconv.Itoa(page)

// 	return MyArView
// }

// func InsArtPipeline2(AV1 ArtVieW3) {
// 	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
// 	CheckError(err, "InsArtPipeline: Connections has failed")
// 	defer Close(client, ctx, cancel)
// 	_, err2 := InsertOne(client, ctx, "artistview2", "artistview2", &AV1)
// 	CheckError(err2, "InsArtPipeline: artistview insertion has failed")
// }

func GDistAlbum() (DAlbAll []map[string]string) {
	DAlbumID := AmpgoDistinct("maindb", "maindb", "albumID")
	for _, albID := range DAlbumID {
		DAlb := AmpgoFindOne("maindb", "maindb", "albumID", albID)
		DAlbAll = append(DAlbAll, DAlb)
	}
	return
}

func get_songs_for_album(fullsonglist []map[string]string) (final_songlist []map[string]string) {
	//a list of just albumid's
	var just_songID_list []string
	for _, song := range fullsonglist {
		just_songID_list = append(just_songID_list, song["fileID"])
	}
	//remove double songID entries
	unique_items := Unique(just_songID_list)
	for _, uitem := range unique_items {
		songINFO := AmpgoFindOne("maindb", "maindb", "fileID", uitem)
		final_songlist = append(final_songlist, songINFO)
	}
	return final_songlist
}

// // // AlbPipeline exported
func AlbPipeline(DAlb map[string]string, page int, idx int) (MyAlbview AlbVieW2) {
	dirtysonglist := AmpgoFind("maindb", "maindb", "albumID", DAlb["albumID"])
	results := get_songs_for_album(dirtysonglist)
	songcount := len(results)
	MyAlbview.Artist = DAlb["artist"]
	MyAlbview.ArtistID = DAlb["artistID"]
	MyAlbview.Album = DAlb["album"]
	MyAlbview.AlbumID = DAlb["albumID"]
	MyAlbview.NumSongs = strconv.Itoa(songcount)
	MyAlbview.PicPath = DAlb["picPath"]
	MyAlbview.Songs = results
	MyAlbview.AlbumPage = strconv.Itoa(page)
	MyAlbview.Idx = strconv.Itoa(idx)
	MyAlbview.PicHttpAddr = DAlb["picHttpAddr"]
	return
}

// //InsAlbViewID exported
func InsAlbViewID(MyAlbview AlbVieW2) {
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgodb")
	CheckError(err, "InsAlbViewID: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, "albumview", "albumview", &MyAlbview)
	CheckError(err2, "InsAlbViewID: AmpgoInsertOne has failed")
}

/////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////

//RanPics exported
func CreateRandomPicsDB() []Imageinfomap {
	thumb_path := os.Getenv("AMPGO_THUMB_PATH")
	thumb_glob_path := thumb_path + "/*.jpg"
	thumb_glob, err := filepath.Glob(thumb_glob_path)
	CheckError(err, "CreateRandomPicsDB: CheckThumbDB has fucked up")
	var BulkImages []Imageinfomap
	var page int
	for i, v := range thumb_glob {
		if i < 5 {
			page = 1
		} else if i%5 == 0 {
			page++
		} else {
			page = page + 0
		}
		var iim Imageinfomap = create_image_info_map(i, v, page)
		BulkImages = append(BulkImages, iim)
	}
	return BulkImages
}

func create_image_info_map(i int, afile string, page int) Imageinfomap {
	itype := get_type(afile)
	dir, filename := filepath.Split(afile)
	image_size := get_image_size(afile)
	image_http_path := create_image_http_addr(afile)
	ii := i + 1
	var ImageInfoMap Imageinfomap
	ImageInfoMap.Dirpath = dir
	ImageInfoMap.Filename = filename
	ImageInfoMap.Imagesize = image_size
	ImageInfoMap.ImageHttpAddr = image_http_path
	ImageInfoMap.Index = strconv.Itoa(ii)
	ImageInfoMap.IType = itype
	ImageInfoMap.Page = strconv.Itoa(page)
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgo")
	CheckError(err, "create_image_info_map: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, "coverart", "coverart", ImageInfoMap)
	CheckError(err2, "create_image_info_map: coverart insertion has failed")
	return ImageInfoMap
}

func CreateFolderJpgImageInfoMap(afile string) {
	itype := get_type(afile)
	dir, filename := filepath.Split(afile)
	image_size := get_image_size(afile)
	image_http_path := create_image_http_addr(afile)
	var ImageInfoMap Imageinfomap
	ImageInfoMap.Dirpath = dir
	ImageInfoMap.Filename = filename
	ImageInfoMap.Imagesize = image_size
	ImageInfoMap.ImageHttpAddr = image_http_path
	ImageInfoMap.IType = itype
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgo")
	CheckError(err, "create_image_info_map: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, "foldercoverart", "foldercoverart", ImageInfoMap)
	CheckError(err2, "create_image_info_map: coverart insertion has failed")
}

func get_type(afile string) string {
	if strings.Contains(afile, "thumb") {
		return "thumb"
	} else {
		return "original"
	}
}

func get_image_size(apath string) string {
	fi, err := os.Stat(apath)
	CheckError(err, "get_image_size: os.stat has failed")
	size := fi.Size()
	newsize := int(size)
	return strconv.Itoa(newsize)
}

func create_image_http_addr(aimage string) string {
	return os.Getenv("AMPGO_SERVER_ADDRESS") + ":" + os.Getenv("AMPGO_SERVER_PORT") + aimage[5:]
}

type randDb struct {
	PlayListName  string              `bson:"playlistname"`
	PlayListID    string              `bson:"playlistID"`
	PlayListCount string              `bson:"playlistcount"`
	Playlist      []map[string]string `bson:"playlist"`
}

func CreateRandomPlaylistDB() string {
	var ranDBInfo randDb
	var emptylist []map[string]string
	var emptyitem map[string]string = map[string]string{"None": "No Songs Found"}
	emptylist = append(emptylist, emptyitem)
	uuid, _ := UUID()
	ranDBInfo.PlayListName = "EmptyRandomPlaylist"
	ranDBInfo.PlayListID = uuid
	ranDBInfo.PlayListCount = "0"
	ranDBInfo.Playlist = emptylist
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgo")
	CheckError(err, "CreateRandomPlaylistDB: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, "randplaylists", "randplaylists", ranDBInfo)
	CheckError(err2, "CreateRandomPlaylistDB: randplaylists insertion has failed")
	return "Created"
}

func ReadDurationFile(apath string) map[string]string {
	data, err := ioutil.ReadFile(apath)
	CheckError(err, "ReadDurationFile: mp3info read has failed")
	var mp3info map[string]string
	err2 := json.Unmarshal(data, &mp3info)
	CheckError(err2, "ReadDurationFile: json unmarshal has failed")
	return mp3info
}

func InsertDurationInfo(apath string) string {
	mp3 := ReadDurationFile(apath)
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgo")
	CheckError(err, "InsertDurationInfo: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, "durdb", "durdb", mp3)
	CheckError(err2, "InsertDurationInfo: durdb insertion has failed")
	return "durdb Created"
}

func CreateCurrentPlayListNameDB() string {
	var curPlayListName map[string]string = map[string]string{"record": "1", "curplaylistname": "None", "curplaylistID": "None"}
	client, ctx, cancel, err := Connect("mongodb://db:27017/ampgo")
	CheckError(err, "InsertDurationInfo: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, "curplaylistname", "curplaylistname", &curPlayListName)
	CheckError(err2, "InsertDurationInfo: curplaylistname insertion has failed")
	return "curplaylistname Created"
}
