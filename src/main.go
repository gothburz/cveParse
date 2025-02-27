package main

import (
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"io/ioutil"
	"log"
	"logging"
	"net/http"
	"os"
	"path/filepath"
	"time"
	//"go.mongodb.org/mongo-driver/mongo"
	//"go.mongodb.org/mongo-driver/mongo/options"
)

/*
   =========================
   ==   CVE FEED STRUCT   ==
   =========================
*/

type cveFeed struct {
	name string
	dir  string
	meta string
	gz   string
	zip  string
}

/*
   =========================
   ==   CVE JSON STRUCT   ==
   =========================
*/

type cveJSON struct {
	CVEDataType         string     `json:"CVE_data_type"`
	CVEDataFormat       string     `json:"CVE_data_format"`
	CVEDataVersion      string     `json:"CVE_data_version"`
	CVEDataNumberOfCVEs string     `json:"CVE_data_numberOfCVEs"`
	CVEDataTimestamp    string     `json:"CVE_data_timestamp"`
	CVEItems            []CVEItems `json:"CVE_Items"`
}

type CVEItems struct {
	CVE              CVEItem        `json:"cve"`
	Configurations   Configurations `json:"configurations"`
	Impact           Impact         `json:"impact"`
	PublishedDate    string         `json:"publishedDate"`
	LastModifiedDate string         `json:"lastModifiedDate"`
}

type CVEItem struct {
	DataType    string      `json:"data_type"`
	DataFormat  string      `json:"data_format"`
	DataVersion string      `json:"data_version"`
	CVEDataMeta CVEDataMeta `json:"CVE_data_meta"`
	ProblemType ProblemType `json:"problemtype"`
	References  References  `json:"references"`
	Description Description `json:"description"`
}

type CVEDataMeta struct {
	ID       string `json:"ID"`
	ASSIGNER string `json:"ASSIGNER"`
}

type ProblemType struct {
	ProblemTypeData []ProblemTypeData `json:"problemtype_data"`
}

type ProblemTypeData struct {
	Desc []ProblemTypeDescription `json:"description"`
}

type ProblemTypeDescription struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type References struct {
	ReferenceData []ReferenceData `json:"reference_data"`
}

type ReferenceData struct {
	URL       string `json:"url"`
	Name      string `json:"name"`
	Refsource string `json:"refsource"`
	Tags      []string
}

type Description struct {
	DescriptionData []DescriptionData `json:"description_data"`
}

type DescriptionData struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type Configurations struct {
	CVEDataVersion string  `json:"CVE_data_version"`
	Nodes          []Nodes `json:"nodes"`
}

type Nodes struct {
	Operator string     `json:"operator"`
	CPEMatch []CPEMatch `json:"cpe_match"`
}

type CPEMatch struct {
	Vulnerable bool   `json:"vulnerable"`
	CPE23URI   string `json:"cpe23Uri"`
}

type Impact struct {
	BaseMetricV3 BaseMetricV3 `json:"baseMetricV3"`
	BaseMetricV2 BaseMetricV2 `json:"baseMetricV2"`
}

type BaseMetricV3 struct {
	CVSSV3              CVSSV3  `json:"cvssV3"`
	ExploitabilityScore float64 `json:"exploitabilityScore"`
	ImpactScore         float64 `json:"impactScore"`
}

type CVSSV3 struct {
	Version               string  `json:"version"`
	VectorString          string  `json:"vectorString"`
	AttackVector          string  `json:"attackVector"`
	AttackComplexity      string  `json:"attackComplexity"`
	PrivilegesRequired    string  `json:"privilegesRequired"`
	UserInteraction       string  `json:"userInteraction"`
	Scope                 string  `json:"scope"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	AvailabilityImpact    string  `json:"availabilityImpact"'`
	BaseScore             float64 `json:"baseScore"`
	BaseSeverity          string  `json:"baseSeverity"`
}

