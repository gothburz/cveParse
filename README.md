#CVE-DB
This golang program pulls [nvd.mist.gov](https://nvd.nist.gov/vuln/data-feeds#JSON_FEED) data feeds, parses the json, and then imports each feed into a local MongoDB instance. Each year is saved as a separate collection and each document contains a unique CVE object.
## TODO
* Clean up main function goroutines.