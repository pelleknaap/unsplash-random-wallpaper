package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/exec"
)

type configValues struct {
	Featured  bool
	Query     string

	//Only valid values are "landscape", "portrait", and "squarish"
	Orientation string
}

type unsplashJSON []struct {
	Urls struct {
		Full string `json:"full"`
	} `json:"urls"`

	// am going to use the download location for future versions
	Links struct {
		DownloadLocation string `json:"download_location"`
	}
}

func main() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(`D:\pelle\Documents\Go workspace\go\src\github.com\pelleknaap\background changer unsplash`)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		fmt.Printf("Fatal error config file: \n%s", err)
		log.Fatal(err)
		return
	}

	config := configValues{}

	// parsing the JSON config into the struct
	err = viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("An fatal has occurred while parsing the config into a struct: \n%s", err)
		log.Fatal(err)
		return
	}

	// Getting the access key for unsplash as environment variable
	accessKey, exists := os.LookupEnv("ACCESS_KEY_UNSPLASH")
	if exists != true {
		fmt.Println("The environment variable 'ACCESS_KEY_UNSPLASH' isn't found, please make sure it's set")
		return
	}

	// Making the URL were we're going to send a request to
	//url := "https://api.unsplash.com/photos/random?count=1&featured=true&query=nature&orientation=landscape"
	url := fmt.Sprintf("https://api.unsplash.com/photos/random?count=1&featured=%b&query=%s&orientation=%s", config.Featured, config.Query, config.Orientation)

	client := &http.Client{}

	// creating the request to unsplash
	req, _ := http.NewRequest("GET", url, nil)

	//  adding the required header to authorize ourselves by unsplash
	req.Header.Set("Authorization", fmt.Sprintf("Client-ID %s", accessKey))

	// performing the request to unsplash
	res, _ := client.Do(req)

	// making sure the body gets closed when the program exits
	defer res.Body.Close()

	var jsonData unsplashJSON

	// converting the Unsplash JSON to the image FULL URL
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&jsonData)
	if err != nil {
		fmt.Printf("Fatal error reading JSON: \n%s", err)
		log.Fatal(err)
		return
	}

	downloadUrl := jsonData[0].Urls.Full

	// executing the command to set the wallpaper, using the Urls.Full URL
	cmd := exec.Command("wallpaper", downloadUrl)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Fatal error executing command \n%s", err)
		log.Fatal(err)
		return
	}
}
