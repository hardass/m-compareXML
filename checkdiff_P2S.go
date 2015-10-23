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
import "strings"

var debug = flag.Bool("debug", false, "enable debugging")
var testing = flag.Bool("testing", false, "output folder")

func main() {

	xmlfoldersA := "/usr/XML/staging/"
	xmlfoldersB := "/usr/XML/production/"
	comparelogsfolder := "/usr/XML/comparisonlogs/staging2production/"
	if *testing {
		comparelogsfolder = "/root/workspace/testing/comparisonlogs/staging2production/"
	}

	// get the latest folder of A
	xmldatesfolders := getsubfolderslist(xmlfoldersA)
	xmldatefolderA := xmlfoldersA + xmldatesfolders[len(xmldatesfolders)-1] + "/"

	// get the latest folder of B
	xmldatesfolders = getsubfolderslist(xmlfoldersB)
	xmldatefolderB := xmlfoldersB + xmldatesfolders[len(xmldatesfolders)-1] + "/"

	// get client folders list from A's date folder
	xmlclientsfolderslist := getsubfolderslist(xmldatefolderA)
	sort.Strings(xmlclientsfolderslist)

	// loop today's clients folders
	for _, clientnameA := range xmlclientsfolderslist {
		clientnameB := strings.TrimRight(clientnameA, "_Staging")
		xmlclientfolderApath := xmldatefolderA + clientnameA + "/"
		xmlclientfolderBpath := xmldatefolderB + clientnameB + "/"
		// check if client folder from today's date folder, exists in yesterday's
		if checkexist(xmlclientfolderBpath) {
			// get xml file list from today's client folder
			xmlsA := getsubfileslist(xmlclientfolderApath)

			// write client level log start
			if !checkexist(comparelogsfolder + clientnameB) {
				os.MkdirAll(comparelogsfolder+clientnameB, 0755)
			}
			clientlogfilepath := comparelogsfolder + clientnameB + "/log.txt"
			content := time.Now().Format("2006/01/02 15:04:05") + " =START="
			filewritein(clientlogfilepath, content)

			dailylogfolderpath := comparelogsfolder + clientnameB + "/" + time.Now().Format("20060102") + "/"

			// loop xml file from today's client folder
			for _, xml := range xmlsA {
				xmlApath := xmlclientfolderApath + xml
				xmlBpath := xmlclientfolderBpath + xml
				xmlfileA, _ := ioutil.ReadFile(xmlApath)

				if checkexist(xmlBpath) { //when this xml exists in yesterday's folder
					xmlfileB, _ := ioutil.ReadFile(xmlBpath)

					if !reflect.DeepEqual(xmlfileA, xmlfileB) {
						if *debug {
							fmt.Println(xmlApath + " is diff")
						}
						// write log
						filewritein(clientlogfilepath, "Difference: "+xml)

						if !checkexist(dailylogfolderpath) {
							os.MkdirAll(dailylogfolderpath, 0755)
						}
						//copy 2 files
						copyfile(xmlBpath, dailylogfolderpath+xml+".production.xml")
						copyfile(xmlApath, dailylogfolderpath+xml+".staging.xml")
					}
				} else { //when one xml exist in today, but not in yesterday
					// write log about missing in yesterday's folder
					filewritein(clientlogfilepath, "Staging Only: "+xml)
					if !checkexist(dailylogfolderpath) {
						os.MkdirAll(dailylogfolderpath, 0755)
					}
					copyfile(xmlApath, dailylogfolderpath+xml+".staging.xml")
				}
			}

			// check xml exists in yesterday's folder but not in today's, which means it's removed
			// get xml file list from yesterday's client folder
			xmlsB := getsubfileslist(xmlclientfolderBpath)
			for _, xml := range xmlsB {
				xmlApath := xmlclientfolderApath + xml
				xmlBpath := xmlclientfolderBpath + xml
				if !checkexist(xmlApath) { // when one xml doesn't exist in today's folder
					// write log missing today's folder
					filewritein(clientlogfilepath, "Production Only: "+xml)
					if !checkexist(dailylogfolderpath) {
						os.MkdirAll(dailylogfolderpath, 0755)
					}
					copyfile(xmlBpath, dailylogfolderpath+xml+".production.xml")
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
