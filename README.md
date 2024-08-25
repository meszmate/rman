# Riot Games RMAN Manifest Parser
Manifest Parser for downloading, updating games or viewing the files

## Explanation
https://technology.riotgames.com/news/supercharging-data-delivery-new-league-patcher

## Usage
This is only an example, but you can build a better downloader if you want. Check "Explanation" if you want to know how to update the game.
```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/meszmate/rman"
)

const DownloadPath string = "/Users/meszmate/manifestparse/" // Leave it empty if you want to install to the current directory
const BaseURL string = "https://valorant.dyn.riotcdn.net/channels/public/bundles"
const max_retries int = 7

func main(){
    	b := rman.LoadFileBytes("/Users/meszmate/Downloads/EB9EF8EA7C032A8B.manifest")
    	manifest, err := rman.ParseManifestData(b)
    	if err != nil{
        	log.Fatalf(err.Error())
    	}
    	for _, f := range manifest.Files{
        	fpath := DownloadPath + f.Name
        	err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
        	if err != nil {
            		log.Fatalf("Failed to create directories: %v", err)
        	}
        	file, err := os.Create(fpath)
        	if err != nil {
            		log.Fatalf("Failed to create file: %v", err)
        	}
        	defer file.Close()

        	for _, i := range f.Chunks{
            		chbytes := getChunkByURL(i.BundleID, i.BundleOffset, i.CompressedSize, max_retries)
            		if chbytes == nil{
                		fmt.Printf("%s Failed to get chunk %d, next...\n", fmt.Sprintf("%016X", i.BundleID), i.ChunkID)
				continue
			}
			file.Write(rman.Decompress(chbytes))
		}
	}
}

func getChunkByURL(BundleID uint64, offset uint32, size uint32, retries int) []byte{
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%016X.bundle", BaseURL, BundleID), nil)
	if err != nil{
		return nil
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+size-1))
	retry := 0
	for retry < retries+1{
		newbytes := rman.LoadHttpReqestBytes(req)
		if newbytes != nil{
			return newbytes
		}
		retry++
	}
	return nil
}```