type BaseMetricV2 struct {
	CVSSV2                  CVSSV2  `json:"cvssV2"`
	Severity                string  `json:"severity"`
	ExploitabilityScore     float64 `json:"exploitabilityScore"`
	ImpactScore             float64 `json:"impactScore"`
	AcInsufInfo             bool    `json:"acInsufInfo"`
	ObtainAllPrivilege      bool    `json:"obtainAllPrivilege"`
	ObtainUserPrivilege     bool    `json:"obtainUserPrivilege"`
	ObtainOtherPrivilege    bool    `json:"obtainOtherPrivilege"`
	UserInteractionRequired bool    `json:"userInteractionRequired"`
}

type CVSSV2 struct {
	Version               string  `json:"version"`
	VectorString          string  `json:"vectorString"`
	AccessVector          string  `json:"accessVector"`
	AccessComplexity      string  `json:"accessComplexity"`
	Authentication        string  `json:"authentication"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	AvailabilityImpact    string  `json:"availabilityImpact"'`
	BaseScore             float64 `json:"baseScore"`
}

/*
    ===================
	==   VARIABLES   ==
	===================
*/

var (
	// Directories
	srcDir, _   = os.Getwd()
	rootDir     = filepath.Dir(srcDir)
	rootFeedDir = filepath.Join(rootDir, "CVE-Feeds")

	timeNow    = time.Now().Format(time.RFC3339)
	cveModFeed = cveFeed{
		name: "CVE-Modified",
		dir:  filepath.Join(rootFeedDir, "CVE-Modified"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-modified.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-modified.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-modified.json.zip",
	}
	cveRecFeed = cveFeed{
		name: "CVE-Recent",
		dir:  filepath.Join(rootFeedDir, "CVE-Recent"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-recent.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-recent.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-recent.json.zip",
	}
	cve2019Feed = cveFeed{
		name: "CVE-2019",
		dir:  filepath.Join(rootFeedDir, "CVE-2019"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2019.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2019.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2019.json.zip",
	}
	cve2018Feed = cveFeed{
		name: "CVE-2018",
		dir:  filepath.Join(rootFeedDir, "CVE-2018"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2018.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2018.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2018.json.zip",
	}
	cve2017Feed = cveFeed{
		name: "CVE-2017",
		dir:  filepath.Join(rootFeedDir, "CVE-2017"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2017.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2017.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2017.json.zip",
	}
	cve2016Feed = cveFeed{
		name: "CVE-2016",
		dir:  filepath.Join(rootFeedDir, "CVE-2016"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2016.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2016.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2016.json.zip",
	}
	cve2015Feed = cveFeed{
		name: "CVE-2015",
		dir:  filepath.Join(rootFeedDir, "CVE-2015"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2015.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2015.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2015.json.zip",
	}
	cve2014Feed = cveFeed{
		name: "CVE-2014",
		dir:  filepath.Join(rootFeedDir, "CVE-2014"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2014.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2014.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2014.json.zip",
	}
	cve2013Feed = cveFeed{
		name: "CVE-2013",
		dir:  filepath.Join(rootFeedDir, "CVE-2013"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2013.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2013.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2013.json.zip",
	}
	cve2012Feed = cveFeed{
		name: "CVE-2012",
		dir:  filepath.Join(rootFeedDir, "CVE-2012"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2012.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2012.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2012.json.zip",
	}
	cve2011Feed = cveFeed{
		name: "CVE-2011",
		dir:  filepath.Join(rootFeedDir, "CVE-2011"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2011.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2011.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2011.json.zip",
	}
	cve2010Feed = cveFeed{
		name: "CVE-2010",
		dir:  filepath.Join(rootFeedDir, "CVE-2010"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2010.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2010.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2010.json.zip",
	}
	cve2009Feed = cveFeed{
		name: "CVE-2009",
		dir:  filepath.Join(rootFeedDir, "CVE-2009"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2009.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2009.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2009.json.zip",
	}
	cve2008Feed = cveFeed{
		name: "CVE-2008",
		dir:  filepath.Join(rootFeedDir, "CVE-2008"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2008.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2008.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2008.json.zip",
	}
	cve2007Feed = cveFeed{
		name: "CVE-2007",
		dir:  filepath.Join(rootFeedDir, "CVE-2007"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2007.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2007.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2007.json.zip",
	}
	cve2006Feed = cveFeed{
		name: "CVE-2006",
		dir:  filepath.Join(rootFeedDir, "CVE-2006"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2006.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2006.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2006.json.zip",
	}
	cve2005Feed = cveFeed{
		name: "CVE-2005",
		dir:  filepath.Join(rootFeedDir, "CVE-2005"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2005.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2005.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2005.json.zip",
	}
	cve2004Feed = cveFeed{
		name: "CVE-2004",
		dir:  filepath.Join(rootFeedDir, "CVE-2004"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2004.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2004.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2004.json.zip",
	}
	cve2003Feed = cveFeed{
		name: "CVE-2003",
		dir:  filepath.Join(rootFeedDir, "CVE-2003"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2003.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2003.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2003.json.zip",
	}
	cve2002Feed = cveFeed{
		name: "CVE-2002",
		dir:  filepath.Join(rootFeedDir, "CVE-2002"),
		meta: "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2002.meta",
		gz:   "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2002.json.gz",
		zip:  "https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-2002.json.zip",
	}

	cveFeeds = []cveFeed{
		cveModFeed,
		cveRecFeed,
		cve2019Feed,
		cve2018Feed,
		cve2017Feed,
		cve2016Feed,
		cve2015Feed,
		cve2014Feed,
		cve2013Feed,
		cve2012Feed,
		cve2011Feed,
		cve2010Feed,
		cve2009Feed,
		cve2008Feed,
		cve2007Feed,
		cve2006Feed,
		cve2005Feed,
		cve2004Feed,
		cve2003Feed,
		cve2002Feed,
	}

	c = make(chan string)
)

func chkFeedRoot(d string) {
	if _, err := os.Stat(d); os.IsNotExist(err) {
		err := os.Mkdir(d, 0755)
		if err != nil {
			logging.Error.Println(err)
			os.Exit(1)
		}
	}
	_ = os.Chdir(d)
}

func (f cveFeed) mkFeedDir() {
	if _, err := os.Stat(f.dir); os.IsNotExist(err) {
		err := os.Mkdir(f.dir, 0755)
		if err != nil {
			logging.Error.Println(err)
			os.Exit(1)
		}
	}
}

// RECEIVER DOWNLOAD FUNCTION
//  - pass download link
//  - pass download file extension
//  - pass channel for goroutine communication
func (f cveFeed) dlFile(link string, ext string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", link, nil)
	req.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	out, err := os.Create(filepath.Join(f.dir, f.name+ext))
	if err != nil {
		logging.Error.Println(err)
		os.Exit(1)
		//c <- "Error fetching " + link
	}
	defer out.Close()
	io.Copy(out, resp.Body)
	logging.Info.Println("Fetched:", link)
	c <- "Fetched " + link
}

// RECEIVER UNZIP FUNCTION
//  - pass channel for goroutine communication
func (f cveFeed) gunzipFile(c chan string) {
	gzFile := filepath.Join(f.dir, f.name+".json.gz")
	gf, err := os.Open(gzFile)
	if err != nil {
		logging.Error.Println(err)
	}
	defer gf.Close()
	gr, err := gzip.NewReader(gf)
	if err != nil {
		logging.Error.Println(err)
	}
	defer gr.Close()
	out, err := os.Create(filepath.Join(f.dir, f.name+".json"))
	io.Copy(out, gr)
	logging.Info.Println("Gunzipped", gzFile)
}

func (f cveFeed) unzipFile(c chan string) {
	filx := filepath.Join(f.dir, f.name+".json.gz")
	fmt.Println(filx)
	zr, err := zip.OpenReader(filepath.Join(f.dir, f.name+".json.gz"))
	if err != nil {
		logging.Error.Println(err)
	}
	defer zr.Close()
	for _, zfile := range zr.File {
		out, err := os.Create(filepath.Join(f.dir, f.name+".json"))
		rc, err := zfile.Open()
		if err != nil {
			logging.Error.Println(err)
		}
		io.Copy(out, rc)
		c <- "unzipped " + f.name + ".json.gz"
		rc.Close()
	}
}

func openMongoDB(addr string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(addr)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logging.Error.Println(err)
	}
	logging.Info.Println("Connected to MongoDB.")
	return client
}

func closeMongoDB(client *mongo.Client) {
	err := client.Disconnect(context.TODO())
	if err != nil {
		logging.Error.Println(err)
	}
	logging.Info.Println("Connection to MongoDB closed.")
}

func listIndexes(client *mongo.Client, database string, collection string) {
	mc := client.Database(database).Collection(collection)
	duration := 10 * time.Second
	batchSize := int32(10)
	cur, err := mc.Indexes().List(context.Background(), &options.ListIndexesOptions{&batchSize, &duration})
	if err != nil {
		logging.Error.Println(err)
	}
	for cur.Next(context.Background()) {
		index := bson.D{}
		cur.Decode(&index)
		logging.Info.Println(fmt.Sprintf("index found %v", index))
	}
}

func createIndex(client *mongo.Client, database string, collection string) {
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	col := client.Database(database).Collection(collection)
	uniqueIndex := true
	mod := mongo.IndexModel{
		Keys: bson.M{
			"cve.cvedatameta.id": 1,
		}, Options: &options.IndexOptions{Unique: &uniqueIndex},
	}
	ind, err := col.Indexes().CreateOne(ctx, mod)
	if err != nil {
		logging.Error.Println("Index Creation ERROR:", err)
	}
	logging.Info.Println("Index:", ind)

}

func (f cveFeed) readJson() cveJSON {
	jsonFile := filepath.Join(f.dir, f.name+".json")
	logging.Info.Println("Reading", jsonFile)
	var cveJSON cveJSON
	jFile, err := os.Open(jsonFile)
	if err != nil {
		log.Fatal(err)
	}
	defer jFile.Close()
	byteValue, _ := ioutil.ReadAll(jFile)
	json.Unmarshal(byteValue, &cveJSON)
	return cveJSON
}

func insertCVEDoc(client *mongo.Client, cveFeed cveFeed, c chan string) {
	cveStruct := cveFeed.readJson()
	collection := client.Database("CVE").Collection(cveFeed.name)
	fmt.Println(len(cveStruct.CVEItems))
	for i := 0; i < len(cveStruct.CVEItems); i++ {
		insertResult, err := collection.InsertOne(context.TODO(), cveStruct.CVEItems[i])
		if err != nil {
			logging.Error.Println("Insert ERROR:", err)
			fmt.Println("ERROR", err.Error())
			continue
		}
		logging.Info.Println("Inserted a single document: ", insertResult.InsertedID)
		c <- "Inserted a single document"
	}
}

// MAIN FUNCTION
func main() {
	logging.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	chkFeedRoot(rootFeedDir)
	mongoClient := openMongoDB("mongodb://localhost:27017")
	for _, cveFeed := range cveFeeds {
		cveFeed.mkFeedDir()
		cveFeed.dlFile(cveFeed.meta, ".meta")
		cveFeed.dlFile(cveFeed.gz, ".json.gz")
		go cveFeed.dlFile(cveFeed.zip, ".json.zip")
		go cveFeed.gunzipFile(c)
		//createIndex(client, "CVE", cveFeed.name)
		go insertCVEDoc(mongoClient, cveFeed, c)
	}
	for i := 0; i < len(cveFeeds)*3; i++ {
		logging.Info.Println(<-c)
	}
	defer closeMongoDB(mongoClient)
	closeMongoDB(mongoClient)
}
