// v0.3

package main

import "fmt"
import "reflect"
import "io"
import "io/ioutil"
import "sort"
import "os"
import "flag"
import "time"

var debug = flag.Bool("debug", false, "enable debugging")
var testing = flag.Bool("testing", false, "output folder")

func main() {

	xmlfolders := "/usr/XML/staging/"
	comparelogsfolder := "/usr/XML/comparisonlogs/staging2staging/"
	if *testing {
		comparelogsfolder = "/root/workspace/testing/comparisonlogs/staging2staging/"
	}

	xmldatesfolders := getsubfolderslist(xmlfolders)

	//get the latest folder
	xmldatefoldertoday := xmlfolders + xmldatesfolders[len(xmldatesfolders)-1] + "/"
	//get the second latest folder
	xmldatefolderyesterday := xmlfolders + xmldatesfolders[len(xmldatesfolders)-2] + "/"

	// get client folders list from today's date folder
	xmlclientsfolderslist := getsubfolderslist(xmldatefoldertoday)
	sort.Strings(xmlclientsfolderslist)

	// loop today's clients folders
	for _, clientname := range xmlclientsfolderslist {
		xmlclientfoldertodaypath := xmldatefoldertoday + clientname + "/"
		xmlclientfolderyesterdaypath := xmldatefolderyesterday + clientname + "/"
		// check if client folder from today's date folder, exists in yesterday's
		if checkexist(xmlclientfolderyesterdaypath) {
			// get xml file list from today's client folder
			xmlstoday := getsubfileslist(xmlclientfoldertodaypath)

			// write client level log start
			if !checkexist(comparelogsfolder + clientname) {
				os.MkdirAll(comparelogsfolder+clientname, 0755)
			}
			clientlogfilepath := comparelogsfolder + clientname + "/log.txt"
			content := time.Now().Format("2006/01/02 15:04:05") + " =START="
			filewritein(clientlogfilepath, content)

			// loop xml file from today's client folder
			for _, xml := range xmlstoday {
				xmltodaypath := xmlclientfoldertodaypath + xml
				xmlyesterdaypath := xmlclientfolderyesterdaypath + xml
				xmlfiletoday, _ := ioutil.ReadFile(xmltodaypath)
				if checkexist(xmlyesterdaypath) { //when this xml exists in yesterday's folder
					xmlfileyesterday, _ := ioutil.ReadFile(xmlyesterdaypath)

					if !reflect.DeepEqual(xmlfiletoday, xmlfileyesterday) {
						if *debug {
							fmt.Println(xmltodaypath + " is diff")
						}
						// write log
						filewritein(clientlogfilepath, "Difference: "+xml)
						dailylogfolderpath := comparelogsfolder + clientname + "/" + time.Now().Format("20060102") + "/"
						if !checkexist(dailylogfolderpath) {
							os.MkdirAll(dailylogfolderpath, 0755)
						}
						//copy 2 files
						copyfile(xmlyesterdaypath, dailylogfolderpath+xml+".before.xml")
						copyfile(xmltodaypath, dailylogfolderpath+xml+".now.xml")
					}
				} else { //when one xml exist in today, but not in yesterday
					// write log about missing in yesterday's folder
					filewritein(clientlogfilepath, "New: "+xml)
				}
			}

			// check xml exists in yesterday's folder but not in today's, which means it's removed
			// get xml file list from yesterday's client folder
			xmlsyesterday := getsubfileslist(xmlclientfolderyesterdaypath)
			for _, xml := range xmlsyesterday {
				xmltodaypath := xmlclientfoldertodaypath + xml
				if !checkexist(xmltodaypath) { // when one xml doesn't exist in today's folder
					// write log missing today's folder
					filewritein(clientlogfilepath, "Removed: "+xml)
				}
			}

			// write client level log end
			content = time.Now().Format("2006/01/02 15:04:05") + " ==END=="
			filewritein(clientlogfilepath, content)
		}

	}
}

func getsubfolderslist(path string) []string {
	contentlist, _ := ioutil.ReadDir(path)
	var foldernames []string
	for _, f := range contentlist {
		if f.IsDir() {
			foldernames = append(foldernames, f.Name())
		}
	}
	sort.Strings(foldernames)
	return foldernames
}

func getsubfileslist(path string) []string {
	contentlist, _ := ioutil.ReadDir(path)
	var filenames []string
	for _, f := range contentlist {
		if !f.IsDir() {
			filenames = append(filenames, f.Name())
		}
	}
	sort.Strings(filenames)
	return filenames
}

func checkexist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func currentdayfoldercreation(parentfolderpath string) string {
	// if _, err := os.Stat("/usr/XML/"); err == nil {
	// 	check(err)
	// }
	datefolderpath := parentfolderpath + time.Now().Format("20060102") + "/"
	os.MkdirAll(datefolderpath, 0755)
	return datefolderpath
}

func filewritein(filepath string, content string) {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	check(err)
	defer f.Close()

	n, err := f.WriteString(content + "\n")

	if *debug {
		fmt.Printf("wrote %d bytes to %s\n", n, filepath)
	}

	f.Sync()
}

func copyfile(src string, dst string) (w int64, err error) {
	srcfile, err := os.Open(src)
	check(err)
	defer srcfile.Close()

	dstfile, err := os.Create(dst)
	check(err)
	defer dstfile.Close()

	return io.Copy(dstfile, srcfile)

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
