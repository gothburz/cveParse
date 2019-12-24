package main

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"logging"
	"net/http"
	"os"
	"path/filepath"
	"time"
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
func (f cveFeed) dlFile(link string, ext string, c chan string) {
	resp, err := http.Get(link)
	defer resp.Body.Close()
	out, err := os.Create(filepath.Join(f.dir, f.name+ext))
	if err != nil {
		logging.Error.Println(err)
		os.Exit(1)
		c <- "Error fetching " + link
	}
	defer out.Close()
	io.Copy(out, resp.Body)
	c <- "Fetched " + link
}

// RECEIVER UNZIP FUNCTION
//  - pass channel for goroutine communication
func (f cveFeed) gunzipFile(c chan string) {
	gf, err := os.Open(filepath.Join(f.dir, f.name+".json.zip"))
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
	c <- "Gunzipped " + f.name + ".json.gz"
}

// MAIN FUNCTION
func main() {
	logging.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	chkFeedRoot(rootFeedDir)
	for _, cveFeed := range cveFeeds {
		cveFeed.mkFeedDir()
		go cveFeed.dlFile(cveFeed.meta, ".meta", c)
		go cveFeed.dlFile(cveFeed.gz, ".json.gz", c)
		go cveFeed.dlFile(cveFeed.gz, ".json.zip", c)
		//go cveFeed.gunzipFile(c)
	}
	for i := 0; i < len(cveFeeds)*3; i++ {
		logging.Info.Println(<-c)
	}
}
